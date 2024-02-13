package bot

import (
	"github.com/intrntsrfr/meido/pkg/mio/discord"
)

type EventEmitter struct {
	eventCh chan *BotEventData
}

type BotEvent string

const (
	BotEventCommandRan                 BotEvent = "command_ran"
	BotEventCommandPanicked            BotEvent = "command_panicked"
	BotEventPassivePanicked            BotEvent = "passive_panicked"
	BotEventApplicationCommandRan      BotEvent = "application_command_ran"
	BotEventApplicationCommandPanicked BotEvent = "application_command_panicked"
	BotEventModalSubmitRan             BotEvent = "modal_submit_ran"
	BotEventModalSubmitPanicked        BotEvent = "modal_submit_panicked"
	BotEventMessageComponentRan        BotEvent = "message_component_ran"
	BotEventMessageComponentPanicked   BotEvent = "message_component_panicked"
)

type BotEventData struct {
	Type BotEvent
	Data interface{}
}

type CommandRan struct {
	Command *ModuleCommand
	Message *discord.DiscordMessage
}

type CommandPanicked struct {
	Command *ModuleCommand
	Message *discord.DiscordMessage
	Reason  any
}

type PassivePanicked struct {
	Passive *ModulePassive
	Message *discord.DiscordMessage
	Reason  any
}

type ApplicationCommandRan struct {
	ApplicationCommand *ModuleApplicationCommand
	Interaction        *discord.DiscordApplicationCommand
}

type ApplicationCommandPanicked struct {
	ApplicationCommand *ModuleApplicationCommand
	Interaction        *discord.DiscordApplicationCommand
	Reason             any
}

type ModalSubmitRan struct {
	ModalSubmit *ModuleModalSubmit
	Interaction *discord.DiscordModalSubmit
}
type ModalSubmitPanicked struct {
	ModalSubmit *ModuleModalSubmit
	Interaction *discord.DiscordModalSubmit
	Reason      any
}

type MessageComponentRan struct {
	MessageComponent *ModuleMessageComponent
	Interaction      *discord.DiscordMessageComponent
}

type MessageComponentPanicked struct {
	MessageComponent *ModuleMessageComponent
	Interaction      *discord.DiscordMessageComponent
	Reason           any
}

func NewEventEmitter() *EventEmitter {
	return &EventEmitter{
		eventCh: make(chan *BotEventData),
	}
}

func (em *EventEmitter) Emit(event BotEvent, data interface{}) {
	em.eventCh <- &BotEventData{Type: event, Data: data}
}

func (em *EventEmitter) Events() chan *BotEventData {
	return em.eventCh
}
