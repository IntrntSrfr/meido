package mio

import (
	"context"

	"go.uber.org/zap"
)

type EventHandler struct {
	Bot       *Bot
	Callbacks *CallbackManager
	logger    *zap.Logger
}

func NewEventHandler(bot *Bot, logger *zap.Logger) *EventHandler {
	return &EventHandler{
		Bot:       bot,
		Callbacks: bot.Callbacks,
		logger:    logger.Named("EventHandler"),
	}
}

func (mp *EventHandler) Listen(ctx context.Context) {
	mp.logger.Info("Started listener")
	for {
		select {
		case msg := <-mp.Bot.Discord.messageChan:
			go mp.DeliverCallbacks(msg)
			go mp.HandleMessage(msg)
		case it := <-mp.Bot.Discord.interactionChan:
			go mp.HandleInteraction(it)
		case <-ctx.Done():
			return
		}
	}
}

func (mp *EventHandler) HandleMessage(msg *DiscordMessage) {
	for _, mod := range mp.Bot.Modules {
		mod.HandleMessage(msg)
	}
}

func (mp *EventHandler) HandleInteraction(it *DiscordInteraction) {
	for _, mod := range mp.Bot.Modules {
		mod.HandleInteraction(it)
	}
}

func (mp *EventHandler) DeliverCallbacks(msg *DiscordMessage) {
	if msg.Type() != MessageTypeCreate {
		return
	}

	ch, err := mp.Callbacks.Get(msg.CallbackKey())
	if err != nil {
		return
	}
	ch <- msg
}
