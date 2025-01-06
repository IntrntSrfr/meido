package mio

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/pkg/mio/discord"
)

// Module is a collection of interactivity which can be plugged
// into a Bot
type Module interface {
	ModuleInfo
	MessageHandler
	InteractionHandler
	CommandHandler
	PassiveHandler
	ApplicationCommandHandler
	MessageComponentHandler
	ModalSubmitHandler

	// Hook should register callbacks and do additional required
	// module setup. It must be user-defined on a per-module basis.
	Hook() error
}

type ModuleInfo interface {
	Name() string
	AllowedTypes() discord.MessageType
	AllowDMs() bool
}

type MessageHandler interface {
	HandleMessage(*discord.DiscordMessage)
}

type InteractionHandler interface {
	HandleInteraction(*discord.DiscordInteraction)
}

type CommandHandler interface {
	Commands() map[string]*ModuleCommand
	RegisterCommands(...*ModuleCommand) error
	FindCommand(name string) (*ModuleCommand, error)
}

type PassiveHandler interface {
	Passives() map[string]*ModulePassive
	RegisterPassives(...*ModulePassive) error
	FindPassive(name string) (*ModulePassive, error)
}

type ApplicationCommandHandler interface {
	ApplicationCommands() map[string]*ModuleApplicationCommand
	RegisterApplicationCommands(...*ModuleApplicationCommand) error
	FindApplicationCommand(name string) (*ModuleApplicationCommand, error)
	ApplicationCommandStructs() []*discordgo.ApplicationCommand
}

type MessageComponentHandler interface {
	MessageComponents() map[string]*ModuleMessageComponent
	RegisterMessageComponents(...*ModuleMessageComponent) error
	FindMessageComponent(name string) (*ModuleMessageComponent, error)
	SetMessageComponentCallback(id, name string)
	RemoveMessageComponentCallback(id string)
}

type ModalSubmitHandler interface {
	ModalSubmits() map[string]*ModuleModalSubmit
	RegisterModalSubmits(...*ModuleModalSubmit) error
	FindModalSubmit(name string) (*ModuleModalSubmit, error)
	SetModalSubmitCallback(id, name string)
	RemoveModalSubmitCallback(id string)
}

var (
	ErrModuleNotFound             = errors.New("module not found")
	ErrCommandNotFound            = errors.New("command not found")
	ErrPassiveNotFound            = errors.New("passive not found")
	ErrApplicationCommandNotFound = errors.New("application command not found")
	ErrMessageComponentNotFound   = errors.New("message component not found")
	ErrModalSubmitNotFound        = errors.New("modal submit not found")
)

// ModuleBase serves as a base for other modules
type ModuleBase struct {
	sync.Mutex
	Bot          *Bot
	Logger       Logger
	name         string
	allowedTypes discord.MessageType
	allowDMs     bool

	commands            map[string]*ModuleCommand
	passives            map[string]*ModulePassive
	applicationCommands map[string]*ModuleApplicationCommand
	modalSubmits        map[string]*ModuleModalSubmit
	messageComponents   map[string]*ModuleMessageComponent

	messageComponentCallbacks map[string]*ModuleMessageComponent
	modalSubmitCallbacks      map[string]*ModuleModalSubmit

	applicationCommandStructs []*discordgo.ApplicationCommand
}

func NewModule(bot *Bot, name string, logger Logger) *ModuleBase {
	return &ModuleBase{
		Bot:                       bot,
		Logger:                    logger,
		name:                      name,
		allowedTypes:              discord.MessageTypeCreate,
		allowDMs:                  true,
		commands:                  make(map[string]*ModuleCommand),
		passives:                  make(map[string]*ModulePassive),
		applicationCommands:       make(map[string]*ModuleApplicationCommand),
		modalSubmits:              make(map[string]*ModuleModalSubmit),
		messageComponents:         make(map[string]*ModuleMessageComponent),
		messageComponentCallbacks: make(map[string]*ModuleMessageComponent),
		modalSubmitCallbacks:      make(map[string]*ModuleModalSubmit),
		applicationCommandStructs: make([]*discordgo.ApplicationCommand, 0),
	}
}

func (m *ModuleBase) Name() string {
	return m.name
}

func (m *ModuleBase) AllowedTypes() discord.MessageType {
	return m.allowedTypes
}

func (m *ModuleBase) AllowDMs() bool {
	return m.allowDMs
}

