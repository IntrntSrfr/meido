package mio

import (
	"context"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/pkg/mio/mocks"
	"go.uber.org/zap"
)

func TestBot_IsOwner(t *testing.T) {
	conf := testConfig()
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
	bot := NewBot(testConfig(), testLogger())
	bot.RegisterModule(newTestModule(bot, "test", testLogger()))
	if len(bot.Modules) != 1 {
		t.Errorf("Bot does not have a module after registering one")
	}
}

func TestBot_Events(t *testing.T) {
	bot := NewBot(testConfig(), testLogger())
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

func TestBot_Run(t *testing.T) {
	bot := NewBot(testConfig(), testLogger())
	sessionMock := mocks.NewDiscordSession("asdf", 1)
	bot.Discord = testDiscord(nil, sessionMock)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bot.Run(ctx)
	defer bot.Close()

	if !sessionMock.IsOpen {
		t.Errorf("Session should have opened")
	}
}

func setupTestBot() (*Bot, *zap.Logger, Module) {
	bot := testBot()
	logger := testLogger()
	mod := newTestModule(bot, "testing", logger)
	return bot, logger, mod
}

func executeTestCommand(bot *Bot, mod Module, cmd *ModuleCommand, message *DiscordMessage) (chan bool, context.CancelFunc) {
	called := make(chan bool)

	mod.RegisterCommand(cmd)
	bot.RegisterModule(mod)
	ctx, cancel := context.WithCancel(context.Background())
	bot.Run(ctx)

	bot.Discord.messageChan <- message
	return called, cancel
}

func TestBot_MessageGetsHandled(t *testing.T) {
	bot, _, mod := setupTestBot()
	pas := testPassive(mod)
	cmd := testCommand(mod)

	cmdCalled := make(chan bool)
	cmd.Run = func(dm *DiscordMessage) {
		cmdCalled <- true
	}

	pasCalled := make(chan bool)
	pas.Run = func(dm *DiscordMessage) {
		pasCalled <- true
	}

	// register and run
	mod.RegisterPassive(pas)
	mod.RegisterCommand(cmd)
	bot.RegisterModule(mod)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bot.Run(ctx)

	bot.Discord.messageChan <- &DiscordMessage{
		Sess:         bot.Discord.Sess,
		Discord:      bot.Discord,
		MessageType:  MessageTypeCreate,
		TimeReceived: time.Now(),
		Message: &discordgo.Message{
			Content: ".test hello",
			GuildID: "1",
			Author: &discordgo.User{
				Username: "jeff",
			},
			ChannelID: "1",
			ID:        "1",
		},
	}

	for i := 0; i < 2; i++ {
		select {
		case <-pasCalled:
			continue
		case <-cmdCalled:
			continue
		case <-time.After(time.Second * 1):
			t.Error("Test timed out")
		}
	}
}

func TestBot_MessageWrongTypeGetsIgnored(t *testing.T) {
	bot, _, _ := setupTestBot()

	mod := newTestModule(bot, "test", testLogger())
	mod.allowedTypes = MessageTypeCreate | MessageTypeUpdate

	pas := testPassive(mod)
	cmd := testCommand(mod)
	cmdCalled := make(chan bool)
	cmd.Run = func(dm *DiscordMessage) {
		cmdCalled <- true
	}
	pasCalled := make(chan bool)
	pas.Run = func(dm *DiscordMessage) {
		pasCalled <- true
	}
	mod.RegisterPassive(pas)
	mod.RegisterCommand(cmd)

	mod2 := newTestModule(bot, "test2", testLogger())
	pas2 := testPassive(mod)
	pas2Called := make(chan bool)
	pas.Run = func(dm *DiscordMessage) {
		pas2Called <- true
	}
	mod2.RegisterPassive(pas2)

	// register and run
	bot.RegisterModule(mod)
	bot.RegisterModule(mod2)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bot.Run(ctx)

	bot.Discord.messageChan <- &DiscordMessage{
		Sess:         bot.Discord.Sess,
		Discord:      bot.Discord,
		MessageType:  MessageTypeUpdate,
		TimeReceived: time.Now(),
		Message: &discordgo.Message{
			Content: ".test hello",
			GuildID: "1",
			Author: &discordgo.User{
				Username: "jeff",
			},
			ChannelID: "1",
			ID:        "1",
		},
	}

	select {
	case <-pasCalled:
		t.Error("Passive was not expected to be called")
	case <-cmdCalled:
		t.Error("Command was not expected to be called")
	case <-pas2Called:
		t.Error("Passive2 was not expected to be called")
	case <-time.After(time.Millisecond * 10):
		break
	}
}

func TestBot_PanicCommandGetsHandled(t *testing.T) {
	var (
		bot    = testBot()
		logger = testLogger()
		mod    = newTestModule(bot, "testing", logger)
		cmd    = testCommand(mod)
	)

	// change the command
	cmd.Run = func(dm *DiscordMessage) {
		panic("i am PANICKING !!!")
	}

	// register and run
	mod.RegisterCommand(cmd)
	bot.RegisterModule(mod)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bot.Run(ctx)

	bot.Discord.Sess.State().GuildAdd(&discordgo.Guild{ID: "1", Channels: []*discordgo.Channel{}})
	bot.Discord.Sess.State().ChannelAdd(&discordgo.Channel{ID: "1", GuildID: "1"})
	bot.Discord.messageChan <- &DiscordMessage{
		Sess:         bot.Discord.Sess,
		Discord:      bot.Discord,
		MessageType:  MessageTypeCreate,
		TimeReceived: time.Now(),
		Message: &discordgo.Message{
			Content: ".test hello",
			GuildID: "1",
			Author: &discordgo.User{
				Username: "jeff",
			},
			ChannelID: "1",
			ID:        "1",
		},
	}

	select {
	case evt := <-bot.EventChannel():
		if evt.Type != BotEventCommandPanicked {
			t.Errorf("Expected panic handler to run")
		}
	case <-time.After(time.Millisecond * 10):
		t.Error("Test timed out")
	}
}

func TestBot_MessageEmptyDoesNotTriggerCommand(t *testing.T) {
	bot, _, _ := setupTestBot()

	mod := newTestModule(bot, "test", testLogger())

	cmd := testCommand(mod)
	cmdCalled := make(chan bool)
	cmd.Run = func(dm *DiscordMessage) {
		cmdCalled <- true
	}
	mod.RegisterCommand(cmd)

	// register and run
	bot.RegisterModule(mod)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bot.Run(ctx)

	bot.Discord.messageChan <- &DiscordMessage{
		Sess:         bot.Discord.Sess,
		Discord:      bot.Discord,
		MessageType:  MessageTypeCreate,
		TimeReceived: time.Now(),
		Message: &discordgo.Message{
			GuildID: "1",
			Author: &discordgo.User{
				Username: "jeff",
			},
			ChannelID: "1",
			ID:        "1",
		},
	}

	select {
	case <-cmdCalled:
		t.Error("Command was not expected to be called")
	case <-time.After(time.Millisecond * 10):
		break
	}
}

func TestBot_MessageGetsCallback(t *testing.T) {
	bot, _, _ := setupTestBot()

	mod := newTestModule(bot, "test", testLogger())

	cmd := testCommand(mod)
	cmdCalled := make(chan bool)
	cmd.Run = func(dm *DiscordMessage) {
		cb, err := mod.Bot.Callbacks.Make(dm.CallbackKey())
		if err != nil {
			t.Error(err)
		}
		select {
		case <-cb:
			cmdCalled <- true
		case <-time.After(time.Millisecond * 10):
			return
		}
	}
	mod.RegisterCommand(cmd)

	// register and run
	bot.RegisterModule(mod)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bot.Run(ctx)

	bot.Discord.messageChan <- &DiscordMessage{
		Sess:         bot.Discord.Sess,
		Discord:      bot.Discord,
		MessageType:  MessageTypeCreate,
		TimeReceived: time.Now(),
		Message: &discordgo.Message{
			Content: ".test hello",
			GuildID: "1",
			Author: &discordgo.User{
				Username: "jeff",
			},
			ChannelID: "1",
			ID:        "1",
		},
	}

	time.Sleep(time.Millisecond * 25)
	bot.Discord.messageChan <- &DiscordMessage{
		Sess:         bot.Discord.Sess,
		Discord:      bot.Discord,
		MessageType:  MessageTypeCreate,
		TimeReceived: time.Now(),
		Message: &discordgo.Message{
			GuildID: "1",
			Author: &discordgo.User{
				Username: "jeff",
			},
			ChannelID: "1",
			ID:        "2",
		},
	}

	select {
	case <-cmdCalled:
		break
	case <-time.After(time.Millisecond * 50):
		t.Error("Command callback was not called. Timed out")
	}
}
