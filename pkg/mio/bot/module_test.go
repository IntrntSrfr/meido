package bot

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/pkg/mio"
	"github.com/intrntsrfr/meido/pkg/mio/discord"
)

func TestNewModule(t *testing.T) {
	want := "testing"
	base := NewModule(nil, "testing", mio.NewDiscardLogger())
	if got := base.Name(); got != want {
		t.Errorf("ModuleBase.New() did not produce correct name; got = %v, want %v", got, want)
	}
}

func TestModuleBase_Name(t *testing.T) {
	want := "testing"
	base := NewModule(nil, "testing", mio.NewDiscardLogger())
	if got := base.Name(); got != "testing" {
		t.Errorf("ModuleBase.Name() = %v, want %v", got, want)
	}
}

func TestModuleBase_Passives(t *testing.T) {
	want := 1
	base := NewModule(nil, "testing", mio.NewDiscardLogger())
	base.RegisterPassives(&ModulePassive{Name: "testing"})
	if got := len(base.Passives()); got != 1 {
		t.Errorf("ModuleBase.Passives() = %v, want %v", got, want)
	}
}

func TestModuleBase_Commands(t *testing.T) {
	want := 1
	base := NewModule(nil, "testing", mio.NewDiscardLogger())
	base.RegisterCommands(&ModuleCommand{Name: "testing"})
	if got := len(base.Commands()); got != 1 {
		t.Errorf("ModuleBase.Commands() = %v, want %v", got, want)
	}
}

func TestModuleBase_AllowedTypes(t *testing.T) {
	want := discord.MessageTypeCreate | discord.MessageTypeUpdate
	base := &ModuleBase{allowedTypes: discord.MessageTypeCreate}
	if got := base.AllowedTypes(); got&discord.MessageTypeCreate != discord.MessageTypeCreate {
		t.Errorf("ModuleBase.AllowedTypes() = %v, want %v", got, want)
	}
}

func TestModuleBase_AllowDMs(t *testing.T) {
	want := true
	base := &ModuleBase{allowDMs: true}
	if got := base.AllowDMs(); got != want {
		t.Errorf("ModuleBase.AllowDMs() = %v, want %v", got, want)
	}
}

func TestModuleBase_RegisterPassives(t *testing.T) {
	want := 1
	base := NewModule(nil, "testing", mio.NewDiscardLogger())
	base.RegisterPassives(&ModulePassive{Name: "testing"})
	if got := len(base.Passives()); got != 1 {
		t.Errorf("ModuleBase.Passives() = %v, want %v", got, want)
	}
	if err := base.RegisterPassives(&ModulePassive{Name: "testing2"}, &ModulePassive{Name: "testing2"}); err == nil {
		t.Errorf("ModuleBase.RegisterPassives() did not error on duplicate passive registration")
	}
}

func TestModuleBase_RegisterCommands(t *testing.T) {
	want := 1
	base := NewModule(nil, "testing", mio.NewDiscardLogger())
	base.RegisterCommands(&ModuleCommand{Name: "testing"})
	if got := len(base.Commands()); got != 1 {
		t.Errorf("ModuleBase.Commands() = %v, want %v", got, want)
	}
	if err := base.RegisterCommands(&ModuleCommand{Name: "testing2"}, &ModuleCommand{Name: "testing2"}); err == nil {
		t.Errorf("ModuleBase.RegisterCommands() did not error on duplicate passive registration")
	}
}

