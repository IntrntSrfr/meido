package testmod

import (
	"fmt"
	"github.com/intrntsrfr/meido/base"
	"math/rand"
	"sync"
)

// TestMod represents the ping mod
type TestMod struct {
	sync.Mutex
	name         string
	commands     map[string]*base.ModCommand
	allowedTypes base.MessageType
	allowDMs     bool
}

// New returns a new TestMod.
func New() base.Mod {
	return &TestMod{
		name:         "Test",
		commands:     make(map[string]*base.ModCommand),
		allowedTypes: base.MessageTypeCreate,
		allowDMs:     true,
	}
}

// Name returns the name of the mod.
func (m *TestMod) Name() string {
	return m.name
}

// Passives returns the mod passives.
func (m *TestMod) Passives() []*base.ModPassive {
	return []*base.ModPassive{}
}

// Commands returns the mod commands.
func (m *TestMod) Commands() map[string]*base.ModCommand {
	return m.commands
}

// AllowedTypes returns the allowed MessageTypes.
func (m *TestMod) AllowedTypes() base.MessageType {
	return m.allowedTypes
}

// AllowDMs returns whether the mod allows DMs.
func (m *TestMod) AllowDMs() bool {
	return m.allowDMs
}

// Hook will hook the Mod into the Bot.
func (m *TestMod) Hook() error {
	m.RegisterCommand(NewTestCommand(m))
	//m.RegisterCommand(NewMonkeyCommand(m))

	return nil
}

// RegisterCommand registers a ModCommand to the Mod
func (m *TestMod) RegisterCommand(cmd *base.ModCommand) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.commands[cmd.Name]; ok {
		panic(fmt.Sprintf("command '%v' already exists in %v", cmd.Name, m.Name()))
	}
	m.commands[cmd.Name] = cmd
}

// NewTestCommand returns a new ping command.
func NewTestCommand(m *TestMod) *base.ModCommand {
	return &base.ModCommand{
		Mod:           m,
		Name:          "test",
		Description:   "This is an incredible test command",
		Triggers:      []string{"m?test"},
		Usage:         "m?testing",
		Cooldown:      2,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  base.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run:           m.testCommand,
	}
}

func (m *TestMod) testCommand(msg *base.DiscordMessage) {
	_, _ = msg.Reply("test")
}

// NewMonkeyCommand returns a new monkey command.
func NewMonkeyCommand(m *TestMod) *base.ModCommand {
	return &base.ModCommand{
		Mod:           m,
		Name:          "monkey",
		Description:   "Monkey",
		Triggers:      []string{"m?monkey", "m?monke", "m?monki", "m?monky"},
		Usage:         "m?monkey",
		Cooldown:      0,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  base.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run:           m.monkeyCommand,
	}
}

func (m *TestMod) monkeyCommand(msg *base.DiscordMessage) {
	msg.Reply(monkeys[rand.Intn(len(monkeys))])
}

var monkeys = []string{
	"🐒",
	"🐒💨",
	"🔫🐒",
	"🎷🐒",
	"\U0001F9FB🖊️🐒",
	"🐒🚿",
	"🐒\n🚽",
	"🍌🐒",
	"🥁🐒",
	"\U0001FA98🐒",
	"🏓🐒",
	"🏸🐒",
	"🏀🐒",
	"🔨🐒",
	"⛏️🐒",
	"\U0001FAA0🐒",
	"👑\n🐒",
}
