package testmod

import (
	"fmt"
	base2 "github.com/intrntsrfr/meido/base"
	"math/rand"
	"sync"
	"time"
)

// TestMod represents the ping mod
type TestMod struct {
	sync.Mutex
	name         string
	commands     map[string]*base2.ModCommand
	allowedTypes base2.MessageType
	allowDMs     bool
}

// New returns a new TestMod.
func New(n string) base2.Mod {
	return &TestMod{
		name:         n,
		commands:     make(map[string]*base2.ModCommand),
		allowedTypes: base2.MessageTypeCreate,
		allowDMs:     true,
	}
}

// Name returns the name of the mod.
func (m *TestMod) Name() string {
	return m.name
}

// Save saves the mod state to a file.
func (m *TestMod) Save() error {
	return nil
}

// Load loads the mod state from a file.
func (m *TestMod) Load() error {
	return nil
}

// Passives returns the mod passives.
func (m *TestMod) Passives() []*base2.ModPassive {
	return []*base2.ModPassive{}
}

// Commands returns the mod commands.
func (m *TestMod) Commands() map[string]*base2.ModCommand {
	return m.commands
}

// AllowedTypes returns the allowed MessageTypes.
func (m *TestMod) AllowedTypes() base2.MessageType {
	return m.allowedTypes
}

// AllowDMs returns whether the mod allows DMs.
func (m *TestMod) AllowDMs() bool {
	return m.allowDMs
}

// Hook will hook the Mod into the Bot.
func (m *TestMod) Hook(b *base2.Bot) error {
	err := m.Load()
	if err != nil {
		return err
	}

	rand.Seed(time.Now().Unix())

	m.RegisterCommand(NewTestCommand(m))
	//m.RegisterCommand(NewMonkeyCommand(m))

	return nil
}

// RegisterCommand registers a ModCommand to the Mod
func (m *TestMod) RegisterCommand(cmd *base2.ModCommand) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.commands[cmd.Name]; ok {
		panic(fmt.Sprintf("command '%v' already exists in %v", cmd.Name, m.Name()))
	}
	m.commands[cmd.Name] = cmd
}

// NewTestCommand returns a new ping command.
func NewTestCommand(m *TestMod) *base2.ModCommand {
	return &base2.ModCommand{
		Mod:           m,
		Name:          "test",
		Description:   "This is an incredible test command",
		Triggers:      []string{"m?test"},
		Usage:         "m?testing",
		Cooldown:      2,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  base2.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run:           m.testCommand,
	}
}

func (m *TestMod) testCommand(msg *base2.DiscordMessage) {
	_, _ = msg.Reply("test")
}

// NewMonkeyCommand returns a new monkey command.
func NewMonkeyCommand(m *TestMod) *base2.ModCommand {
	return &base2.ModCommand{
		Mod:           m,
		Name:          "monkey",
		Description:   "Monkey",
		Triggers:      []string{"m?monkey", "m?monke", "m?monki", "m?monky"},
		Usage:         "m?monkey",
		Cooldown:      0,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  base2.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run:           m.monkeyCommand,
	}
}

func (m *TestMod) monkeyCommand(msg *base2.DiscordMessage) {
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
