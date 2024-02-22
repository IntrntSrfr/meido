package meido

import (
	"strings"
	"time"

	"github.com/intrntsrfr/meido/internal/structs"
	"github.com/intrntsrfr/meido/pkg/mio/bot"
	"go.uber.org/zap"
)

func (m *Meido) listenMioEvents() {
	m.Bot.AddHandler(bot.BotEventCommandRan, m.logCommand)
	m.Bot.AddHandler(bot.BotEventCommandRan, m.logCommandRan)
	m.Bot.AddHandler(bot.BotEventCommandPanicked, m.logCommandPanicked)
	m.Bot.AddHandler(bot.BotEventPassiveRan, m.logPassiveRan)
	m.Bot.AddHandler(bot.BotEventPassivePanicked, m.logPassivePanicked)
	m.Bot.AddHandler(bot.BotEventApplicationCommandRan, m.logApplicationCommandRan)
	m.Bot.AddHandler(bot.BotEventApplicationCommandPanicked, m.logApplicationCommandPanicked)
	m.Bot.AddHandler(bot.BotEventMessageComponentRan, m.logMessageComponentRan)
	m.Bot.AddHandler(bot.BotEventMessageComponentPanicked, m.logMessageComponentPanicked)
	m.Bot.AddHandler(bot.BotEventModalSubmitRan, m.logModalSubmitRan)
	m.Bot.AddHandler(bot.BotEventModalSubmitPanicked, m.logModalSubmitPanicked)
}

func (m *Meido) logCommand(cmd *bot.CommandRan) {
	entry := &structs.CommandLogEntry{
		Command:   cmd.Command.Name,
		Args:      strings.Join(cmd.Message.Args(), " "),
		UserID:    cmd.Message.AuthorID(),
		GuildID:   cmd.Message.GuildID(),
		ChannelID: cmd.Message.ChannelID(),
		MessageID: cmd.Message.Message.ID,
		SentAt:    time.Now(),
	}
	if err := m.db.CreateCommandLogEntry(entry); err != nil {
		m.logger.Error("Command write to DB failed", zap.Error(err))
	}
}

func (m *Meido) logCommandRan(cmd *bot.CommandRan) {
	m.logger.Info("Command",
		zap.String("name", cmd.Command.Name),
		zap.String("id", cmd.Message.ID()),
		zap.String("channelID", cmd.Message.ChannelID()),
		zap.String("userID", cmd.Message.AuthorID()),
		zap.String("content", cmd.Message.RawContent()),
	)
}
func (m *Meido) logCommandPanicked(cmd *bot.CommandPanicked) {
	m.logger.Error("Command panic",
		zap.Any("command", cmd.Command),
		zap.Any("message", cmd.Message),
		zap.Any("reason", cmd.Reason),
	)
}

func (m *Meido) logPassiveRan(pas *bot.PassiveRan) {
	m.logger.Debug("Passive",
		zap.String("name", pas.Passive.Name),
		zap.String("id", pas.Message.ID()),
		zap.String("channelID", pas.Message.ChannelID()),
		zap.String("userID", pas.Message.AuthorID()),
	)
}

func (m *Meido) logPassivePanicked(pas *bot.PassivePanicked) {
	m.logger.Error("Passive panic",
		zap.Any("passive", pas.Passive),
		zap.Any("message", pas.Message),
		zap.Any("reason", pas.Reason),
	)
}

func (m *Meido) logApplicationCommandRan(cmd *bot.ApplicationCommandRan) {
	m.logger.Info("Slash",
		zap.String("name", cmd.Interaction.Name()),
		zap.String("id", cmd.Interaction.ID()),
		zap.String("channelID", cmd.Interaction.ChannelID()),
		zap.String("userID", cmd.Interaction.AuthorID()),
	)
}

func (m *Meido) logApplicationCommandPanicked(cmd *bot.ApplicationCommandPanicked) {
	m.logger.Error("Slash panic",
		zap.Any("slash", cmd.ApplicationCommand),
		zap.Any("interaction", cmd.Interaction),
		zap.Any("reason", cmd.Reason),
	)
}

func (m *Meido) logMessageComponentRan(cmd *bot.MessageComponentRan) {
	m.logger.Info("Component",
		zap.String("component", cmd.MessageComponent.Name),
	)
}

func (m *Meido) logMessageComponentPanicked(cmd *bot.MessageComponentPanicked) {
	m.logger.Error("Component panic",
		zap.Any("component", cmd.MessageComponent),
		zap.Any("interaction", cmd.Interaction),
		zap.Any("reason", cmd.Reason),
	)
}

func (m *Meido) logModalSubmitRan(cmd *bot.ModalSubmitRan) {
	m.logger.Info("Modal",
		zap.String("modal", cmd.ModalSubmit.Name),
	)
}

func (m *Meido) logModalSubmitPanicked(cmd *bot.ModalSubmitPanicked) {
	m.logger.Error("Modal panic",
		zap.Any("modal", cmd.ModalSubmit),
		zap.Any("interaction", cmd.Interaction),
		zap.Any("reason", cmd.Reason),
	)
}
