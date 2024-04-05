package mio

import (
	"reflect"
	"sync"
)

type EventBus struct {
	lock     sync.Mutex
	handlers map[string][]*eventHandler
}

type eventHandler struct {
	callback reflect.Value
	once     bool
}

func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[string][]*eventHandler),
	}
}

func (eb *EventBus) removeHandler(eventType string, handler *eventHandler) {
	eb.lock.Lock()
	defer eb.lock.Unlock()

	handlers := eb.handlers[eventType]
	for i, h := range handlers {
		if h == handler {
			eb.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			return
		}
	}
}

func (eb *EventBus) AddHandler(handler any) func() {
	eb.lock.Lock()
	defer eb.lock.Unlock()

	handlerValue := reflect.ValueOf(handler)
	handlerType := handlerValue.Type()
	if handlerType.Kind() != reflect.Func || handlerType.NumIn() != 1 || handlerType.NumOut() != 0 {
		panic("handler must be a function that takes exactly one argument and returns nothing")
	}

	eventType := handlerType.In(0).Elem().Name()
	if _, ok := eb.handlers[eventType]; !ok {
		eb.handlers[eventType] = make([]*eventHandler, 0)
	}

	eh := &eventHandler{
		callback: handlerValue,
		once:     false,
	}
	eb.handlers[eventType] = append(eb.handlers[eventType], eh)

	return func() {
		eb.removeHandler(eventType, eh)
	}
}

func (eb *EventBus) AddOnceHandler(handler any) func() {
	eb.lock.Lock()
	defer eb.lock.Unlock()

	handlerValue := reflect.ValueOf(handler)
	handlerType := handlerValue.Type()
	if handlerType.Kind() != reflect.Func || handlerType.NumIn() != 1 || handlerType.NumOut() != 0 {
		panic("handler must be a function that takes exactly one argument and returns nothing")
	}

	eventType := handlerType.In(0).Elem().Name()
	if _, ok := eb.handlers[eventType]; !ok {
		eb.handlers[eventType] = make([]*eventHandler, 0)
	}

	eh := &eventHandler{
		callback: handlerValue,
		once:     true,
	}
	eb.handlers[eventType] = append(eb.handlers[eventType], eh)

	return func() {
		eb.removeHandler(eventType, eh)
	}
}

func (eb *EventBus) Emit(event any) {
	eventType := reflect.TypeOf(event).Elem().Name()
	if handlers, ok := eb.handlers[eventType]; ok {
		for _, handler := range handlers {
			if handler.once {
				eb.removeHandler(eventType, handler)
			}
			go handler.callback.Call([]reflect.Value{reflect.ValueOf(event)})
		}
	}
}
