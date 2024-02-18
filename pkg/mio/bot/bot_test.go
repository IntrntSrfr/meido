package bot

import (
	"context"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/pkg/mio/discord"
	"github.com/intrntsrfr/meido/pkg/mio/discord/mocks"
	"github.com/intrntsrfr/meido/pkg/mio/test"
	"go.uber.org/zap"
)

func TestBot_IsOwner(t *testing.T) {
	conf := test.NewTestConfig()
	conf.Set("owner_ids", []string{"123"})

	b := NewBotBuilder(conf, test.NewTestLogger()).Build()
	if ok := b.IsOwner("123"); !ok {
		t.Errorf("Bot.IsOwner('123') = %v, want %v", ok, true)
	}

	if ok := b.IsOwner("456"); ok {
		t.Errorf("Bot.IsOwner('456') = %v, want %v", ok, false)
	}
}

func TestBot_RegisterModule(t *testing.T) {
	bot := NewBotBuilder(test.NewTestConfig(), test.NewTestLogger()).Build()
	bot.RegisterModule(NewTestModule(bot, "test", test.NewTestLogger()))
	if len(bot.Modules) != 1 {
		t.Errorf("Bot does not have a module after registering one")
	}
}

func TestBot_Events(t *testing.T) {
	bot := NewBotBuilder(test.NewTestConfig(), test.NewTestLogger()).Build()
	done := make(chan bool)
	go func() {
		select {
		case <-bot.Events():
			done <- true
		case <-time.After(time.Second):
			t.Errorf("Timed out ")
		}
	}()
	bot.Emit(BotEventCommandRan, nil)
	<-done
}

func TestBot_Run(t *testing.T) {
	bot := NewBotBuilder(test.NewTestConfig(), test.NewTestLogger()).Build()
	sessionMock := mocks.NewDiscordSession("asdf", 1)
	bot.Discord = discord.NewTestDiscord(nil, sessionMock, nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bot.Run(ctx)
	defer bot.Close()

	if !sessionMock.IsOpen {
		t.Errorf("Session should have opened")
	}
}

func setupTestBot() (*Bot, *zap.Logger, Module) {
	bot := NewTestBot()
	logger := test.NewTestLogger()
	mod := NewTestModule(bot, "testing", logger)
	return bot, logger, mod
}

func executeNewTestCommand(bot *Bot, mod Module, cmd *ModuleCommand, message *discord.DiscordMessage) (chan bool, context.CancelFunc) {
	called := make(chan bool)

	mod.RegisterCommands(cmd)
	bot.RegisterModule(mod)
	ctx, cancel := context.WithCancel(context.Background())
	bot.Run(ctx)

	bot.Discord.Messages() <- message
	return called, cancel
}

func TestBot_MessageGetsHandled(t *testing.T) {
	bot, _, mod := setupTestBot()
	pas := NewTestPassive(mod)
	cmd := NewTestCommand(mod)

	cmdCalled := make(chan bool)
	cmd.Run = func(dm *discord.DiscordMessage) {
		cmdCalled <- true
	}

	pasCalled := make(chan bool)
	pas.Run = func(dm *discord.DiscordMessage) {
		pasCalled <- true
	}

	// register and run
	mod.RegisterPassives(pas)
	mod.RegisterCommands(cmd)
	bot.RegisterModule(mod)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bot.Run(ctx)

	bot.Discord.Messages() <- &discord.DiscordMessage{
		Sess:         bot.Discord.Sess,
		Discord:      bot.Discord,
		MessageType:  discord.MessageTypeCreate,
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

	mod := NewTestModule(bot, "test", test.NewTestLogger())
	mod.allowedTypes = discord.MessageTypeCreate | discord.MessageTypeUpdate

	pas := NewTestPassive(mod)
	cmd := NewTestCommand(mod)
	cmdCalled := make(chan bool)
	cmd.Run = func(dm *discord.DiscordMessage) {
		cmdCalled <- true
	}
	pasCalled := make(chan bool)
	pas.Run = func(dm *discord.DiscordMessage) {
		pasCalled <- true
	}
	mod.RegisterPassives(pas)
	mod.RegisterCommands(cmd)

	mod2 := NewTestModule(bot, "test2", test.NewTestLogger())
	pas2 := NewTestPassive(mod)
	pas2Called := make(chan bool)
	pas.Run = func(dm *discord.DiscordMessage) {
		pas2Called <- true
	}
	mod2.RegisterPassives(pas2)

	// register and run
	bot.RegisterModule(mod)
	bot.RegisterModule(mod2)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bot.Run(ctx)

	bot.Discord.Messages() <- &discord.DiscordMessage{
		Sess:         bot.Discord.Sess,
		Discord:      bot.Discord,
		MessageType:  discord.MessageTypeUpdate,
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
		bot    = NewTestBot()
		logger = test.NewTestLogger()
		mod    = NewTestModule(bot, "testing", logger)
		cmd    = NewTestCommand(mod)
	)

	// change the command
	cmd.Run = func(dm *discord.DiscordMessage) {
		panic("i am PANICKING !!!")
	}

	// register and run
	mod.RegisterCommands(cmd)
	bot.RegisterModule(mod)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bot.Run(ctx)

	bot.Discord.Sess.State().GuildAdd(&discordgo.Guild{ID: "1", Channels: []*discordgo.Channel{}})
	bot.Discord.Sess.State().ChannelAdd(&discordgo.Channel{ID: "1", GuildID: "1"})
	bot.Discord.Messages() <- &discord.DiscordMessage{
		Sess:         bot.Discord.Sess,
		Discord:      bot.Discord,
		MessageType:  discord.MessageTypeCreate,
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
	case evt := <-bot.Events():
		if evt.Type != BotEventCommandPanicked {
			t.Errorf("Expected panic handler to run")
		}
	case <-time.After(time.Millisecond * 10):
		t.Error("Test timed out")
	}
}

func TestBot_MessageEmptyDoesNotTriggerCommand(t *testing.T) {
	bot, _, _ := setupTestBot()

	mod := NewTestModule(bot, "test", test.NewTestLogger())

	cmd := NewTestCommand(mod)
	cmdCalled := make(chan bool)
	cmd.Run = func(dm *discord.DiscordMessage) {
		cmdCalled <- true
	}
	mod.RegisterCommands(cmd)

	// register and run
	bot.RegisterModule(mod)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bot.Run(ctx)

	bot.Discord.Messages() <- &discord.DiscordMessage{
		Sess:         bot.Discord.Sess,
		Discord:      bot.Discord,
		MessageType:  discord.MessageTypeCreate,
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

	mod := NewTestModule(bot, "test", test.NewTestLogger())

	cmd := NewTestCommand(mod)
	cmdCalled := make(chan bool)
	cmd.Run = func(dm *discord.DiscordMessage) {
		cb, err := mod.Bot.Callbacks.Make(dm.CallbackKey())
		if err != nil {
			t.Error(err)
		}
		select {
		case <-cb:
			cmdCalled <- true
		case <-time.After(time.Millisecond * 50):
			return
		}
	}
	mod.RegisterCommands(cmd)

	// register and run
	bot.RegisterModule(mod)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bot.Run(ctx)

	bot.Discord.Messages() <- &discord.DiscordMessage{
		Sess:         bot.Discord.Sess,
		Discord:      bot.Discord,
		MessageType:  discord.MessageTypeCreate,
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
	bot.Discord.Messages() <- &discord.DiscordMessage{
		Sess:         bot.Discord.Sess,
		Discord:      bot.Discord,
		MessageType:  discord.MessageTypeCreate,
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
