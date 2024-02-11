package bot

import (
	"github.com/intrntsrfr/meido/pkg/mio/discord"
)

type EventEmitter struct {
	eventCh chan *BotEventData
}

type BotEvent string

const (
	BotEventCommandRan      BotEvent = "command_ran"
	BotEventCommandPanicked BotEvent = "command_panicked"
	BotEventPassivePanicked BotEvent = "passive_panicked"
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
