package mio

import (
	"context"
)

type EventHandler struct {
	discord   *Discord
	modules   *ModuleManager
	callbacks *CallbackManager
	logger    Logger
	emitter   *EventBus
}

func NewEventHandler(d *Discord, m *ModuleManager, c *CallbackManager, bus *EventBus, logger Logger) *EventHandler {
	return &EventHandler{
		discord:   d,
		modules:   m,
		callbacks: c,
		emitter:   bus,
		logger:    logger.Named("EventHandler"),
	}
}

func (mp *EventHandler) Listen(ctx context.Context) {
	mp.logger.Info("Started listener")
	for {
		select {
		case msg, ok := <-mp.discord.Messages():
			if !ok {
				continue
			}
			go mp.DeliverCallbacks(msg)
			go mp.HandleMessage(msg)
			go mp.emitter.Emit(&MessageProcessed{})
		case it, ok := <-mp.discord.Interactions():
			if !ok {
				continue
			}
			go mp.HandleInteraction(it)
			go mp.emitter.Emit(&InteractionProcessed{})
		case <-ctx.Done():
			return
		}
	}
}

func (mp *EventHandler) HandleMessage(msg *DiscordMessage) {
	for _, mod := range mp.modules.Modules {
		mod.HandleMessage(msg)
	}
}

func (mp *EventHandler) HandleInteraction(it *DiscordInteraction) {
	for _, mod := range mp.modules.Modules {
		mod.HandleInteraction(it)
	}
}

func (mp *EventHandler) DeliverCallbacks(msg *DiscordMessage) {
	if msg.Type() != MessageTypeCreate {
		return
	}

	if ch, err := mp.callbacks.Get(msg.CallbackKey()); err == nil {
		ch <- msg
	}
}
