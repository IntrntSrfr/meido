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
func setupTestBot() (*Bot, *zap.Logger, Module, *ModuleCommand) {
	bot := testBot()
	logger := testLogger()
	mod := newTestModule(bot, "testing", logger)
	cmd := testCommand(mod)
	return bot, logger, mod, cmd
}

func executeTestCommand(bot *Bot, mod Module, cmd *ModuleCommand, runFunc func(chan bool, *DiscordMessage), message *DiscordMessage) (chan bool, context.CancelFunc) {
	called := make(chan bool)
	cmd.Run = func(dm *DiscordMessage) {
		runFunc(called, dm)
	}

	mod.RegisterCommand(cmd)
	bot.RegisterModule(mod)
	ctx, cancel := context.WithCancel(context.Background())
	bot.Run(ctx)

	bot.Discord.messageChan <- message
	return called, cancel
}

func TestBot_SimpleCommandGetsHandled(t *testing.T) {
	bot, _, mod, cmd := setupTestBot()
	runFunc := func(called chan bool, dm *DiscordMessage) {
		called <- true
	}

	msg := &DiscordMessage{
		Sess:         bot.Discord.Sess,
		Discord:      bot.Discord,
		MessageType:  MessageTypeCreate,
		TimeReceived: time.Now(),
		Message: &discordgo.Message{
			Content: ".test hello",
			GuildID: "111",
			Author: &discordgo.User{
				Username: "jeff",
			},
			ChannelID: "1234",
			ID:        "112233",
		},
	}

	called, cancel := executeTestCommand(bot, mod, cmd, runFunc, msg)
	defer cancel()

	select {
	case <-called:
		break
	case <-time.After(time.Millisecond * 10):
		t.Error("Test timed out")
	}

	select {
	case evt := <-bot.EventChannel():
		if evt.Type != BotEventCommandRan {
			t.Errorf("Expected ran handler to run")
		}
	case <-time.After(time.Millisecond * 10):
		t.Error("Test timed out")
	}
}

func TestBot_SimplePassiveGetsHandled(t *testing.T) {
	var (
		bot    = testBot()
		logger = testLogger()
		mod    = newTestModule(bot, "testing", logger)
		pas    = testPassive(mod)
	)

	// change the command
	called := make(chan bool)
	pas.Run = func(dm *DiscordMessage) {
		called <- true
	}

	// register and run
	mod.RegisterPassive(pas)
	bot.RegisterModule(mod)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bot.Run(ctx)

	bot.Discord.Sess.State().GuildAdd(&discordgo.Guild{ID: "111", Channels: []*discordgo.Channel{}})
	bot.Discord.Sess.State().ChannelAdd(&discordgo.Channel{ID: "1234", GuildID: "111"})
	bot.Discord.messageChan <- &DiscordMessage{
		Sess:         bot.Discord.Sess,
		Discord:      bot.Discord,
		MessageType:  MessageTypeCreate,
		TimeReceived: time.Now(),
		Message: &discordgo.Message{
			GuildID: "111",
			Author: &discordgo.User{
				Username: "jeff",
			},
			ChannelID: "1234",
			ID:        "112233",
		},
	}

	select {
	case <-called:
		break
	case <-time.After(time.Millisecond * 10):
		t.Error("Test timed out")
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

	bot.Discord.Sess.State().GuildAdd(&discordgo.Guild{ID: "111", Channels: []*discordgo.Channel{}})
	bot.Discord.Sess.State().ChannelAdd(&discordgo.Channel{ID: "1234", GuildID: "111"})
	bot.Discord.messageChan <- &DiscordMessage{
		Sess:         bot.Discord.Sess,
		Discord:      bot.Discord,
		MessageType:  MessageTypeCreate,
		TimeReceived: time.Now(),
		Message: &discordgo.Message{
			Content: ".test hello",
			GuildID: "111",
			Author: &discordgo.User{
				Username: "jeff",
			},
			ChannelID: "1234",
			ID:        "112233",
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
