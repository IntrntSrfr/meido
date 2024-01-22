package mio

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestBot_IsOwner(t *testing.T) {
	conf := NewConfig()
	conf.Set("owner_ids", []string{"123"})

	b := NewBot(conf, testLogger())
	if ok := b.IsOwner("123"); !ok {
		t.Errorf("Bot.IsOwner('123') = %v, want %v", ok, true)
	}

	if ok := b.IsOwner("456"); ok {
		t.Errorf("Bot.IsOwner('456') = %v, want %v", ok, false)
	}
}

func newTestModule(bot *Bot, name string, log *zap.Logger) Module {
	return &testModule{ModuleBase: *NewModule(bot, name, log)}
}

type testModule struct {
	ModuleBase
}

func (m *testModule) Hook() error {
	return nil
}

func TestBot_RegisterModule(t *testing.T) {
	bot := NewBot(NewConfig(), testLogger())
	bot.RegisterModule(newTestModule(bot, "test", testLogger()))
	if len(bot.Modules) != 1 {
		t.Errorf("Bot does not have a module after registering one")
	}
}

func TestBot_Events(t *testing.T) {
	bot := NewBot(NewConfig(), testLogger())
	called := false
	done := make(chan bool)

	bot.AddEventHandler(BotEventCommandRan, func(i interface{}) {
		called = true
		done <- true
	})
	bot.emit(BotEventCommandRan, nil)
	select {
	case <-done:
		if !called {
			t.Errorf("Callback did not switch bool")
		}
	case <-time.After(time.Second):
		t.Errorf("Timed out ")
	}
}
