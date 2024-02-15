package bot

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/pkg/mio/discord"
	"go.uber.org/zap"
)

type Module interface {
	ModuleInfo
	CommandHandler
	PassiveHandler
	ApplicationCommandHandler
	ModalSubmitHandler
	MessageComponentHandler
	InteractionHandler
	MessageHandler

	// Hook should register callbacks and do additional required
	// module setup. It must be user-defined on a per-module basis.
	Hook() error
}

type ModuleInfo interface {
	Name() string
	AllowedTypes() discord.MessageType
	AllowDMs() bool
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
}

type ModalSubmitHandler interface {
	ModalSubmits() map[string]*ModuleModalSubmit
	FindModalSubmit(name string) (*ModuleModalSubmit, error)
}

type MessageComponentHandler interface {
	MessageComponents() map[string]*ModuleMessageComponent
	RegisterMessageComponents(...*ModuleMessageComponent) error
	FindMessageComponent(name string) (*ModuleMessageComponent, error)
	SetMessageComponentCallback(id, name string)
	RemoveMessageComponentCallback(id string)
}

type InteractionHandler interface {
	HandleInteraction(*discord.DiscordInteraction)
	AllowsInteraction(*discord.DiscordInteraction) bool
}

type MessageHandler interface {
	HandleMessage(*discord.DiscordMessage)
	AllowsMessage(*discord.DiscordMessage) bool
}

var (
	ErrModuleNotFound  = errors.New("module not found")
	ErrCommandNotFound = errors.New("command not found")
	ErrPassiveNotFound = errors.New("passive not found")
)

// ModuleBase serves as a base for every other module, so every Module does
// not have to reimplement the methods that are all the same.
type ModuleBase struct {
	sync.Mutex
	Bot          *Bot
	Logger       *zap.Logger
	name         string
	allowedTypes discord.MessageType
	allowDMs     bool

	commands            map[string]*ModuleCommand
	passives            map[string]*ModulePassive
	applicationCommands map[string]*ModuleApplicationCommand
	modalSubmits        map[string]*ModuleModalSubmit
	messageComponents   map[string]*ModuleMessageComponent

	messageComponentCallbacks map[string]*ModuleMessageComponent
}

func NewModule(bot *Bot, name string, logger *zap.Logger) *ModuleBase {
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
	}
}

