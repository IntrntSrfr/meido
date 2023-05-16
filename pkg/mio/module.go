package mio

import (
	"errors"
	"fmt"
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
	Log          *zap.Logger
	name         string
	commands     map[string]*ModuleCommand
	passives     map[string]*ModulePassive
	allowedTypes MessageType
	allowDMs     bool
}

type CooldownScope int

const (
	User CooldownScope = iota
	Channel
	Guild
)

type UserType int

const (
	UserTypeAny UserType = iota
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
	IsEnabled        bool
	Run              func(*DiscordMessage) `json:"-"`
}

// ModulePassive represents a passive for a Module.
type ModulePassive struct {
	Mod          Module
	Name         string
	Description  string
	AllowedTypes MessageType
	Enabled      bool
	Run          func(*DiscordMessage) `json:"-"`
}

func NewModule(bot *Bot, name string, logger *zap.Logger) *ModuleBase {
	return &ModuleBase{
		Bot:          bot,
		Log:          logger,
		name:         name,
		commands:     make(map[string]*ModuleCommand),
		passives:     make(map[string]*ModulePassive),
		allowedTypes: MessageTypeCreate,
		allowDMs:     true,
	}
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
	if m.Log != nil {
		m.Log.Info("registered passive", zap.String("name", pas.Name))
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
	if m.Log != nil {
		m.Log.Info("registered command", zap.String("name", cmd.Name))
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
