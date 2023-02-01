package mio

import (
	"errors"
	"fmt"
	"go.uber.org/zap"
	"strings"
	"sync"
)

// Module represents a collection of commands and passives.
type Module interface {
	Name() string
	Passives() []*ModulePassive
	Commands() map[string]*ModuleCommand
	AllowedTypes() MessageType
	AllowDMs() bool
	Hook() error
	RegisterCommand(*ModuleCommand) error
	RegisterCommands([]*ModuleCommand) error
	RegisterPassive(*ModulePassive) error
	RegisterPassives([]*ModulePassive) error
}

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

func (m *ModuleBase) Passives() []*ModulePassive {
	return []*ModulePassive{}
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
		return errors.New(fmt.Sprintf("passive '%v' already exists in %v", pas.Name, m.Name()))
	}
	m.passives[pas.Name] = pas
	m.Log.Info("registered passive", zap.String("name", pas.Name))
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
		return errors.New(fmt.Sprintf("command '%v' already exists in %v", cmd.Name, m.Name()))
	}
	m.commands[cmd.Name] = cmd
	m.Log.Info("registered command", zap.String("name", cmd.Name))
	return nil
}

// ModuleCommand represents a command for a Module.
type ModuleCommand struct {
	Mod           Module
	Name          string
	Description   string
	Triggers      []string
	Usage         string
	Cooldown      int
	CooldownUser  bool
	RequiredPerms int64
	RequiresOwner bool
	CheckBotPerms bool
	AllowedTypes  MessageType
	AllowDMs      bool
	Enabled       bool
	Run           func(*DiscordMessage) `json:"-"`
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

func (b *Bot) FindModule(name string) Module {
	for _, m := range b.Modules {
		if strings.ToLower(m.Name()) == strings.ToLower(name) {
			return m
		}
	}
	return nil
}

func (b *Bot) FindCommand(name string) *ModuleCommand {
	for _, m := range b.Modules {
		for _, cmd := range m.Commands() {
			if strings.ToLower(cmd.Name) == strings.ToLower(name) {
				return cmd
			}

			for _, t := range cmd.Triggers {
				if strings.ToLower(t) == strings.ToLower(name) {
					return cmd
				}
			}
		}
	}
	return nil
}

func (b *Bot) FindPassive(name string) *ModulePassive {
	for _, m := range b.Modules {
		for _, cmd := range m.Passives() {
			if strings.ToLower(cmd.Name) == strings.ToLower(name) {
				return cmd
			}
		}
	}
	return nil
}

func FindCommand(mod Module, args []string) (*ModuleCommand, bool) {
	for _, cmd := range mod.Commands() {
		for _, trig := range cmd.Triggers {
			splitTrig := strings.Split(trig, " ")

			if len(args) < len(splitTrig) {
				continue
			}
			if strings.Join(args[:len(splitTrig)], " ") == trig {
				return cmd, true
			}
		}
	}
	return nil, false
}
