package meido

import (
	"context"
	"strings"
	"time"

	"github.com/intrntsrfr/meido/internal/structs"
	"github.com/intrntsrfr/meido/pkg/mio/bot"
	"go.uber.org/zap"
)

func (m *Meido) listenMioEvents(ctx context.Context) {
	for {
		select {
		case evt := <-m.Bot.Events():
			switch evt.Type {
			case bot.BotEventCommandRan:
				m.logCommand(evt.Data.(*bot.CommandRan))
				m.logCommandRan(evt.Data.(*bot.CommandRan))
				m.countProcessedEvent(bot.BotEventCommandRan.String())
			case bot.BotEventCommandPanicked:
				m.logCommandPanicked(evt.Data.(*bot.CommandPanicked))
			case bot.BotEventPassiveRan:
				m.logPassiveRan(evt.Data.(*bot.PassiveRan))
			case bot.BotEventPassivePanicked:
				m.logPassivePanicked(evt.Data.(*bot.PassivePanicked))
			case bot.BotEventApplicationCommandRan:
				m.logApplicationCommandRan(evt.Data.(*bot.ApplicationCommandRan))
				m.countProcessedEvent(bot.BotEventApplicationCommandRan.String())
			case bot.BotEventApplicationCommandPanicked:
				m.logApplicationCommandPanicked(evt.Data.(*bot.ApplicationCommandPanicked))
			case bot.BotEventMessageComponentRan:
				m.logMessageComponentRan(evt.Data.(*bot.MessageComponentRan))
				m.countProcessedEvent(bot.BotEventMessageComponentRan.String())
			case bot.BotEventMessageComponentPanicked:
				m.logMessageComponentPanicked(evt.Data.(*bot.MessageComponentPanicked))
			case bot.BotEventModalSubmitRan:
				m.logModalSubmitRan(evt.Data.(*bot.ModalSubmitRan))
				m.countProcessedEvent(bot.BotEventModalSubmitRan.String())
			case bot.BotEventModalSubmitPanicked:
				m.logModalSubmitPanicked(evt.Data.(*bot.ModalSubmitPanicked))
			case bot.BotEventMessageProcessed:
				m.countProcessedEvent(bot.BotEventMessageProcessed.String())
			case bot.BotEventInteractionProcessed:
				m.countProcessedEvent(bot.BotEventInteractionProcessed.String())
			}
		case <-ctx.Done():
			return
		}
	}
}

func (m *Meido) countProcessedEvent(eventType string) {
	if err := m.db.UpsertCount(eventType, time.Now()); err != nil {
		m.logger.Error("Process event upsert failed", zap.Error(err))
	}
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
