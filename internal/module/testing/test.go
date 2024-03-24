package testing

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/pkg/mio/bot"
	"github.com/intrntsrfr/meido/pkg/mio/discord"
	"github.com/intrntsrfr/meido/pkg/utils/builders"
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
	if err := m.RegisterApplicationCommands(newSlashCommand(m)); err != nil {
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
		Cooldown:         time.Second * 2,
		CooldownScope:    bot.CooldownScopeChannel,
		RequiredPerms:    0,
		CheckBotPerms:    false,
		RequiresUserType: bot.UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         true,
		Enabled:          true,
		Execute: func(msg *discord.DiscordMessage) {
			_, _ = msg.Reply("test")
		},
	}
}

func newSlashCommand(m *module) *bot.ModuleApplicationCommand {
	bld := bot.NewModuleApplicationCommandBuilder(m, "permissions").
		Type(discordgo.ChatApplicationCommand).
		Description("Get or edit permissions for a user or a role").
		AddSubcommandGroup(
			builders.NewSubCommandGroupBuilder("user", "Get or edit permissions for a user").
				AddSubCommand(builders.NewSubCommandBuilder("get", "Get permissions for a user").
					AddOption(&discordgo.ApplicationCommandOption{
						Name:        "user",
						Description: "The user to get",
						Type:        discordgo.ApplicationCommandOptionUser,
						Required:    true,
					}).Build(),
				).
				AddSubCommand(builders.NewSubCommandBuilder("edit", "Edit permissions for a user").
					AddOption(&discordgo.ApplicationCommandOption{
						Name:        "user",
						Description: "The user to edit",
						Type:        discordgo.ApplicationCommandOptionUser,
						Required:    true,
					}).Build(),
				).Build(),
		).
		AddSubcommandGroup(
			builders.NewSubCommandGroupBuilder("role", "Get or edit permissions for a role").
				AddSubCommand(builders.NewSubCommandBuilder("get", "Get permissions for a role").
					AddOption(&discordgo.ApplicationCommandOption{
						Name:        "role",
						Description: "The role to get",
						Type:        discordgo.ApplicationCommandOptionRole,
						Required:    true,
					}).Build(),
				).
				AddSubCommand(builders.NewSubCommandBuilder("edit", "Edit permissions for a role").
					AddOption(&discordgo.ApplicationCommandOption{
						Name:        "user",
						Description: "The role to edit",
						Type:        discordgo.ApplicationCommandOptionRole,
						Required:    true,
					}).Build(),
				).Build(),
		).
		NoDM()

	exec := func(dac *discord.DiscordApplicationCommand) {
		if _, ok := dac.Options("user"); ok {
			if _, ok := dac.Options("user:get"); ok {
				userOpt, _ := dac.Options("user:get:user")
				user := userOpt.UserValue(dac.Sess.Real())
				dac.Respond(fmt.Sprintf("get %s", user.Mention()))
			} else if _, ok := dac.Options("user:edit"); ok {
				userOpt, _ := dac.Options("user:edit:user")
				user := userOpt.UserValue(dac.Sess.Real())
				dac.Respond(fmt.Sprintf("edit %s", user.Mention()))
			}
		} else if _, ok := dac.Options("role"); ok {
			if _, ok := dac.Options("role:get"); ok {
				roleOpt, _ := dac.Options("role:get:role")
				role := roleOpt.RoleValue(dac.Sess.Real(), dac.GuildID())
				dac.Respond(fmt.Sprintf("get %s", role.Mention()))
			} else if _, ok := dac.Options("role:edit"); ok {
				roleOpt, _ := dac.Options("role:edit:role")
				role := roleOpt.RoleValue(dac.Sess.Real(), dac.GuildID())
				dac.Respond(fmt.Sprintf("edit %s", role.Mention()))
			}
		} else {
			dac.RespondEphemeral("Something went wrong.")
		}
	}

	return bld.Execute(exec).Build()
}

func newMonkeyCommand(m *module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "monkey",
		Description:      "Monkey",
		Triggers:         []string{"m?monkey", "m?monke", "m?monki", "m?monky"},
		Usage:            "m?monkey",
		Cooldown:         time.Second * 2,
		CooldownScope:    bot.CooldownScopeUser,
		RequiredPerms:    0,
		CheckBotPerms:    false,
		RequiresUserType: bot.UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         true,
		Enabled:          true,
		Execute:          m.monkeyCommand,
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
