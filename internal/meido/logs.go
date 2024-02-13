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
			case bot.BotEventCommandPanicked:
				//m.logCommandPanicked(evt.Data.(*bot.CommandPanicked))
			case bot.BotEventPassivePanicked:
				//m.logPassivePanicked(evt.Data.(*bot.PassivePanicked))
			case bot.BotEventApplicationCommandPanicked:
				//m.logApplicationCommandPanicked(evt.Data.(*bot.ApplicationCommandPanicked))
			case bot.BotEventMessageComponentPanicked:
				//m.logMessageComponentPanicked(evt.Data.(*bot.MessageComponentPanicked))
			case bot.BotEventModalSubmitPanicked:
				//m.logModalSubmitPanicked(evt.Data.(*bot.ModalSubmitPanicked))
			}
		case <-ctx.Done():
			return
		}
	}
}

func (m *Meido) logCommand(cmd *bot.CommandRan) {
	err := m.db.CreateCommandLogEntry(&structs.CommandLogEntry{
		Command:   cmd.Command.Name,
		Args:      strings.Join(cmd.Message.Args(), " "),
		UserID:    cmd.Message.AuthorID(),
		GuildID:   cmd.Message.GuildID(),
		ChannelID: cmd.Message.ChannelID(),
		MessageID: cmd.Message.Message.ID,
		SentAt:    time.Now(),
	})
	if err != nil {
		m.logger.Error("Command write to DB failed", zap.Error(err))
	}
}

func (m *Meido) logCommandPanicked(cmd *bot.CommandPanicked) {
	m.logger.Error("Command panic",
		zap.Any("command", cmd.Command),
		zap.Any("message", cmd.Message),
		zap.Any("reason", cmd.Reason),
	)
}

func (m *Meido) logPassivePanicked(pas *bot.PassivePanicked) {
	m.logger.Error("Passive panic",
		zap.Any("passive", pas.Passive),
		zap.Any("message", pas.Message),
		zap.Any("reason", pas.Reason),
	)
}

func (m *Meido) logApplicationCommandPanicked(cmd *bot.ApplicationCommandPanicked) {
	m.logger.Error("Application command panic",
		zap.Any("application command", cmd.ApplicationCommand),
		zap.Any("interaction", cmd.Interaction),
		zap.Any("reason", cmd.Reason),
	)
}

func (m *Meido) logMessageComponentPanicked(cmd *bot.MessageComponentPanicked) {
	m.logger.Error("Message component panic",
		zap.Any("message component", cmd.MessageComponent),
		zap.Any("interaction", cmd.Interaction),
		zap.Any("reason", cmd.Reason),
	)
}

func (m *Meido) logModalSubmitPanicked(cmd *bot.ModalSubmitPanicked) {
	m.logger.Error("Modal submit panic",
		zap.Any("modal submit", cmd.ModalSubmit),
		zap.Any("interaction", cmd.Interaction),
		zap.Any("reason", cmd.Reason),
	)
}
