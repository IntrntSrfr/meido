package mio

import (
	"fmt"
)

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
	Message *DiscordMessage
}

type CommandPanicked struct {
	Command *ModuleCommand
	Message *DiscordMessage
	Reason  any
}

type PassiveRan struct {
	Passive *ModulePassive
	Message *DiscordMessage
}

type PassivePanicked struct {
	Passive *ModulePassive
	Message *DiscordMessage
	Reason  any
}

type ApplicationCommandRan struct {
	ApplicationCommand *ModuleApplicationCommand
	Interaction        *DiscordApplicationCommand
}

type ApplicationCommandPanicked struct {
	ApplicationCommand *ModuleApplicationCommand
	Interaction        *DiscordApplicationCommand
	Reason             any
}

type ModalSubmitRan struct {
	ModalSubmit *ModuleModalSubmit
	Interaction *DiscordModalSubmit
}
type ModalSubmitPanicked struct {
	ModalSubmit *ModuleModalSubmit
	Interaction *DiscordModalSubmit
	Reason      any
}

type MessageComponentRan struct {
	MessageComponent *ModuleMessageComponent
	Interaction      *DiscordMessageComponent
}

type MessageComponentPanicked struct {
	MessageComponent *ModuleMessageComponent
	Interaction      *DiscordMessageComponent
	Reason           any
}

type MessageProcessed struct{}

type InteractionProcessed struct{}
