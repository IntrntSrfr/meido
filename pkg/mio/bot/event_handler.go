package bot

import (
	"context"

	"github.com/intrntsrfr/meido/pkg/mio/discord"
	"github.com/intrntsrfr/meido/pkg/mio/util"
	"go.uber.org/zap"
)

type EventHandler struct {
	discord   *discord.Discord
	modules   *ModuleManager
	callbacks *util.CallbackManager
	logger    *zap.Logger
}

func NewEventHandler(d *discord.Discord, m *ModuleManager, c *util.CallbackManager, logger *zap.Logger) *EventHandler {
	return &EventHandler{
		discord:   d,
		modules:   m,
		callbacks: c,
		logger:    logger.Named("EventHandler"),
	}
}

func (mp *EventHandler) Listen(ctx context.Context) {
	mp.logger.Info("Started listener")
	for {
		select {
		case msg := <-mp.discord.Messages():
			go mp.DeliverCallbacks(msg)
			go mp.HandleMessage(msg)
		case it := <-mp.discord.Interactions():
			go mp.HandleInteraction(it)
		case <-ctx.Done():
			return
		}
	}
}

func (mp *EventHandler) HandleMessage(msg *discord.DiscordMessage) {
	for _, mod := range mp.modules.Modules {
		mod.HandleMessage(msg)
	}
}

func (mp *EventHandler) HandleInteraction(it *discord.DiscordInteraction) {
	for _, mod := range mp.modules.Modules {
		mod.HandleInteraction(it)
	}
}

func (mp *EventHandler) DeliverCallbacks(msg *discord.DiscordMessage) {
	if msg.Type() != discord.MessageTypeCreate {
		return
	}

	ch, err := mp.callbacks.Get(msg.CallbackKey())
	if err != nil {
		return
	}
	ch <- msg
}
