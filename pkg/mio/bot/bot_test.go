package bot

import (
	"bytes"
	"context"
	"strings"
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
	done := make(chan *BotEventData)
	go func() {
		select {
		case evt := <-bot.Events():
			done <- evt
		case <-time.After(time.Second):
			t.Errorf("Timed out ")
		}
	}()
	bot.Emit(BotEventCommandRan, &CommandRan{Command: NewTestCommand(nil)})
	data := <-done
	if data.Type != BotEventCommandRan {
		t.Errorf("Wrong event type; expected %v, got %v", BotEventCommandRan, data.Type)
	}
	got, ok := data.Data.(*CommandRan)
	if !ok {
		t.Error("Expected type cast to not fail.")
	}
	if got.Command.Name != "test" {
		t.Errorf("Expected command name to be %v, got %v", "test", got.Command.Name)
	}
}

func TestBot_Run(t *testing.T) {
	t.Run("session open", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		bot := NewBotBuilder(test.NewTestConfig(), test.NewTestLogger()).Build()
		sessionMock := mocks.NewDiscordSession("test", 1)
		bot.Discord = discord.NewTestDiscord(nil, sessionMock, nil)

		bot.Run(ctx)
		defer bot.Close()

		if !sessionMock.IsOpen {
			t.Errorf("Session should have opened")
		}
	})
	t.Run("session open with good application commands", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var buf bytes.Buffer
		bot := NewBotBuilder(test.NewTestConfig(), test.NewTestLoggerWithBuffer(&buf)).Build()
		sessionMock := mocks.NewDiscordSession("test", 1)
		bot.Discord = discord.NewTestDiscord(nil, sessionMock, nil)

		mod := NewTestModule(bot, "test", test.NewTestLogger())
		mod.RegisterApplicationCommands(&ModuleApplicationCommand{ApplicationCommand: &discordgo.ApplicationCommand{Name: "fishing"}})
		bot.RegisterModule(mod)

		bot.Run(ctx)
		defer bot.Close()

		if !strings.Contains(buf.String(), "fishing") {
			t.Errorf("Expected logs to contain 'fishing', got %v", buf.String())
		}
	})

	t.Run("session open with bad application commands", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var buf bytes.Buffer
		bot := NewBotBuilder(test.NewTestConfig(), test.NewTestLoggerWithBuffer(&buf)).Build()
		sessionMock := mocks.NewDiscordSession("test", 1)
		bot.Discord = discord.NewTestDiscord(nil, sessionMock, nil)

		mod := NewTestModule(bot, "test", test.NewTestLogger())
		mod.RegisterApplicationCommands(&ModuleApplicationCommand{ApplicationCommand: &discordgo.ApplicationCommand{Name: "Fishing"}})
		bot.RegisterModule(mod)

		err := bot.Run(ctx)
		if err == nil {
			t.Errorf("Expected error, but there was none")
		}
		defer bot.Close()
	})
}

func setupTestBot() (*Bot, *zap.Logger, Module) {
	bot := NewTestBot()
	logger := test.NewTestLogger()
	mod := NewTestModule(bot, "testing", logger)
	return bot, logger, mod
}

func TestBot_MessageGetsHandled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bot, _, mod := setupTestBot()
	cmd := NewTestCommand(mod)

	cmdCalled := make(chan bool)
	cmd.Run = func(dm *discord.DiscordMessage) {
		cmdCalled <- true
	}

	mod.RegisterCommands(cmd)
	bot.RegisterModule(mod)
	bot.Run(ctx)
	go drainBotEvents(ctx, bot.Events())
	bot.Discord.Messages() <- NewTestMessage(bot, "1")

	select {
	case <-cmdCalled:
	case <-time.After(time.Second * 1):
		t.Error("Test timed out")
	}
}

func TestBot_InteractionGetsHandled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bot, _, mod := setupTestBot()
	cmd := NewTestApplicationCommand(mod)

	cmdCalled := make(chan bool)
	cmd.Run = func(dm *discord.DiscordApplicationCommand) {
		cmdCalled <- true
	}

	mod.RegisterApplicationCommands(cmd)
	bot.RegisterModule(mod)
	bot.Run(ctx)
	go drainBotEvents(ctx, bot.Events())
	bot.Discord.Interactions() <- NewTestApplicationCommandInteraction(bot, "1")

	select {
	case <-cmdCalled:
	case <-time.After(time.Second * 1):
		t.Error("Test timed out")
	}
}

func TestBot_MessageWrongTypeGetsIgnored(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	bot.RegisterModule(mod)
	bot.RegisterModule(mod2)
	bot.Run(ctx)
	bot.Discord.Messages() <- NewTestMessage(bot, "1")

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

func TestBot_MessageEmptyDoesNotTriggerCommand(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bot, _, _ := setupTestBot()
	mod := NewTestModule(bot, "test", test.NewTestLogger())
	cmd := NewTestCommand(mod)
	cmdCalled := make(chan bool)
	cmd.Run = func(dm *discord.DiscordMessage) {
		cmdCalled <- true
	}
	mod.RegisterCommands(cmd)

	bot.RegisterModule(mod)
	bot.Run(ctx)

	msg := NewTestMessage(bot, "1")
	msg.Message.Content = ""
	bot.Discord.Messages() <- msg

	select {
	case <-cmdCalled:
		t.Error("Command was not expected to be called")
	case <-time.After(time.Millisecond * 10):
		break
	}
}

func TestBot_MessageGetsCallback(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	bot.RegisterModule(mod)
	bot.Run(ctx)
	go drainBotEvents(ctx, bot.Events())

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