func (m *ModuleBase) HandleMessage(msg *discord.DiscordMessage) {
	if !m.allowsMessage(msg) {
		return
	}

	for _, pas := range m.Passives() {
		m.handlePassive(pas, msg)
	}

	if len(msg.Args()) <= 0 {
		return
	}

	if cmd, err := m.findCommandByTriggers(msg.RawContent()); err == nil {
		m.handleCommand(cmd, msg)
	}
}

func (m *ModuleBase) allowsMessage(msg *discord.DiscordMessage) bool {
	if msg.IsDM() && !m.allowDMs {
		return false
	}
	if msg.Type()&m.allowedTypes == 0 {
		return false
	}
	return true
}

func (m *ModuleBase) handleCommand(cmd *ModuleCommand, msg *discord.DiscordMessage) {
	if !cmd.Enabled || !cmd.allowsMessage(msg) {
		return
	}

	if cmd.RequiresUserType == UserTypeBotOwner && !m.Bot.IsOwner(msg.AuthorID()) {
		_, _ = msg.Reply("This command is owner only")
		return
	}

	if cdKey := cmd.CooldownKey(msg); cdKey != "" {
		if t, ok := m.Bot.Cooldowns.Check(cdKey); ok {
			_, _ = msg.ReplyAndDelete(fmt.Sprintf("This command is on cooldown for another %v", t), time.Second*2)
			return
		}
		m.Bot.Cooldowns.Set(cdKey, time.Duration(cmd.Cooldown))
	}
	go m.runCommand(cmd, msg)
}

func (m *ModuleBase) recoverCommand(cmd *ModuleCommand, msg *discord.DiscordMessage) {
	if r := recover(); r != nil {
		m.Bot.Emit(&CommandPanicked{cmd, msg, r})
		_, _ = msg.Reply("Something terrible happened. Please try again. If that does not work, send a DM to bot dev(s)")
	}
}

func (m *ModuleBase) runCommand(cmd *ModuleCommand, msg *discord.DiscordMessage) {
	defer m.recoverCommand(cmd, msg)
	m.Bot.Emit(&CommandRan{cmd, msg})
	cmd.Execute(msg)
}

func (m *ModuleBase) handlePassive(pas *ModulePassive, msg *discord.DiscordMessage) {
	if !pas.Enabled || !pas.allowsMessage(msg) {
		return
	}
	go m.runPassive(pas, msg)
}

func (m *ModuleBase) recoverPassive(pas *ModulePassive, msg *discord.DiscordMessage) {
	if r := recover(); r != nil {
		m.Bot.Emit(&PassivePanicked{pas, msg, r})
		m.Logger.Error("Panic", "reason", r, "message", msg)
	}
}

func (m *ModuleBase) runPassive(pas *ModulePassive, msg *discord.DiscordMessage) {
	defer m.recoverPassive(pas, msg)
	m.Bot.Emit(&PassiveRan{pas, msg})
	pas.Execute(msg)
}

func (m *ModuleBase) HandleInteraction(it *discord.DiscordInteraction) {
	if !m.allowsInteraction(it) {
		return
	}

	switch it.Interaction.Type {
	case discordgo.InteractionApplicationCommand:
		data := it.Interaction.ApplicationCommandData()
		if cmd, err := m.FindApplicationCommand(data.Name); err == nil {
			m.handleApplicationCommand(cmd, &discord.DiscordApplicationCommand{
				DiscordInteraction: it,
				Data:               data,
			})
		}
	case discordgo.InteractionModalSubmit:
		data := it.Interaction.ModalSubmitData()
		if cmd, ok := m.modalSubmitCallbacks[data.CustomID]; ok {
			m.handleModalSubmit(cmd, &discord.DiscordModalSubmit{
				DiscordInteraction: it,
				Data:               data,
			})
		}
	case discordgo.InteractionMessageComponent:
		data := it.Interaction.MessageComponentData()
		dmc := &discord.DiscordMessageComponent{
			DiscordInteraction: it,
			Data:               data,
		}

		if cmd, ok := m.messageComponentCallbacks[data.CustomID]; ok {
			m.handleMessageComponent(cmd, dmc)
		} else if cmd, err := m.FindMessageComponent(data.CustomID); err == nil {
			m.handleMessageComponent(cmd, dmc)
		} else {
			if parts := strings.Split(data.CustomID, ":"); len(parts) > 1 {
				if cmd, err := m.FindMessageComponent(parts[0]); err == nil {
					m.handleMessageComponent(cmd, dmc)
				}
			}
		}
	}
}

