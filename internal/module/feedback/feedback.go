package feedback

import (
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/internal/database"
	"github.com/intrntsrfr/meido/pkg/mio/bot"
	"github.com/intrntsrfr/meido/pkg/mio/discord"
	"go.uber.org/zap"
)

type module struct {
	*bot.ModuleBase
	db IFeedbackDB
}

func New(b *bot.Bot, db database.DB, logger *zap.Logger) bot.Module {
	logger = logger.Named("Feedback")
	return &module{
		ModuleBase: bot.NewModule(b, "Feedback", logger),
		db:         &FeedbackDB{db},
	}
}

func (m *module) blacklistUser(userID string) error {
	return m.db.BlacklistUser(userID)
}

func (m *module) unblacklistUser(userID string) error {
	return m.db.UnblacklistUser(userID)
}

func (m *module) isUserBlacklisted(userID string) (bool, error) {
	return m.db.IsUserBlacklisted(userID)
}

func (m *module) Hook() error {
	if err := m.RegisterApplicationCommands(
		newFeedbackSlash(m),
		newBlacklistSlash(m),
		newUnblacklistSlash(m),
	); err != nil {
		return err
	}

	return nil
}

func newFeedbackSlash(m *module) *bot.ModuleApplicationCommand {
	bld := bot.NewModuleApplicationCommandBuilder(m, "feedback").
		Type(discordgo.ChatApplicationCommand).
		Description("Send feedback to the bot owner.")

	exec := func(dac *discord.DiscordApplicationCommand) {
		dac.Respond("Feedback is not yet implemented.")
	}

	return bld.Execute(exec).Build()
}

func newBlacklistSlash(m *module) *bot.ModuleApplicationCommand {
	bld := bot.NewModuleApplicationCommandBuilder(m, "blacklist").
		Type(discordgo.UserApplicationCommand).
		Description("Blacklist a user from sending feedback.").
		AddOption(&discordgo.ApplicationCommandOption{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "user",
			Description: "The user to blacklist.",
			Required:    true,
		})

	exec := func(dac *discord.DiscordApplicationCommand) {
		userID, ok := dac.Options("user")
		if !ok {
			dac.Respond("Failed to get user.")
			return
		}
		err := m.blacklistUser(userID.StringValue())
		if err != nil {
			dac.Respond("Failed to blacklist user.")
			return
		}
		dac.Respond("User has been blacklisted.")
	}

	return bld.Execute(exec).Build()
}

func newUnblacklistSlash(m *module) *bot.ModuleApplicationCommand {
	bld := bot.NewModuleApplicationCommandBuilder(m, "unblacklist").
		Type(discordgo.UserApplicationCommand).
		Description("Unblacklist a user from sending feedback.").
		AddOption(&discordgo.ApplicationCommandOption{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "user",
			Description: "The user to unblacklist.",
			Required:    true,
		})

	exec := func(dac *discord.DiscordApplicationCommand) {
		userID, ok := dac.Options("user")
		if !ok {
			dac.Respond("Failed to get user.")
			return
		}
		err := m.unblacklistUser(userID.StringValue())
		if err != nil {
			dac.Respond("Failed to unblacklist user.")
			return
		}
		dac.Respond("User has been unblacklisted.")
	}

	return bld.Execute(exec).Build()
}
