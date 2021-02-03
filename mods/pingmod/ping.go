package pingmod

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/intrntsrfr/meido"
)

// PingMod represents the ping mod
type PingMod struct {
	sync.Mutex
	name         string
	commands     map[string]*meido.ModCommand
	allowedTypes meido.MessageType
	allowDMs     bool
}

// New returns a new PingMod.
func New(n string) meido.Mod {
	return &PingMod{
		name:         n,
		commands:     make(map[string]*meido.ModCommand),
		allowedTypes: meido.MessageTypeCreate,
		allowDMs:     true,
	}
}

// Name returns the name of the mod.
func (m *PingMod) Name() string {
	return m.name
}

// Save saves the mod state to a file.
func (m *PingMod) Save() error {
	return nil
}

// Load loads the mod state from a file.
func (m *PingMod) Load() error {
	return nil
}

// Passives returns the mod passives.
func (m *PingMod) Passives() []*meido.ModPassive {
	return []*meido.ModPassive{}
}

// Commands returns the mod commands.
func (m *PingMod) Commands() map[string]*meido.ModCommand {
	return m.commands
}

// AllowedTypes returns the allowed MessageTypes.
func (m *PingMod) AllowedTypes() meido.MessageType {
	return m.allowedTypes
}

// AllowDMs returns whether the mod allows DMs.
func (m *PingMod) AllowDMs() bool {
	return m.allowDMs
}

// Hook will hook the Mod into the Bot.
func (m *PingMod) Hook(b *meido.Bot) error {
	err := m.Load()
	if err != nil {
		return err
	}

	rand.Seed(time.Now().Unix())

	m.RegisterCommand(NewPingCommand(m))
	m.RegisterCommand(NewMonkeyCommand(m))

	return nil
}

// RegisterCommand registers a ModCommand to the Mod
func (m *PingMod) RegisterCommand(cmd *meido.ModCommand) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.commands[cmd.Name]; ok {
		panic(fmt.Sprintf("command '%v' already exists in %v", cmd.Name, m.Name()))
	}
	m.commands[cmd.Name] = cmd
}

// NewPingCommand returns a new ping command.
func NewPingCommand(m *PingMod) *meido.ModCommand {
	return &meido.ModCommand{
		Mod:           m,
		Name:          "ping",
		Description:   "Checks ping",
		Triggers:      []string{"m?ping"},
		Usage:         "m?ping",
		Cooldown:      2,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  meido.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run:           m.pingCommand,
	}
}

func (m *PingMod) pingCommand(msg *meido.DiscordMessage) {
	if msg.LenArgs() < 1 {
		return
	}

	startTime := time.Now()

	first, err := msg.Reply("Ping")
	if err != nil {
		return
	}

	now := time.Now()
	discordLatency := now.Sub(startTime)
	botLatency := now.Sub(msg.TimeReceived)

	msg.Sess.ChannelMessageEdit(msg.Message.ChannelID, first.ID,
		fmt.Sprintf("Pong!\nDiscord delay: %s\nBot delay: %s", discordLatency, botLatency))
}

// NewMonkeyCommand returns a new monkey command.
func NewMonkeyCommand(m *PingMod) *meido.ModCommand {
	return &meido.ModCommand{
		Mod:           m,
		Name:          "monkey",
		Description:   "Monkey",
		Triggers:      []string{"m?monkey", "m?monke", "m?monki", "m?monky"},
		Usage:         "m?monkey",
		Cooldown:      0,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  meido.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run:           m.monkeyCommand,
	}
}

func (m *PingMod) monkeyCommand(msg *meido.DiscordMessage) {
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
