package mio

import (
	"errors"
	"fmt"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Module represents a collection of commands and passives.
type Module interface {
	Name() string
	Passives() map[string]*ModulePassive
	Commands() map[string]*ModuleCommand
	AllowedTypes() MessageType
	AllowDMs() bool

	Hook() error
	HandleMessage(*DiscordMessage)
	HandleInteraction(*DiscordInteraction)
	AllowsMessage(*DiscordMessage) bool
	AllowsInteraction(*DiscordInteraction) bool
	RegisterCommand(*ModuleCommand) error
	RegisterCommands([]*ModuleCommand) error
	RegisterPassive(*ModulePassive) error
	RegisterPassives([]*ModulePassive) error
	FindCommand(name string) (*ModuleCommand, error)
	FindCommandByName(name string) (*ModuleCommand, error)
	FindCommandByTriggers(name string) (*ModuleCommand, error)
	FindPassive(name string) (*ModulePassive, error)
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
	commands     map[string]*ModuleCommand
	passives     map[string]*ModulePassive
	allowedTypes MessageType
	allowDMs     bool
}

func NewModule(bot *Bot, name string, logger *zap.Logger) *ModuleBase {
	return &ModuleBase{
		Bot:          bot,
		Logger:       logger,
		name:         name,
		commands:     make(map[string]*ModuleCommand),
		passives:     make(map[string]*ModulePassive),
		allowedTypes: MessageTypeCreate,
		allowDMs:     true,
	}
}

func (m *ModuleBase) HandleMessage(msg *DiscordMessage) {
	if !m.AllowsMessage(msg) {
		return
	}

	for _, pas := range m.Passives() {
		m.handlePassive(pas, msg)
	}

	if len(msg.Args()) <= 0 {
		return
	}

	if cmd, err := m.FindCommandByTriggers(msg.RawContent()); err == nil {
		m.handleCommand(cmd, msg)
	}
}

func (m *ModuleBase) handleCommand(cmd *ModuleCommand, msg *DiscordMessage) {
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

func (m *ModuleBase) recoverCommand(cmd *ModuleCommand, msg *DiscordMessage) {
	if r := recover(); r != nil {
		m.Logger.Warn("Recovery needed", zap.Any("error", r))
		m.Bot.Emit(BotEventCommandPanicked, &CommandPanicked{cmd, msg, string(debug.Stack())})
		_, _ = msg.Reply("Something terrible happened. Please try again. If that does not work, send a DM to bot dev(s)")
	}
}

func (m *ModuleBase) runCommand(cmd *ModuleCommand, msg *DiscordMessage) {
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

func (m *ModuleBase) handlePassive(pas *ModulePassive, msg *DiscordMessage) {
	if !pas.Enabled || !pas.AllowsMessage(msg) {
		return
	}
	go m.runPassive(pas, msg)
}

func (m *ModuleBase) recoverPassive(pas *ModulePassive, msg *DiscordMessage) {
	if r := recover(); r != nil {
		m.Logger.Warn("Recovery needed", zap.Any("error", r))
		m.Bot.Emit(BotEventPassivePanicked, &PassivePanicked{pas, msg, string(debug.Stack())})
	}
}

func (m *ModuleBase) runPassive(pas *ModulePassive, msg *DiscordMessage) {
	defer m.recoverPassive(pas, msg)
	pas.Run(msg)
	m.Logger.Info("Passive",
		zap.String("id", msg.ID()),
		zap.String("channelID", msg.ChannelID()),
		zap.String("userID", msg.AuthorID()),
	)
}

func (m *ModuleBase) HandleInteraction(it *DiscordInteraction) {
	panic("not implemented")
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

func (m *ModuleBase) AllowedTypes() MessageType {
	return m.allowedTypes
}

func (m *ModuleBase) AllowDMs() bool {
	return m.allowDMs
}

func (m *ModuleBase) RegisterPassives(passives []*ModulePassive) error {
	for _, pas := range passives {
		if err := m.RegisterPassive(pas); err != nil {
			return err
		}
	}
	return nil
}

func (m *ModuleBase) RegisterPassive(pas *ModulePassive) error {
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

func (m *ModuleBase) RegisterCommands(commands []*ModuleCommand) error {
	for _, cmd := range commands {
		if err := m.RegisterCommand(cmd); err != nil {
			return err
		}
	}
	return nil
}

func (m *ModuleBase) RegisterCommand(cmd *ModuleCommand) error {
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

func (m *ModuleBase) FindCommandByName(name string) (*ModuleCommand, error) {
	for _, cmd := range m.Commands() {
		if strings.EqualFold(cmd.Name, name) {
			return cmd, nil
		}
	}
	return nil, ErrCommandNotFound
}

func (m *ModuleBase) FindCommandByTriggers(name string) (*ModuleCommand, error) {
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
	if cmd, err := m.FindCommandByName(name); err == nil {
		return cmd, nil
	}
	if cmd, err := m.FindCommandByTriggers(name); err == nil {
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

func (m *ModuleBase) AllowsMessage(msg *DiscordMessage) bool {
	if msg.IsDM() && !m.allowDMs {
		return false
	}
	if msg.Type()&m.allowedTypes == 0 {
		return false
	}
	return true
}

func (m *ModuleBase) AllowsInteraction(it *DiscordInteraction) bool {
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
	AllowedTypes     MessageType
	AllowDMs         bool
	Enabled          bool
	Run              func(*DiscordMessage) `json:"-"`
}

func (cmd *ModuleCommand) AllowsMessage(msg *DiscordMessage) bool {
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

func (cmd *ModuleCommand) CooldownKey(msg *DiscordMessage) string {
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
	AllowedTypes MessageType
	AllowDMs     bool
	Enabled      bool
	Run          func(*DiscordMessage) `json:"-"`
}

func (pas *ModulePassive) AllowsMessage(msg *DiscordMessage) bool {
	if msg.IsDM() && !pas.AllowDMs {
		return false
	}

	if msg.Type()&pas.AllowedTypes == 0 {
		return false
	}
	return true
}

type ModuleSlash struct {
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
	Run           func(*DiscordInteraction) `json:"-"`
}

func (s *ModuleSlash) AllowsMessage(it *DiscordInteraction) bool {
	return !(it.IsDM() && !s.AllowDMs)
}