func TestModuleBase_FindCommand(t *testing.T) {
	base := NewModule(nil, "testing", mio.NewDiscardLogger())
	cmd := &ModuleCommand{
		Name:     "test",
		Triggers: []string{"m?test", "m?settings test"},
	}
	base.RegisterCommands(cmd)

	type args struct {
		name string
	}
	tests := []struct {
		name    string
		m       *ModuleBase
		args    args
		want    *ModuleCommand
		wantErr bool
	}{
		{
			name:    "positive test 1",
			m:       base,
			args:    args{"m?test"},
			want:    cmd,
			wantErr: false,
		},
		{
			name:    "positive test 2",
			m:       base,
			args:    args{"m?settings test abc"},
			want:    cmd,
			wantErr: false,
		},
		{
			name:    "positive test 3",
			m:       base,
			args:    args{"test"},
			want:    cmd,
			wantErr: false,
		},
		{
			name:    "negative test 1",
			m:       base,
			args:    args{"m?testing"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.m.FindCommand(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("ModuleBase.FindCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ModuleBase.FindCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModuleBase_FindPassive(t *testing.T) {
	base := NewModule(nil, "testing", mio.NewDiscardLogger())
	cmd := &ModulePassive{
		Name: "test",
	}
	base.RegisterPassives(cmd)

	type args struct {
		name string
	}
	tests := []struct {
		name    string
		m       *ModuleBase
		args    args
		want    *ModulePassive
		wantErr bool
	}{
		{
			name:    "positive test",
			m:       base,
			args:    args{"test"},
			want:    cmd,
			wantErr: false,
		},
		{
			name:    "negative test",
			m:       base,
			args:    args{"testing"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "negative test",
			m:       base,
			args:    args{"testing test"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.m.FindPassive(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("ModuleBase.FindPassive() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ModuleBase.FindPassive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModuleBase_AllowsMessage(t *testing.T) {
	m := NewModule(nil, "testing", mio.NewDiscardLogger())
	msg := &discord.DiscordMessage{
		Message:     &discordgo.Message{Type: discordgo.MessageTypeDefault, GuildID: ""},
		MessageType: discord.MessageTypeCreate,
	}

	t.Run("dm ok if allows dms", func(t *testing.T) {
		expected := true
		if got := m.allowsMessage(msg); got != true {
			t.Errorf("Module.AllowsMessage(msg) = %v, want %v", got, expected)
		}
	})

	t.Run("dm not ok if not allows dms", func(t *testing.T) {
		m.allowDMs = false
		expected := true
		if got := m.allowsMessage(msg); got != false {
			t.Errorf("Module.AllowsMessage(msg) = %v, want %v", got, expected)
		}
	})

	m.allowDMs = true
	t.Run("ok if good type", func(t *testing.T) {
		msg.MessageType = discord.MessageTypeCreate | discord.MessageTypeUpdate
		expected := true
		if got := m.allowsMessage(msg); got != true {
			t.Errorf("Module.AllowsMessage(msg) = %v, want %v", got, expected)
		}
	})

	t.Run("not ok if not good type", func(t *testing.T) {
		msg.MessageType = discord.MessageTypeUpdate
		expected := true
		if got := m.allowsMessage(msg); got != false {
			t.Errorf("Module.AllowsMessage(msg) = %v, want %v", got, expected)
		}
	})
}

func TestModuleBase_HandleCommand(t *testing.T) {
	t.Run("it runs correctly", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		bot := NewTestBot()
		go drainBotEvents(ctx, bot.Events())
		mod := NewTestModule(bot, "testing", mio.NewDiscardLogger())
		cmdCalled := make(chan bool, 1)
		cmd := NewTestCommand(mod)
		cmd.Execute = func(dm *discord.DiscordMessage) {
			cmdCalled <- true
		}
		mod.RegisterCommands(cmd)

		msg := NewTestMessage(bot, "1")
		mod.HandleMessage(msg)
		select {
		case <-cmdCalled:
		case <-time.After(time.Millisecond * 50):
			t.Error("Expected event, but timed out")
		}
	})

	t.Run("panic gets handled", func(t *testing.T) {
		bot := NewTestBot()
		mod := NewTestModule(bot, "testing", mio.NewDiscardLogger())
		cmd := NewTestCommand(mod)
		cmd.Execute = func(dm *discord.DiscordMessage) {
			panic("command panic")
		}
		mod.RegisterCommands(cmd)

		msg := NewTestMessage(bot, "1")
		mod.HandleMessage(msg)
		select {
		case <-bot.eventCh:
		case <-time.After(time.Millisecond * 50):
			t.Error("Expected event, but timed out")
		}
	})

	t.Run("DM does not run when DMs not allowed", func(t *testing.T) {
		bot := NewTestBot()
		mod := NewTestModule(bot, "testing", mio.NewDiscardLogger())
		cmdCalled := make(chan bool, 1)
		cmd := NewTestCommand(mod)
		cmd.Execute = func(dm *discord.DiscordMessage) {
			cmdCalled <- true
		}
		mod.RegisterCommands(cmd)

		msg := NewTestMessage(bot, "")
		mod.HandleMessage(msg)
		select {
		case <-cmdCalled:
			t.Errorf("Command was not expected to be called")
		case <-time.After(time.Millisecond * 50):
		}
	})
}

func TestModuleBase_HandlePassive(t *testing.T) {
	t.Run("it runs correctly", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		bot := NewTestBot()
		go drainBotEvents(ctx, bot.Events())
		mod := NewTestModule(bot, "testing", mio.NewDiscardLogger())
		pasCalled := make(chan bool, 1)
		pas := NewTestPassive(mod)
		pas.Execute = func(dm *discord.DiscordMessage) {
			pasCalled <- true
		}
		mod.RegisterPassives(pas)

		msg := NewTestMessage(bot, "1")
		mod.HandleMessage(msg)
		select {
		case <-pasCalled:
		case <-time.After(time.Millisecond * 50):
			t.Error("Expected event, but timed out")
		}
	})

	t.Run("panic gets handled", func(t *testing.T) {
		bot := NewTestBot()
		mod := NewTestModule(bot, "testing", mio.NewDiscardLogger())
		pas := NewTestPassive(mod)
		pas.Execute = func(dm *discord.DiscordMessage) {
			panic("passive panic")
		}
		mod.RegisterPassives(pas)

		msg := NewTestMessage(bot, "1")
		mod.HandleMessage(msg)
		select {
		case <-bot.eventCh:
		case <-time.After(time.Millisecond * 50):
			t.Error("Expected event, but timed out")
		}
	})

	t.Run("DM does not run when DMs not allowed", func(t *testing.T) {
		bot := NewTestBot()
		mod := NewTestModule(bot, "testing", mio.NewDiscardLogger())
		pasCalled := make(chan bool, 1)
		pas := NewTestPassive(mod)
		pas.Execute = func(dm *discord.DiscordMessage) {
			pasCalled <- true
		}
		mod.RegisterPassives(pas)

		msg := NewTestMessage(bot, "")
		mod.HandleMessage(msg)
		select {
		case <-pasCalled:
			t.Errorf("Passive was not expected to be called")
		case <-time.After(time.Millisecond * 50):
		}
	})
}

func TestModuleBase_HandleApplicationCommand(t *testing.T) {
	t.Run("it runs correctly", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		bot := NewTestBot()
		go drainBotEvents(ctx, bot.Events())
		mod := NewTestModule(bot, "testing", mio.NewDiscardLogger())
		cmdCalled := make(chan bool, 1)
		cmd := NewTestApplicationCommand(mod)
		cmd.Execute = func(*discord.DiscordApplicationCommand) {
			cmdCalled <- true
		}
		mod.RegisterApplicationCommands(cmd)

		it := NewTestApplicationCommandInteraction(bot, "1")
		mod.HandleInteraction(it)
		select {
		case <-cmdCalled:
		case <-time.After(time.Millisecond * 50):
			t.Error("Expected event, but timed out")
		}
	})

	t.Run("panic gets handled", func(t *testing.T) {
		bot := NewTestBot()
		mod := NewTestModule(bot, "testing", mio.NewDiscardLogger())
		cmd := NewTestApplicationCommand(mod)
		cmd.Execute = func(*discord.DiscordApplicationCommand) {
			panic("application command panic")
		}
		mod.RegisterApplicationCommands(cmd)

		it := NewTestApplicationCommandInteraction(bot, "1")
		mod.HandleInteraction(it)
		select {
		case <-bot.eventCh:
		case <-time.After(time.Second):
			t.Error("Expected event, but timed out")
		}
	})
}

func TestModuleBase_HandleMessageComponent(t *testing.T) {
	t.Run("callback gets handled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		bot := NewTestBot()
		go drainBotEvents(ctx, bot.Events())
		mod := NewTestModule(bot, "testing", mio.NewDiscardLogger())
		cmdCalled := make(chan bool, 1)
		cmd := NewTestMessageComponent(mod)
		cmd.Execute = func(*discord.DiscordMessageComponent) {
			cmdCalled <- true
		}
		mod.RegisterMessageComponents(cmd)
		customID := "key"
		mod.SetMessageComponentCallback(customID, "test")

		it := NewTestMessageComponentInteraction(bot, "1", customID)
		mod.HandleInteraction(it)
		select {
		case <-cmdCalled:
		case <-time.After(time.Millisecond * 50):
			t.Error("Expected event, but timed out")
		}
	})

	t.Run("by name gets handled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		bot := NewTestBot()
		go drainBotEvents(ctx, bot.Events())
		mod := NewTestModule(bot, "testing", mio.NewDiscardLogger())
		cmdCalled := make(chan bool, 1)
		cmd := NewTestMessageComponent(mod)
		cmd.Execute = func(*discord.DiscordMessageComponent) {
			cmdCalled <- true
		}
		mod.RegisterMessageComponents(cmd)

		it := NewTestMessageComponentInteraction(bot, "1", "test")
		mod.HandleInteraction(it)
		select {
		case <-cmdCalled:
		case <-time.After(time.Millisecond * 50):
			t.Error("Expected event, but timed out")
		}
	})

	t.Run("with suffix gets handled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		bot := NewTestBot()
		go drainBotEvents(ctx, bot.Events())
		mod := NewTestModule(bot, "testing", mio.NewDiscardLogger())
		cmdCalled := make(chan bool, 1)
		cmd := NewTestMessageComponent(mod)
		cmd.Execute = func(*discord.DiscordMessageComponent) {
			cmdCalled <- true
		}
		mod.RegisterMessageComponents(cmd)
		customID := cmd.Name + ":key"

		it := NewTestMessageComponentInteraction(bot, "1", customID)
		mod.HandleInteraction(it)
		select {
		case <-cmdCalled:
		case <-time.After(time.Millisecond * 50):
			t.Error("Expected event, but timed out")
		}
	})

	t.Run("panic gets handled", func(t *testing.T) {
		bot := NewTestBot()
		mod := NewTestModule(bot, "testing", mio.NewDiscardLogger())
		cmd := NewTestMessageComponent(mod)
		cmd.Execute = func(*discord.DiscordMessageComponent) {
			panic("message component panic")
		}
		mod.RegisterMessageComponents(cmd)
		customID := "key"
		mod.SetMessageComponentCallback(customID, "test")

		it := NewTestMessageComponentInteraction(bot, "1", customID)
		mod.HandleInteraction(it)
		select {
		case <-bot.eventCh:
		case <-time.After(time.Second):
			t.Error("Expected event, but timed out")
		}
	})
}

func TestModuleBase_HandleModalSubmit(t *testing.T) {
	t.Run("it runs correctly", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		bot := NewTestBot()
		go drainBotEvents(ctx, bot.Events())
		mod := NewTestModule(bot, "testing", mio.NewDiscardLogger())
		cmdCalled := make(chan bool, 1)
		cmd := NewTestModalSubmit(mod)
		cmd.Execute = func(*discord.DiscordModalSubmit) {
			cmdCalled <- true
		}
		mod.RegisterModalSubmits(cmd)
		customID := "key"
		mod.SetModalSubmitCallback(customID, "test")

		it := NewTestModalSubmitInteraction(bot, "1", customID)
		mod.HandleInteraction(it)
		select {
		case <-cmdCalled:
		case <-time.After(time.Millisecond * 50):
			t.Error("Expected event, but timed out")
		}
	})

	t.Run("panic gets handled", func(t *testing.T) {
		bot := NewTestBot()
		mod := NewTestModule(bot, "testing", mio.NewDiscardLogger())
		cmd := NewTestModalSubmit(mod)
		cmd.Execute = func(*discord.DiscordModalSubmit) {
			panic("modal submit panic")
		}
		mod.RegisterModalSubmits(cmd)
		customID := "key"
		mod.SetModalSubmitCallback(customID, "test")

		it := NewTestModalSubmitInteraction(bot, "1", customID)
		mod.HandleInteraction(it)
		select {
		case <-bot.eventCh:
		case <-time.After(time.Second):
			t.Error("Expected event, but timed out")
		}
	})
}

func TestModuleCommand_CooldownKey(t *testing.T) {
	gid, chid, uid := "1234", "2345", "3456"

	msg := &discord.DiscordMessage{
		Message: &discordgo.Message{
			GuildID:   gid,
			ChannelID: chid,
			Author:    &discordgo.User{ID: uid},
		},
		MessageType: discord.MessageTypeCreate,
	}

	tests := []struct {
		name string
		cmd  *ModuleCommand
		want string
	}{
		{"empty", &ModuleCommand{CooldownScope: -1, Name: "test"}, ""},
		{"user", &ModuleCommand{CooldownScope: CooldownScopeUser, Name: "test"}, fmt.Sprintf("user:%v:%v", uid, "test")},
		{"channel", &ModuleCommand{CooldownScope: CooldownScopeChannel, Name: "test"}, fmt.Sprintf("channel:%v:%v", chid, "test")},
		{"guild", &ModuleCommand{CooldownScope: CooldownScopeGuild, Name: "test"}, fmt.Sprintf("guild:%v:%v", gid, "test")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cmd.CooldownKey(msg); got != tt.want {
				t.Errorf("ModuleCommand.CooldownKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