func (m *ModuleBase) HandleMessage(msg *discord.DiscordMessage) {
	if !m.AllowsMessage(msg) {
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

func (m *ModuleBase) handleCommand(cmd *ModuleCommand, msg *discord.DiscordMessage) {
	if !cmd.Enabled || !cmd.AllowsMessage(msg) {
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
		m.Bot.Emit(BotEventCommandPanicked, &CommandPanicked{cmd, msg, r})
		m.Logger.Error("Panic", zap.Any("reason", r), zap.Any("message", msg))
		_, _ = msg.Reply("Something terrible happened. Please try again. If that does not work, send a DM to bot dev(s)")
	}
}

func (m *ModuleBase) runCommand(cmd *ModuleCommand, msg *discord.DiscordMessage) {
	defer m.recoverCommand(cmd, msg)

	cmd.Run(msg)
	m.Bot.Emit(BotEventCommandRan, &CommandRan{cmd, msg})
	m.Logger.Info("Command",
		zap.String("id", msg.ID()),
		zap.String("channelID", msg.ChannelID()),
		zap.String("userID", msg.AuthorID()),
		zap.String("content", msg.RawContent()),
	)
}

func (m *ModuleBase) handlePassive(pas *ModulePassive, msg *discord.DiscordMessage) {
	if !pas.Enabled || !pas.AllowsMessage(msg) {
		return
	}
	go m.runPassive(pas, msg)
}

func (m *ModuleBase) recoverPassive(pas *ModulePassive, msg *discord.DiscordMessage) {
	if r := recover(); r != nil {
		m.Bot.Emit(BotEventPassivePanicked, &PassivePanicked{pas, msg, r})
		m.Logger.Error("Panic", zap.Any("reason", r), zap.Any("message", msg))
	}
}

func (m *ModuleBase) runPassive(pas *ModulePassive, msg *discord.DiscordMessage) {
	defer m.recoverPassive(pas, msg)
	pas.Run(msg)
	m.Logger.Info("Passive",
		zap.String("id", msg.ID()),
		zap.String("channelID", msg.ChannelID()),
		zap.String("userID", msg.AuthorID()),
	)
}

func (m *ModuleBase) HandleInteraction(it *discord.DiscordInteraction) {
	if !m.AllowsInteraction(it) {
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
		if cmd, err := m.FindModalSubmit(data.CustomID); err == nil {
			m.handleModalSubmit(cmd, &discord.DiscordModalSubmit{
				DiscordInteraction: it,
				Data:               data,
			})
		}
	case discordgo.InteractionMessageComponent:
		data := it.Interaction.MessageComponentData()
		if cmd, ok := m.messageComponentCallbacks[data.CustomID]; ok {
			m.handleMessageComponent(cmd, &discord.DiscordMessageComponent{
				DiscordInteraction: it,
				Data:               data,
			})
		}
	default:
		return
	}
}

func (m *ModuleBase) handleApplicationCommand(c *ModuleApplicationCommand, it *discord.DiscordApplicationCommand) {
	go m.runApplicationCommand(c, it)
}

func (m *ModuleBase) recoverApplicationCommand(c *ModuleApplicationCommand, it *discord.DiscordApplicationCommand) {
	if r := recover(); r != nil {
		m.Bot.Emit(BotEventApplicationCommandPanicked, &ApplicationCommandPanicked{c, it, r})
		m.Logger.Error("Panic", zap.Any("reason", r), zap.Any("interaction", it))
	}
}

func (m *ModuleBase) runApplicationCommand(c *ModuleApplicationCommand, it *discord.DiscordApplicationCommand) {
	defer m.recoverApplicationCommand(c, it)
	c.Run(it)
	m.Bot.Emit(BotEventApplicationCommandRan, &ApplicationCommandRan{c, it})
}

func (m *ModuleBase) handleMessageComponent(c *ModuleMessageComponent, it *discord.DiscordMessageComponent) {
	go m.runMessageComponent(c, it)
}

func (m *ModuleBase) recoverMessageComponent(c *ModuleMessageComponent, it *discord.DiscordMessageComponent) {
	if r := recover(); r != nil {
		m.Bot.Emit(BotEventMessageComponentPanicked, &MessageComponentPanicked{c, it, r})
		m.Logger.Error("Panic", zap.Any("reason", r), zap.Any("interaction", it))
	}
}

func (m *ModuleBase) runMessageComponent(c *ModuleMessageComponent, it *discord.DiscordMessageComponent) {
	defer m.recoverMessageComponent(c, it)
	c.Run(it)
	m.Bot.Emit(BotEventMessageComponentRan, &MessageComponentRan{c, it})
}

func (m *ModuleBase) handleModalSubmit(s *ModuleModalSubmit, it *discord.DiscordModalSubmit) {
	go m.runModalSubmit(s, it)
}

func (m *ModuleBase) recoverModalSubmit(s *ModuleModalSubmit, it *discord.DiscordModalSubmit) {
	if r := recover(); r != nil {
		m.Bot.Emit(BotEventModalSubmitPanicked, &ModalSubmitPanicked{s, it, r})
		m.Logger.Error("Panic", zap.Any("reason", r), zap.Any("interaction", it))
	}
}

func (m *ModuleBase) runModalSubmit(s *ModuleModalSubmit, it *discord.DiscordModalSubmit) {
	defer m.recoverModalSubmit(s, it)
	s.Run(it)
	m.Bot.Emit(BotEventMessageComponentRan, &ModalSubmitRan{s, it})
}

func (m *ModuleBase) Name() string {
	return m.name
}

func (m *ModuleBase) Passives() map[string]*ModulePassive {
	return m.passives
}

func (m *ModuleBase) Commands() map[string]*ModuleCommand {
	return m.commands
}

func (m *ModuleBase) ApplicationCommands() map[string]*ModuleApplicationCommand {
	return m.applicationCommands
}

func (m *ModuleBase) ModalSubmits() map[string]*ModuleModalSubmit {
	return m.modalSubmits
}

func (m *ModuleBase) MessageComponents() map[string]*ModuleMessageComponent {
	return m.messageComponents
}

func (m *ModuleBase) AllowedTypes() discord.MessageType {
	return m.allowedTypes
}

func (m *ModuleBase) AllowDMs() bool {
	return m.allowDMs
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
	if m.Logger != nil {
		m.Logger.Info("Registered passive", zap.String("name", pas.Name))
	}
	return nil
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
	if m.Logger != nil {
		m.Logger.Info("Registered command", zap.String("name", cmd.Name))
	}
	return nil
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
	if m.Logger != nil {
		m.Logger.Info("Registered application command", zap.String("name", command.Name))
	}
	return nil
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
	if m.Logger != nil {
		m.Logger.Info("Registered message component", zap.String("name", component.Name))
	}
	return nil
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

func (m *ModuleBase) FindCommand(name string) (*ModuleCommand, error) {
	if cmd, err := m.findCommandByName(name); err == nil {
		return cmd, nil
	}
	if cmd, err := m.findCommandByTriggers(name); err == nil {
		return cmd, nil
	}
	return nil, ErrCommandNotFound
}

func (m *ModuleBase) FindPassive(name string) (*ModulePassive, error) {
	for _, cmd := range m.Passives() {
		if strings.EqualFold(cmd.Name, name) {
			return cmd, nil
		}
	}
	return nil, ErrPassiveNotFound
}

func (m *ModuleBase) FindApplicationCommand(name string) (*ModuleApplicationCommand, error) {
	for _, s := range m.ApplicationCommands() {
		if strings.EqualFold(s.Name, name) {
			return s, nil
		}
	}
	return nil, ErrPassiveNotFound
}

func (m *ModuleBase) FindModalSubmit(name string) (*ModuleModalSubmit, error) {
	for _, s := range m.ModalSubmits() {
		if strings.EqualFold(s.ID, name) {
			return s, nil
		}
	}
	return nil, ErrPassiveNotFound
}

func (m *ModuleBase) FindMessageComponent(name string) (*ModuleMessageComponent, error) {
	for _, s := range m.MessageComponents() {
		if strings.EqualFold(s.Name, name) {
			return s, nil
		}
	}
	return nil, ErrPassiveNotFound
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

func (m *ModuleBase) AllowsMessage(msg *discord.DiscordMessage) bool {
	if msg.IsDM() && !m.allowDMs {
		return false
	}
	if msg.Type()&m.allowedTypes == 0 {
		return false
	}
	return true
}

func (m *ModuleBase) AllowsInteraction(it *discord.DiscordInteraction) bool {
	return !(it.IsDM() && !m.allowDMs)
}

type CooldownScope int

const (
	None CooldownScope = 1 << iota
	User
	Channel
	Guild
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
	Run              func(*discord.DiscordMessage) `json:"-"`
}

func (cmd *ModuleCommand) AllowsMessage(msg *discord.DiscordMessage) bool {
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
	case User:
		return fmt.Sprintf("user:%v:%v", msg.AuthorID(), cmd.Name)
	case Channel:
		return fmt.Sprintf("channel:%v:%v", msg.ChannelID(), cmd.Name)
	case Guild:
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
	Run          func(*discord.DiscordMessage) `json:"-"`
}

func (pas *ModulePassive) AllowsMessage(msg *discord.DiscordMessage) bool {
	if msg.IsDM() && !pas.AllowDMs {
		return false
	}

	if msg.Type()&pas.AllowedTypes == 0 {
		return false
	}
	return true
}

type ModuleApplicationCommand struct {
	Mod           Module
	Name          string
	Description   string
	Cooldown      time.Duration
	CooldownScope CooldownScope
	Permissions   int64
	UserType      UserType
	CheckBotPerms bool
	AllowDMs      bool
	Enabled       bool
	Run           func(*discord.DiscordApplicationCommand) `json:"-"`
}

func (s *ModuleApplicationCommand) AllowsMessage(it *discord.DiscordInteraction) bool {
	return !(it.IsDM() && !s.AllowDMs)
}

type ModuleModalSubmit struct {
	Mod     Module
	ID      string
	Enabled bool
	Run     func(*discord.DiscordModalSubmit) `json:"-"`
}

func (s *ModuleModalSubmit) AllowsMessage(it *discord.DiscordModalSubmit) bool {
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
	AllowDMs      bool
	Enabled       bool
	Run           func(*discord.DiscordMessageComponent) `json:"-"`
}

func (s *ModuleMessageComponent) AllowsMessage(it *discord.DiscordMessageComponent) bool {
	return !(it.IsDM() && !s.AllowDMs)
}