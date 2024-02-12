package bot

import (
	"github.com/intrntsrfr/meido/pkg/mio/discord"
)

type EventEmitter struct {
	eventCh chan *BotEventData
}

type BotEvent string

const (
	BotEventCommandRan            BotEvent = "command_ran"
	BotEventCommandPanicked       BotEvent = "command_panicked"
	BotEventPassivePanicked       BotEvent = "passive_panicked"
	BotEventApplicationCommandRan BotEvent = "application_command_ran"
	BotEventModalSubmitRan        BotEvent = "modal_submit_ran"
	BotEventMessageComponentRan   BotEvent = "message_component_ran"
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
	Command    *ModuleCommand
	Message    *discord.DiscordMessage
	StackTrace string
}

type PassivePanicked struct {
	Passive    *ModulePassive
	Message    *discord.DiscordMessage
	StackTrace string
}

type ApplicationCommandRan struct {
	ApplicationCommand *ModuleApplicationCommand
	Interaction        *discord.DiscordApplicationCommand
}

type ModalSubmitRan struct {
	ModalSubmit *ModuleModalSubmit
	Interaction *discord.DiscordModalSubmit
}

type MessageComponentRan struct {
	MessageComponent *ModuleMessageComponent
	Interaction      *discord.DiscordMessageComponent
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
