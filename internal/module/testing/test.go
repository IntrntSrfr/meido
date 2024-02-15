package testing

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/pkg/mio/bot"
	"github.com/intrntsrfr/meido/pkg/mio/discord"
	"go.uber.org/zap"
)

// Module represents the ping mod
type Module struct {
	*bot.ModuleBase
}

// New returns a new Module.
func New(b *bot.Bot, logger *zap.Logger) bot.Module {
	logger = logger.Named("Testing")
	return &Module{
		ModuleBase: bot.NewModule(b, "Testing", logger),
	}
}

// Hook will hook the Module into the Bot.
func (m *Module) Hook() error {
	if err := m.RegisterCommands(newTestCommand(m)); err != nil {
		return err
	}
	if err := m.RegisterApplicationCommands(newTestSlash(m)); err != nil {
		return err
	}
	return nil
}

// newTestCommand returns a new ping command.
func newTestCommand(m *Module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "test",
		Description:      "This is an incredible test command",
		Triggers:         []string{"m?test"},
		Usage:            "m?test",
		Cooldown:         2,
		CooldownScope:    bot.Channel,
		RequiredPerms:    0,
		CheckBotPerms:    false,
		RequiresUserType: bot.UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         true,
		Enabled:          true,
		Run: func(msg *discord.DiscordMessage) {
			_, _ = msg.Reply("test")
		},
	}
}

func newTestSlash(m *Module) *bot.ModuleApplicationCommand {
	return &bot.ModuleApplicationCommand{
		Mod:           m,
		Name:          "ping",
		Description:   "pong",
		Cooldown:      0,
		CooldownScope: bot.Channel,
		Permissions:   0,
		UserType:      bot.UserTypeAny,
		CheckBotPerms: false,
		AllowDMs:      true,
		Enabled:       true,
		Run: func(dac *discord.DiscordApplicationCommand) {
			startTime := time.Now()
			dac.Respond("Pong")
			resp := fmt.Sprintf("Pong!\nDelay: %s", time.Since(startTime))
			respData := &discordgo.WebhookEdit{
				Content: &resp,
			}
			dac.Sess.Real().InteractionResponseEdit(dac.Interaction, respData)
		},
	}
}

// NewMonkeyCommand returns a new monkey command.
func newMonkeyCommand(m *Module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "monkey",
		Description:      "Monkey",
		Triggers:         []string{"m?monkey", "m?monke", "m?monki", "m?monky"},
		Usage:            "m?monkey",
		Cooldown:         2,
		CooldownScope:    bot.User,
		RequiredPerms:    0,
		CheckBotPerms:    false,
		RequiresUserType: bot.UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         true,
		Enabled:          true,
		Run:              m.monkeyCommand,
	}
}

func (m *Module) monkeyCommand(msg *discord.DiscordMessage) {
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
