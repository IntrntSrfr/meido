package mio_test

import (
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/intrntsrfr/meido/pkg/mio"
	"github.com/stretchr/testify/assert"
)

type testEvent struct {
	Value int
}

func testEventHandler(event *testEvent) {}

func TestNewEventBus(t *testing.T) {
	bus := mio.NewEventBus()
	assert.NotNil(t, bus)
	assert.IsType(t, &mio.EventBus{}, bus)
}

func TestAddHandler(t *testing.T) {
	bus := mio.NewEventBus()
	removeFunc := bus.AddHandler(testEventHandler)
	assert.NotNil(t, removeFunc)

	eventType := reflect.TypeOf((*testEvent)(nil)).Elem().Name()
	assert.Contains(t, bus.handlers, eventType)
	assert.Len(t, bus.handlers[eventType], 1)

	assert.Panics(t, func() {
		bus.AddHandler(func() {})
	}, "Adding a handler that does not take exactly one argument or returns a value should panic")

	assert.Panics(t, func() {
		bus.AddHandler(func(e *testEvent, e2 *testEvent) {})
	}, "Adding a handler that takes more than one argument should panic")

	assert.Panics(t, func() {
		bus.AddHandler(func(e *testEvent) string { return "" })
	}, "Adding a handler that returns a value should panic")
}

func TestAddOnceHandler(t *testing.T) {
	bus := mio.NewEventBus()
	removeFunc := bus.AddOnceHandler(testEventHandler)
	assert.NotNil(t, removeFunc)

	eventType := reflect.TypeOf((*testEvent)(nil)).Elem().Name()
	assert.Contains(t, bus.handlers, eventType)
	assert.Len(t, bus.handlers[eventType], 1)
	assert.True(t, bus.handlers[eventType][0].once)

	assert.Panics(t, func() {
		bus.AddOnceHandler(func() {})
	}, "Adding a handler that does not take exactly one argument or returns a value should panic")

	assert.Panics(t, func() {
		bus.AddOnceHandler(func(e *testEvent, e2 *testEvent) {})
	}, "Adding a handler that takes more than one argument should panic")

	assert.Panics(t, func() {
		bus.AddOnceHandler(func(e *testEvent) string { return "" })
	}, "Adding a handler that returns a value should panic")
}

func TestEmit(t *testing.T) {
	t.Run("Single handler", func(t *testing.T) {
		bus := mio.NewEventBus()
		called := sync.WaitGroup{}
		called.Add(1)
		val := 0

		bus.AddHandler(func(e *testEvent) {
			val = e.Value
			called.Done()
		})

		bus.Emit(&testEvent{Value: 42})
		called.Wait()
		assert.Equal(t, val, 42)
	})

	t.Run("Multiple handlers", func(t *testing.T) {
		bus := mio.NewEventBus()
		called := sync.WaitGroup{}
		called.Add(2)
		val := 0

		bus.AddHandler(func(e *testEvent) {
			val += e.Value
			called.Done()
		})

		bus.AddHandler(func(e *testEvent) {
			val += e.Value
			called.Done()
		})

		bus.Emit(&testEvent{Value: 1})
		called.Wait()
		assert.Equal(t, val, 2)
	})

	t.Run("Once handler", func(t *testing.T) {
		bus := mio.NewEventBus()
		called := sync.WaitGroup{}
		called.Add(1)
		val := 0

		bus.AddOnceHandler(func(e *testEvent) {
			val = e.Value
			called.Done()
		})

		bus.Emit(&testEvent{Value: 1})
		bus.Emit(&testEvent{Value: 2})
		called.Wait()
		assert.Equal(t, val, 1)
	})
}

func TestRemoveHandler(t *testing.T) {
	t.Run("Remove handler", func(t *testing.T) {
		bus := mio.NewEventBus()
		called := sync.WaitGroup{}
		called.Add(1)
		val := 0

		remove := bus.AddHandler(func(e *testEvent) {
			val = e.Value
			called.Done()
		})

		remove()
		bus.Emit(&testEvent{Value: 1})

		done := make(chan struct{})
		go func() {
			called.Wait()
			close(done)
		}()

		select {
		case <-done:
			t.Error("Handler was called after being removed")
		case <-time.After(time.Millisecond * 100):
			assert.Equal(t, val, 0)
		}
	})

	t.Run("Remove once handler", func(t *testing.T) {
		bus := mio.NewEventBus()
		called := sync.WaitGroup{}
		called.Add(1)
		val := 0

		remove := bus.AddOnceHandler(func(e *testEvent) {
			val = e.Value
			called.Done()
		})

		remove()
		bus.Emit(&testEvent{Value: 1})

		done := make(chan struct{})
		go func() {
			called.Wait()
			close(done)
		}()

		select {
		case <-done:
			t.Error("Handler was called after being removed")
		case <-time.After(time.Millisecond * 100):
			assert.Equal(t, val, 0)
		}
	})
}
