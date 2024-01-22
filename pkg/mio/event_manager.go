package mio

type EventManager struct {
	eventCh chan *BotEventData
}

type BotEvent string

const (
	BotEventCommandRan      BotEvent = "command_ran"
	BotEventCommandPanicked BotEvent = "command_panicked"
)

type BotEventData struct {
	Type BotEvent
	Data interface{}
}

type CommandRan struct {
	Command *ModuleCommand
	Message *DiscordMessage
}

type CommandPanicked struct {
	Command    *ModuleCommand
	Message    *DiscordMessage
	StackTrace string
}

func NewEventManager() *EventManager {
	return &EventManager{
		eventCh: make(chan *BotEventData),
	}
}

func (em *EventManager) Emit(event BotEvent, data interface{}) {
	em.eventCh <- &BotEventData{Type: event, Data: data}
}

func (em *EventManager) EventChannel() chan *BotEventData {
	return em.eventCh
}
