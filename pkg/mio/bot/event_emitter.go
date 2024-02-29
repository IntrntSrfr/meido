package bot

import (
	"fmt"

	"github.com/intrntsrfr/meido/pkg/mio/discord"
)

type EventEmitter struct {
	eventCh chan *BotEventData
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

type BotEvent int

const (
	BotEventCommandRan BotEvent = 1 << iota
	BotEventCommandPanicked
	BotEventPassiveRan
	BotEventPassivePanicked
	BotEventApplicationCommandRan
	BotEventApplicationCommandPanicked
	BotEventMessageComponentRan
	BotEventMessageComponentPanicked
	BotEventModalSubmitRan
	BotEventModalSubmitPanicked
	BotEventMessageProcessed
	BotEventInteractionProcessed
)

func (b BotEvent) String() string {
	switch b {
	case BotEventCommandRan:
		return "command_ran"
	case BotEventCommandPanicked:
		return "command_panicked"
	case BotEventPassiveRan:
		return "passive_ran"
	case BotEventPassivePanicked:
		return "passive_panicked"
	case BotEventApplicationCommandRan:
		return "application_command_ran"
	case BotEventApplicationCommandPanicked:
		return "application_command_panicked"
	case BotEventMessageComponentRan:
		return "message_component_ran"
	case BotEventMessageComponentPanicked:
		return "message_component_panicked"
	case BotEventModalSubmitRan:
		return "modal_submit_ran"
	case BotEventModalSubmitPanicked:
		return "modal_submit_panicked"
	case BotEventMessageProcessed:
		return "message_processed"
	case BotEventInteractionProcessed:
		return "interaction_processed"
	default:
		return fmt.Sprintf("Unknown: BotEvent(%d)", b)
	}
}

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

type PassiveRan struct {
	Passive *ModulePassive
	Message *discord.DiscordMessage
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

type MessageProcessed struct{}

type InteractionProcessed struct{}
