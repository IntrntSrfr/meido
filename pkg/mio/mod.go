package mio

import (
	"strings"
	"sync"
)

// Mod represents a collection of commands and passives.
type Mod interface {
	Name() string
	Passives() []*ModPassive
	Commands() map[string]*ModCommand
	AllowedTypes() MessageType
	AllowDMs() bool
	Hook() error
	RegisterCommand(*ModCommand)
}

type ModBase struct {
	sync.Mutex
	name         string
	commands     map[string]*ModCommand
	allowedTypes MessageType
	allowDMs     bool
}

// ModCommand represents a command for a Mod.
type ModCommand struct {
	Mod           Mod
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

// ModPassive represents a passive for a Mod.
type ModPassive struct {
	Mod          Mod
	Name         string
	Description  string
	AllowedTypes MessageType
	Enabled      bool
	Run          func(*DiscordMessage) `json:"-"`
}

func FindCommand(mod Mod, args []string) (*ModCommand, bool) {
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
