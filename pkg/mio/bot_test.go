package mio

import (
	"context"
	"testing"
	"time"

	"github.com/intrntsrfr/meido/pkg/mio/mocks"
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

func TestBot_RegisterModule(t *testing.T) {
	bot := NewBot(NewConfig(), testLogger())
	bot.RegisterModule(newTestModule(bot, "test", testLogger()))
	if len(bot.Modules) != 1 {
		t.Errorf("Bot does not have a module after registering one")
	}
}

func TestBot_Events(t *testing.T) {
	bot := NewBot(NewConfig(), testLogger())
	done := make(chan bool)
	go func() {
		select {
		case <-bot.EventChannel():
			done <- true
		case <-time.After(time.Second):
			t.Errorf("Timed out ")
		}
	}()
	bot.Emit(BotEventCommandRan, nil)
	<-done
}

func TestBot_Open(t *testing.T) {
	shards := 1
	conf := NewConfig()
	conf.Set("shards", shards)

	bot := NewBot(conf, testLogger())
	bot.Open(true)
	if len(bot.Discord.Sessions) != shards {
		t.Errorf("Bot does not have 1 session after open with 1 shard")
	}
}

func TestBot_Run(t *testing.T) {
	conf := NewConfig()
	conf.Set("shards", 1)
	bot := NewBot(conf, testLogger())
	sessionMock := mocks.NewDiscordSession("asdf")

	bot.Open(false)
	bot.Discord = testDiscord(sessionMock)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bot.Run(ctx)
	defer bot.Close()

	if !sessionMock.IsOpen {
		t.Errorf("Session should have opened")
	}
}
