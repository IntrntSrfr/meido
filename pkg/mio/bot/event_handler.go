package bot

import (
	"context"

	"github.com/intrntsrfr/meido/pkg/mio"
	"github.com/intrntsrfr/meido/pkg/mio/discord"
	"github.com/intrntsrfr/meido/pkg/mio/utils"
)

type EventHandler struct {
	discord   *discord.Discord
	modules   *ModuleManager
	callbacks *utils.CallbackManager
	logger    mio.Logger
	emitter   *EventEmitter
}

func NewEventHandler(d *discord.Discord, m *ModuleManager, c *utils.CallbackManager, emit *EventEmitter, logger mio.Logger) *EventHandler {
	return &EventHandler{
		discord:   d,
		modules:   m,
		callbacks: c,
		emitter:   emit,
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
			go mp.emitter.Emit(BotEventMessageProcessed, &MessageProcessed{})
		case it := <-mp.discord.Interactions():
			go mp.HandleInteraction(it)
			go mp.emitter.Emit(BotEventInteractionProcessed, &InteractionProcessed{})
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

	if ch, err := mp.callbacks.Get(msg.CallbackKey()); err == nil {
		ch <- msg
	}
}