func (m *ModuleBase) allowsInteraction(it *discord.DiscordInteraction) bool {
	return !(it.IsDM() && !m.allowDMs)
}

func (m *ModuleBase) handleApplicationCommand(c *ModuleApplicationCommand, it *discord.DiscordApplicationCommand) {
	if !c.Enabled || !c.allowsInteraction(it) {
		return
	}
	go m.runApplicationCommand(c, it)
}

func (m *ModuleBase) recoverApplicationCommand(c *ModuleApplicationCommand, it *discord.DiscordApplicationCommand) {
	if r := recover(); r != nil {
		m.Bot.Emit(&ApplicationCommandPanicked{c, it, r})
	}
}

func (m *ModuleBase) runApplicationCommand(c *ModuleApplicationCommand, it *discord.DiscordApplicationCommand) {
	defer m.recoverApplicationCommand(c, it)
	m.Bot.Emit(&ApplicationCommandRan{c, it})
	c.Execute(it)
}

func (m *ModuleBase) handleMessageComponent(c *ModuleMessageComponent, it *discord.DiscordMessageComponent) {
	if !c.Enabled || !c.allowsInteraction(it) {
		return
	}
	go m.runMessageComponent(c, it)
}

func (m *ModuleBase) recoverMessageComponent(c *ModuleMessageComponent, it *discord.DiscordMessageComponent) {
	if r := recover(); r != nil {
		m.Bot.Emit(&MessageComponentPanicked{c, it, r})
	}
}

func (m *ModuleBase) runMessageComponent(c *ModuleMessageComponent, it *discord.DiscordMessageComponent) {
	defer m.recoverMessageComponent(c, it)
	m.Bot.Emit(&MessageComponentRan{c, it})
	c.Execute(it)
}

func (m *ModuleBase) handleModalSubmit(s *ModuleModalSubmit, it *discord.DiscordModalSubmit) {
	if !s.Enabled || !s.allowsInteraction(it) {
		return
	}
	go m.runModalSubmit(s, it)
}

func (m *ModuleBase) recoverModalSubmit(s *ModuleModalSubmit, it *discord.DiscordModalSubmit) {
	if r := recover(); r != nil {
		m.Bot.Emit(&ModalSubmitPanicked{s, it, r})
	}
}

func (m *ModuleBase) runModalSubmit(s *ModuleModalSubmit, it *discord.DiscordModalSubmit) {
	defer m.recoverModalSubmit(s, it)
	m.Bot.Emit(&ModalSubmitRan{s, it})
	s.Execute(it)
}

func (m *ModuleBase) Commands() map[string]*ModuleCommand {
	return m.commands
}

func (m *ModuleBase) RegisterCommands(commands ...*ModuleCommand) error {
	for _, cmd := range commands {
		if err := m.registerCommand(cmd); err != nil {
			return err
		}
	}
	return nil
}

func (m *ModuleBase) registerCommand(cmd *ModuleCommand) error {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.commands[cmd.Name]; ok {
		return fmt.Errorf("command '%v' already exists in %v", cmd.Name, m.Name())
	}
	m.commands[cmd.Name] = cmd
	m.Logger.Info("Registered command", "name", cmd.Name)
	return nil
}

func (m *ModuleBase) FindCommand(name string) (*ModuleCommand, error) {
	if cmd, err := m.findCommandByName(name); err == nil {
		return cmd, nil
	}
	if cmd, err := m.findCommandByTriggers(name); err == nil {
		return cmd, nil
	}
	return nil, ErrCommandNotFound
}

func (m *ModuleBase) findCommandByName(name string) (*ModuleCommand, error) {
	for _, cmd := range m.Commands() {
		if strings.EqualFold(cmd.Name, name) {
			return cmd, nil
		}
	}
	return nil, ErrCommandNotFound
}

func (m *ModuleBase) findCommandByTriggers(name string) (*ModuleCommand, error) {
	for _, cmd := range m.Commands() {
		for _, trig := range cmd.Triggers {
			splitTrig := strings.Fields(trig)
			splitName := strings.Fields(name)
			if len(splitName) < len(splitTrig) {
				continue
			}
			if strings.EqualFold(strings.Join(splitName[:len(splitTrig)], " "), trig) {
				return cmd, nil
			}
		}
	}
	return nil, ErrCommandNotFound
}

