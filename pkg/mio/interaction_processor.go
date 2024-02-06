package mio

import (
	"context"

	"go.uber.org/zap"
)

type InteractionProcessor struct {
	Bot       *Bot
	Cooldowns *CooldownManager
	logger    *zap.Logger
}

func NewInteractionProcessor(bot *Bot, logger *zap.Logger) *InteractionProcessor {
	return &InteractionProcessor{
		Bot:       bot,
		Cooldowns: NewCooldownManager(),
		logger:    logger.Named("InteractionProcessor"),
	}
}

func (ip *InteractionProcessor) Listen(ctx context.Context) {
	ip.logger.Info("Started listener")
	for {
		select {
		case it := <-ip.Bot.Discord.interactionChan:
			go ip.ProcessInteraction(it)
		case <-ctx.Done():
			return
		}
	}
}

func (ip *InteractionProcessor) ProcessInteraction(it *DiscordInteraction) {
	for _, mod := range ip.Bot.Modules {
		if !mod.AllowsInteraction(it) {
			continue
		}

		// more stuff
	}
}
