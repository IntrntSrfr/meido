package testing

import (
	"github.com/intrntsrfr/meido/pkg/mio"
	"go.uber.org/zap"
	"math/rand"
)

// Module represents the ping mod
type Module struct {
	*mio.ModuleBase
}

// New returns a new Module.
func New(bot *mio.Bot, logger *zap.Logger) mio.Module {
	return &Module{
		ModuleBase: mio.NewModule(bot, "Testing", logger.Named("testing")),
	}
}

// Hook will hook the Module into the Bot.
func (m *Module) Hook() error {
	return m.RegisterCommand(NewTestCommand(m))
	//m.RegisterCommand(NewMonkeyCommand(m))
}

// NewTestCommand returns a new ping command.
func NewTestCommand(m *Module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:           m,
		Name:          "test",
		Description:   "This is an incredible test command",
		Triggers:      []string{"m?test"},
		Usage:         "m?test",
		Cooldown:      2,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run: func(msg *mio.DiscordMessage) {
			_, _ = msg.Reply("test")
		},
	}
}

// NewMonkeyCommand returns a new monkey command.
func NewMonkeyCommand(m *Module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:           m,
		Name:          "monkey",
		Description:   "Monkey",
		Triggers:      []string{"m?monkey", "m?monke", "m?monki", "m?monky"},
		Usage:         "m?monkey",
		Cooldown:      0,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run:           m.monkeyCommand,
	}
}

func (m *Module) monkeyCommand(msg *mio.DiscordMessage) {
	_, _ = msg.Reply(monkeys[rand.Intn(len(monkeys))])
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