func (m *ModuleBase) Passives() map[string]*ModulePassive {
	return m.passives
}
func (m *ModuleBase) RegisterPassives(passives ...*ModulePassive) error {
	for _, pas := range passives {
		if err := m.registerPassive(pas); err != nil {
			return err
		}
	}
	return nil
}

func (m *ModuleBase) registerPassive(pas *ModulePassive) error {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.passives[pas.Name]; ok {
		return fmt.Errorf("passive '%v' already exists in %v", pas.Name, m.Name())
	}
	m.passives[pas.Name] = pas
	m.Logger.Info("Registered passive", "name", pas.Name)
	return nil
}

func (m *ModuleBase) FindPassive(name string) (*ModulePassive, error) {
	for _, cmd := range m.Passives() {
		if strings.EqualFold(cmd.Name, name) {
			return cmd, nil
		}
	}
	return nil, ErrPassiveNotFound
}

func (m *ModuleBase) ApplicationCommands() map[string]*ModuleApplicationCommand {
	return m.applicationCommands
}

func (m *ModuleBase) RegisterApplicationCommands(commands ...*ModuleApplicationCommand) error {
	for _, cmd := range commands {
		if err := m.registerApplicationCommand(cmd); err != nil {
			return err
		}
	}
	return nil
}

func (m *ModuleBase) registerApplicationCommand(command *ModuleApplicationCommand) error {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.commands[command.Name]; ok {
		return fmt.Errorf("application command '%v' already exists in %v", command.Name, m.Name())
	}
	m.applicationCommands[command.Name] = command
	m.applicationCommandStructs = append(m.applicationCommandStructs, command.ApplicationCommand)
	m.Logger.Info("Registered application command", "name", command.Name)
	return nil
}

func (m *ModuleBase) FindApplicationCommand(name string) (*ModuleApplicationCommand, error) {
	for _, s := range m.ApplicationCommands() {
		if strings.EqualFold(s.Name, name) {
			return s, nil
		}
	}
	return nil, ErrApplicationCommandNotFound
}

func (m *ModuleBase) ApplicationCommandStructs() []*discordgo.ApplicationCommand {
	return m.applicationCommandStructs
}

func (m *ModuleBase) MessageComponents() map[string]*ModuleMessageComponent {
	return m.messageComponents
}

func (m *ModuleBase) RegisterMessageComponents(components ...*ModuleMessageComponent) error {
	for _, comp := range components {
		if err := m.registerMessageComponent(comp); err != nil {
			return err
		}
	}
	return nil
}

func (m *ModuleBase) registerMessageComponent(component *ModuleMessageComponent) error {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.commands[component.Name]; ok {
		return fmt.Errorf("message component '%v' already exists in %v", component.Name, m.Name())
	}
	m.messageComponents[component.Name] = component
	m.Logger.Info("Registered message component", "name", component.Name)
	return nil
}

func (m *ModuleBase) FindMessageComponent(name string) (*ModuleMessageComponent, error) {
	for _, s := range m.MessageComponents() {
		if strings.EqualFold(s.Name, name) {
			return s, nil
		}
	}
	return nil, ErrMessageComponentNotFound
}

func (m *ModuleBase) SetMessageComponentCallback(id, name string) {
	m.Lock()
	defer m.Unlock()
	if comp, err := m.FindMessageComponent(name); err == nil {
		m.messageComponentCallbacks[id] = comp
	}
}

func (m *ModuleBase) RemoveMessageComponentCallback(id string) {
	delete(m.messageComponentCallbacks, id)
}

func (m *ModuleBase) ModalSubmits() map[string]*ModuleModalSubmit {
	return m.modalSubmits
}

func (m *ModuleBase) RegisterModalSubmits(components ...*ModuleModalSubmit) error {
	for _, comp := range components {
		if err := m.registerModalSubmit(comp); err != nil {
			return err
		}
	}
	return nil
}

func (m *ModuleBase) registerModalSubmit(component *ModuleModalSubmit) error {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.commands[component.Name]; ok {
		return fmt.Errorf("modal submit '%v' already exists in %v", component.Name, m.Name())
	}
	m.modalSubmits[component.Name] = component
	m.Logger.Info("Registered message component", "name", component.Name)
	return nil
}

