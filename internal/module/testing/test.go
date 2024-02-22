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

type module struct {
	*bot.ModuleBase
}

func New(b *bot.Bot, logger *zap.Logger) bot.Module {
	logger = logger.Named("Testing")
	return &module{
		ModuleBase: bot.NewModule(b, "Testing", logger),
	}
}

func (m *module) Hook() error {
	if err := m.RegisterCommands(newTestCommand(m)); err != nil {
		return err
	}
	if err := m.RegisterApplicationCommands(newTestSlash(m)); err != nil {
		return err
	}
	return nil
}

func newTestCommand(m *module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "test",
		Description:      "This is an incredible test command",
		Triggers:         []string{"m?test"},
		Usage:            "m?test",
		Cooldown:         2,
		CooldownScope:    bot.CooldownScopeChannel,
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

func newTestSlash(m *module) *bot.ModuleApplicationCommand {
	bld := bot.NewModuleApplicationCommandBuilder(m, "pingo").
		Type(discordgo.ChatApplicationCommand).
		Description("pongo")

	run := func(dac *discord.DiscordApplicationCommand) {
		startTime := time.Now()
		dac.Respond("Pongo")
		resp := fmt.Sprintf("Pongo!\nDelay: %s", time.Since(startTime))
		respData := &discordgo.WebhookEdit{
			Content: &resp,
		}
		dac.Sess.Real().InteractionResponseEdit(dac.Interaction, respData)
	}

	return bld.Run(run).Build()
}

func newMonkeyCommand(m *module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "monkey",
		Description:      "Monkey",
		Triggers:         []string{"m?monkey", "m?monke", "m?monki", "m?monky"},
		Usage:            "m?monkey",
		Cooldown:         2,
		CooldownScope:    bot.CooldownScopeUser,
		RequiredPerms:    0,
		CheckBotPerms:    false,
		RequiresUserType: bot.UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         true,
		Enabled:          true,
		Run:              m.monkeyCommand,
	}
}

func (m *module) monkeyCommand(msg *discord.DiscordMessage) {
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
