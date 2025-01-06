package mio_test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/pkg/mio"
)

func TestBot_IsOwner(t *testing.T) {
	conf := NewTestConfig()
	conf.Set("owner_ids", []string{"123"})

	b := mio.NewBotBuilder(conf).
		WithLogger(mio.NewDiscardLogger()).
		Build()
	if ok := b.IsOwner("123"); !ok {
		t.Errorf("Bot.IsOwner('123') = %v, want %v", ok, true)
	}

	if ok := b.IsOwner("456"); ok {
		t.Errorf("Bot.IsOwner('456') = %v, want %v", ok, false)
	}
}

func TestBot_RegisterModule(t *testing.T) {
	bot := mio.NewBotBuilder(NewTestConfig()).
		WithLogger(mio.NewDiscardLogger()).
		Build()
	bot.RegisterModule(NewTestModule(bot, "test", mio.NewDiscardLogger()))
	if len(bot.Modules) != 1 {
		t.Errorf("Bot does not have a module after registering one")
	}
}

func TestBot_Run(t *testing.T) {
	t.Run("session open", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		bot := mio.NewBotBuilder(NewTestConfig()).
			WithLogger(mio.NewDiscardLogger()).
			Build()
		sessionMock := NewDiscordSession("test", 1)
		bot.Discord = NewTestDiscord(nil, sessionMock, nil)

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
		bot := mio.NewBotBuilder(NewTestConfig()).
			WithLogger(mio.NewLogger(&buf)).
			Build()
		sessionMock := NewDiscordSession("test", 1)
		bot.Discord = NewTestDiscord(nil, sessionMock, nil)

		mod := NewTestModule(bot, "test", mio.NewDiscardLogger())
		mod.RegisterApplicationCommands(&mio.ModuleApplicationCommand{ApplicationCommand: &discordgo.ApplicationCommand{Name: "fishing"}})
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
		bot := mio.NewBotBuilder(NewTestConfig()).
			WithLogger(mio.NewLogger(&buf)).
			Build()
		sessionMock := NewDiscordSession("test", 1)
		bot.Discord = NewTestDiscord(nil, sessionMock, nil)

		mod := NewTestModule(bot, "test", mio.NewDiscardLogger())
		mod.RegisterApplicationCommands(&mio.ModuleApplicationCommand{ApplicationCommand: &discordgo.ApplicationCommand{Name: "Fishing"}})
		bot.RegisterModule(mod)

		err := bot.Run(ctx)
		if err == nil {
			t.Errorf("Expected error, but there was none")
		}
		defer bot.Close()
	})
}

func setupTestBot() (*mio.Bot, mio.Logger, mio.Module) {
	bot := NewTestBot()
	logger := mio.NewDiscardLogger()
	mod := NewTestModule(bot, "testing", logger)
	return bot, logger, mod
}

func TestBot_MessageGetsHandled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bot, _, mod := setupTestBot()
	cmd := NewTestCommand(mod)

	cmdCalled := make(chan bool)
	cmd.Execute = func(dm *mio.DiscordMessage) {
		cmdCalled <- true
	}

	mod.RegisterCommands(cmd)
	bot.RegisterModule(mod)
	bot.Run(ctx)
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
	cmd.Execute = func(dm *mio.DiscordApplicationCommand) {
		cmdCalled <- true
	}

	mod.RegisterApplicationCommands(cmd)
	bot.RegisterModule(mod)
	bot.Run(ctx)
	bot.Discord.Interactions() <- NewTestApplicationCommandInteraction(bot, "1")

	select {
	case <-cmdCalled:
	case <-time.After(time.Second * 1):
		t.Error("Test timed out")
	}
}

func TestBot_MessageWrongTypeGetsIgnored(t *testing.T) {
	t.Run("Command", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		bot, _, _ := setupTestBot()
		mod := NewTestModule(bot, "test", mio.NewDiscardLogger())
		mod.allowedTypes = mio.MessageTypeUpdate

		called := make(chan bool)
		cmd := mio.NewModuleCommandBuilder(mod, "test").
			Triggers(".test").
			AllowedTypes(mio.MessageTypeCreate).
			Execute(func(dm *mio.DiscordMessage) {
				called <- true
			}).Build()
		mod.RegisterCommands(cmd)

		bot.RegisterModule(mod)
		bot.Run(ctx)
		bot.Discord.Messages() <- NewTestMessage(bot, "1")

		select {
		case <-called:
			t.Error("Command was not expected to be called")
		case <-time.After(time.Millisecond * 50):
		}
	})

	t.Run("Passive", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		bot, _, _ := setupTestBot()
		mod := NewTestModule(bot, "test", mio.NewDiscardLogger())
		mod.allowedTypes = mio.MessageTypeUpdate

		called := make(chan bool)
		passive := mio.NewModulePassiveBuilder(mod, "test").
			AllowedTypes(mio.MessageTypeCreate).
			Execute(func(dm *mio.DiscordMessage) {
				called <- true
			}).Build()
		mod.RegisterPassives(passive)

		bot.RegisterModule(mod)
		bot.Run(ctx)
		bot.Discord.Messages() <- NewTestMessage(bot, "1")

		select {
		case <-called:
			t.Error("Passive was not expected to be called")
		case <-time.After(time.Millisecond * 50):
		}
	})
}

func TestBot_MessageEmptyDoesNotTriggerCommand(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bot, _, _ := setupTestBot()
	mod := NewTestModule(bot, "test", mio.NewDiscardLogger())
	cmd := NewTestCommand(mod)
	cmdCalled := make(chan bool)
	cmd.Execute = func(dm *mio.DiscordMessage) {
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
	mod := NewTestModule(bot, "test", mio.NewDiscardLogger())
	cmd := NewTestCommand(mod)
	cmdCalled := make(chan bool)
	cmd.Execute = func(dm *mio.DiscordMessage) {
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

	bot.Discord.Messages() <- &mio.DiscordMessage{
		Sess:         bot.Discord.Sess,
		Discord:      bot.Discord,
		MessageType:  mio.MessageTypeCreate,
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
	bot.Discord.Messages() <- &mio.DiscordMessage{
		Sess:         bot.Discord.Sess,
		Discord:      bot.Discord,
		MessageType:  mio.MessageTypeCreate,
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