func (m *ModuleBase) FindModalSubmit(name string) (*ModuleModalSubmit, error) {
	for _, s := range m.ModalSubmits() {
		if strings.EqualFold(s.Name, name) {
			return s, nil
		}
	}
	return nil, ErrModalSubmitNotFound
}

func (m *ModuleBase) SetModalSubmitCallback(id, name string) {
	m.Lock()
	defer m.Unlock()
	if comp, err := m.FindModalSubmit(name); err == nil {
		m.modalSubmitCallbacks[id] = comp
	}
}

func (m *ModuleBase) RemoveModalSubmitCallback(id string) {
	delete(m.modalSubmitCallbacks, id)
}

type CooldownScope int

const (
	CooldownScopeNone CooldownScope = 1 << iota
	CooldownScopeUser
	CooldownScopeChannel
	CooldownScopeGuild
)

type UserType int

const (
	UserTypeAny UserType = 1 << iota
	UserTypeBotOwner
)

// ModuleCommand represents a command for a Module.
type ModuleCommand struct {
	Mod              Module
	Name             string
	Description      string
	Triggers         []string
	Usage            string
	Cooldown         time.Duration
	CooldownScope    CooldownScope
	RequiredPerms    int64
	RequiresUserType UserType
	CheckBotPerms    bool
	AllowedTypes     discord.MessageType
	AllowDMs         bool
	Enabled          bool
	Execute          func(*discord.DiscordMessage) `json:"-"`
}

func (cmd *ModuleCommand) allowsMessage(msg *discord.DiscordMessage) bool {
	if msg.IsDM() && !cmd.AllowDMs {
		return false
	}

	if msg.Type()&cmd.AllowedTypes == 0 {
		return false
	}

	if cmd.RequiredPerms != 0 {
		if allow, err := msg.AuthorHasPermissions(cmd.RequiredPerms); err != nil || !allow {
			return false
		}
		if cmd.CheckBotPerms {
			if botAllow, err := msg.Discord.BotHasPermissions(msg.ChannelID(), cmd.RequiredPerms); err != nil || !botAllow {
				return false
			}
		}
	}

	return true
}

func (cmd *ModuleCommand) CooldownKey(msg *discord.DiscordMessage) string {
	switch cmd.CooldownScope {
	case CooldownScopeUser:
		return fmt.Sprintf("user:%v:%v", msg.AuthorID(), cmd.Name)
	case CooldownScopeChannel:
		return fmt.Sprintf("channel:%v:%v", msg.ChannelID(), cmd.Name)
	case CooldownScopeGuild:
		return fmt.Sprintf("guild:%v:%v", msg.GuildID(), cmd.Name)
	}
	return ""
}

// ModulePassive represents a passive for a Module.
type ModulePassive struct {
	Mod          Module
	Name         string
	Description  string
	AllowedTypes discord.MessageType
	AllowDMs     bool
	Enabled      bool
	Execute      func(*discord.DiscordMessage) `json:"-"`
}

func (pas *ModulePassive) allowsMessage(msg *discord.DiscordMessage) bool {
	if msg.IsDM() && !pas.AllowDMs {
		return false
	}

	if msg.Type()&pas.AllowedTypes == 0 {
		return false
	}
	return true
}

type ModuleApplicationCommand struct {
	*discordgo.ApplicationCommand
	Mod           Module
	Cooldown      time.Duration
	CooldownScope CooldownScope
	UserType      UserType
	CheckBotPerms bool
	Enabled       bool
	Execute       func(*discord.DiscordApplicationCommand) `json:"-"`
}

func (m *ModuleApplicationCommand) allowsInteraction(it *discord.DiscordApplicationCommand) bool {
	return true
}

type ModuleModalSubmit struct {
	Mod     Module
	Name    string
	Enabled bool
	Execute func(*discord.DiscordModalSubmit) `json:"-"`
}

func (s *ModuleModalSubmit) allowsInteraction(it *discord.DiscordModalSubmit) bool {
	return true
}

type ModuleMessageComponent struct {
	Mod           Module
	Name          string
	Cooldown      time.Duration
	CooldownScope CooldownScope
	Permissions   int64
	UserType      UserType
	CheckBotPerms bool
	Enabled       bool
	Execute       func(*discord.DiscordMessageComponent) `json:"-"`
}

func (s *ModuleMessageComponent) allowsInteraction(it *discord.DiscordMessageComponent) bool {
	return true
}
