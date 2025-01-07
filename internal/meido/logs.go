package meido

import (
	"strings"
	"time"

	"github.com/intrntsrfr/meido/internal/structs"
	"github.com/intrntsrfr/meido/pkg/mio"
	"go.uber.org/zap"
)

func (m *Meido) addHandlers() {
	m.Bot.AddHandler(func(evt *mio.CommandRan) {
		m.logCommand(evt)
		m.logCommandRan(evt)
		m.countProcessedEvent(mio.BotEventCommandRan.String())
	})

	m.Bot.AddHandler(func(evt *mio.CommandPanicked) {
		m.logCommandPanicked(evt)
		m.countProcessedEvent(mio.BotEventCommandPanicked.String())
	})

	m.Bot.AddHandler(func(evt *mio.PassiveRan) {
		m.logPassiveRan(evt)
		m.countProcessedEvent(mio.BotEventPassiveRan.String())
	})

	m.Bot.AddHandler(func(evt *mio.PassivePanicked) {
		m.logPassivePanicked(evt)
		m.countProcessedEvent(mio.BotEventPassivePanicked.String())
	})

	m.Bot.AddHandler(func(evt *mio.ApplicationCommandRan) {
		m.logApplicationCommandRan(evt)
		m.countProcessedEvent(mio.BotEventApplicationCommandRan.String())
	})

	m.Bot.AddHandler(func(evt *mio.ApplicationCommandPanicked) {
		m.logApplicationCommandPanicked(evt)
		m.countProcessedEvent(mio.BotEventApplicationCommandPanicked.String())
	})

	m.Bot.AddHandler(func(evt *mio.MessageComponentRan) {
		m.logMessageComponentRan(evt)
		m.countProcessedEvent(mio.BotEventMessageComponentRan.String())
	})

	m.Bot.AddHandler(func(evt *mio.MessageComponentPanicked) {
		m.logMessageComponentPanicked(evt)
		m.countProcessedEvent(mio.BotEventMessageComponentPanicked.String())
	})

	m.Bot.AddHandler(func(evt *mio.ModalSubmitRan) {
		m.logModalSubmitRan(evt)
		m.countProcessedEvent(mio.BotEventModalSubmitRan.String())
	})

	m.Bot.AddHandler(func(evt *mio.ModalSubmitPanicked) {
		m.logModalSubmitPanicked(evt)
		m.countProcessedEvent(mio.BotEventModalSubmitPanicked.String())
	})

	m.Bot.AddHandler(func(evt *mio.MessageProcessed) {
		m.countProcessedEvent(mio.BotEventMessageProcessed.String())
	})

	m.Bot.AddHandler(func(evt *mio.InteractionProcessed) {
		m.countProcessedEvent(mio.BotEventInteractionProcessed.String())
	})
}

func (m *Meido) countProcessedEvent(eventType string) {
	if err := m.db.UpsertCount(eventType, time.Now()); err != nil {
		m.logger.Error("Process event upsert failed", zap.Error(err))
	}
}

func (m *Meido) logCommand(cmd *mio.CommandRan) {
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

func (m *Meido) logCommandRan(cmd *mio.CommandRan) {
	m.logger.Info("Command",
		zap.String("name", cmd.Command.Name),
		zap.String("id", cmd.Message.ID()),
		zap.String("channelID", cmd.Message.ChannelID()),
		zap.String("userID", cmd.Message.AuthorID()),
		zap.String("content", cmd.Message.RawContent()),
	)
}
func (m *Meido) logCommandPanicked(cmd *mio.CommandPanicked) {
	m.logger.Error("Command panic",
		zap.Any("command", cmd.Command),
		zap.Any("message", cmd.Message),
		zap.Any("reason", cmd.Reason),
	)
}

func (m *Meido) logPassiveRan(pas *mio.PassiveRan) {
	m.logger.Debug("Passive",
		zap.String("name", pas.Passive.Name),
		zap.String("id", pas.Message.ID()),
		zap.String("channelID", pas.Message.ChannelID()),
		zap.String("userID", pas.Message.AuthorID()),
	)
}

func (m *Meido) logPassivePanicked(pas *mio.PassivePanicked) {
	m.logger.Error("Passive panic",
		zap.Any("passive", pas.Passive),
		zap.Any("message", pas.Message),
		zap.Any("reason", pas.Reason),
	)
}

func (m *Meido) logApplicationCommandRan(cmd *mio.ApplicationCommandRan) {
	m.logger.Info("Slash",
		zap.String("name", cmd.Interaction.Name()),
		zap.String("id", cmd.Interaction.ID()),
		zap.String("channelID", cmd.Interaction.ChannelID()),
		zap.String("userID", cmd.Interaction.AuthorID()),
	)
}

func (m *Meido) logApplicationCommandPanicked(cmd *mio.ApplicationCommandPanicked) {
	m.logger.Error("Slash panic",
		zap.Any("slash", cmd.ApplicationCommand),
		zap.Any("interaction", cmd.Interaction),
		zap.Any("reason", cmd.Reason),
	)
}

func (m *Meido) logMessageComponentRan(cmd *mio.MessageComponentRan) {
	m.logger.Info("Component",
		zap.String("component", cmd.MessageComponent.Name),
		zap.String("customID", cmd.Interaction.Data.CustomID),
	)
}

func (m *Meido) logMessageComponentPanicked(cmd *mio.MessageComponentPanicked) {
	m.logger.Error("Component panic",
		zap.Any("component", cmd.MessageComponent),
		zap.Any("interaction", cmd.Interaction),
		zap.Any("reason", cmd.Reason),
	)
}

func (m *Meido) logModalSubmitRan(cmd *mio.ModalSubmitRan) {
	m.logger.Info("Modal",
		zap.String("modal", cmd.ModalSubmit.Name),
		zap.String("customID", cmd.Interaction.Data.CustomID),
	)
}

func (m *Meido) logModalSubmitPanicked(cmd *mio.ModalSubmitPanicked) {
	m.logger.Error("Modal panic",
		zap.Any("modal", cmd.ModalSubmit),
		zap.Any("interaction", cmd.Interaction),
		zap.Any("reason", cmd.Reason),
	)
}
