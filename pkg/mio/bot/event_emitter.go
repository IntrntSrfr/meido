package bot

import (
	"fmt"

	"github.com/intrntsrfr/meido/pkg/mio/discord"
)

type EventEmitter struct {
	handlers map[BotEvent][]any
	//eventCh  chan *BotEventData
}

func NewEventEmitter() *EventEmitter {
	return &EventEmitter{
		handlers: make(map[BotEvent][]any),
		//eventCh:  make(chan *BotEventData),
	}
}

func (em *EventEmitter) Emit(event BotEvent, data interface{}) {
	//em.eventCh <- &BotEventData{Type: event, Data: data}

	for _, h := range em.handlers[event] {
		switch event {
		case BotEventCommandRan:
			go h.(func(*CommandRan))(data.(*CommandRan))
		case BotEventCommandPanicked:
			go h.(func(*CommandPanicked))(data.(*CommandPanicked))
		case BotEventPassiveRan:
			go h.(func(*PassiveRan))(data.(*PassiveRan))
		case BotEventPassivePanicked:
			go h.(func(*PassivePanicked))(data.(*PassivePanicked))
		case BotEventApplicationCommandRan:
			go h.(func(*ApplicationCommandRan))(data.(*ApplicationCommandRan))
		case BotEventApplicationCommandPanicked:
			go h.(func(*ApplicationCommandPanicked))(data.(*ApplicationCommandPanicked))
		case BotEventMessageComponentRan:
			go h.(func(*MessageComponentRan))(data.(*MessageComponentRan))
		case BotEventMessageComponentPanicked:
			go h.(func(*MessageComponentPanicked))(data.(*MessageComponentPanicked))
		case BotEventModalSubmitRan:
			go h.(func(*ModalSubmitRan))(data.(*ModalSubmitRan))
		case BotEventModalSubmitPanicked:
			go h.(func(*ModalSubmitPanicked))(data.(*ModalSubmitPanicked))
		}
	}
}

func (em *EventEmitter) AddHandler(event BotEvent, handler any) {
	if _, ok := em.handlers[event]; !ok {
		em.handlers[event] = make([]any, 0)
	}
	em.handlers[event] = append(em.handlers[event], handler)
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
)

func (b BotEvent) String() string {
	switch b {
	case BotEventCommandRan:
		return "CommandRan"
	case BotEventCommandPanicked:
		return "CommandPanicked"
	case BotEventPassiveRan:
		return "PassiveRan"
	case BotEventPassivePanicked:
		return "PassivePanicked"
	case BotEventApplicationCommandRan:
		return "ApplicationCommandRan"
	case BotEventApplicationCommandPanicked:
		return "ApplicationCommandPanicked"
	case BotEventMessageComponentRan:
		return "MessageComponentRan"
	case BotEventMessageComponentPanicked:
		return "MessageComponentPanicked"
	case BotEventModalSubmitRan:
		return "ModalSubmitRan"
	case BotEventModalSubmitPanicked:
		return "ModalSubmitPanicked"
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
