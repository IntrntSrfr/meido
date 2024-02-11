package bot

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/pkg/mio/discord"
	"github.com/intrntsrfr/meido/pkg/mio/test"
)

func TestNewModule(t *testing.T) {
	want := "testing"
	base := NewModule(nil, "testing", test.NewTestLogger())
	if got := base.Name(); got != want {
		t.Errorf("ModuleBase.New() did not produce correct name; got = %v, want %v", got, want)
	}
}

func TestModuleBase_Name(t *testing.T) {
	want := "testing"
	base := NewModule(nil, "testing", test.NewTestLogger())
	if got := base.Name(); got != "testing" {
		t.Errorf("ModuleBase.Name() = %v, want %v", got, want)
	}
}

func TestModuleBase_Passives(t *testing.T) {
	want := 1
	base := NewModule(nil, "testing", test.NewTestLogger())
	base.RegisterPassives(&ModulePassive{Name: "testing"})
	if got := len(base.Passives()); got != 1 {
		t.Errorf("ModuleBase.Passives() = %v, want %v", got, want)
	}
}

func TestModuleBase_Commands(t *testing.T) {
	want := 1
	base := NewModule(nil, "testing", test.NewTestLogger())
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

func TestModuleBase_RegisterPassive(t *testing.T) {
	want := 1
	base := NewModule(nil, "testing", test.NewTestLogger())
	base.RegisterPassives(&ModulePassive{Name: "testing"})
	if got := len(base.Passives()); got != 1 {
		t.Errorf("ModuleBase.Passives() = %v, want %v", got, want)
	}
	if err := base.RegisterPassives(&ModulePassive{Name: "testing"}); err == nil {
		t.Errorf("ModuleBase.RegisterPassive() did not error on duplicate passive registration")
	}
}

func TestModuleBase_RegisterPassives(t *testing.T) {
	want := 1
	base := NewModule(nil, "testing", test.NewTestLogger())
	base.RegisterPassives(&ModulePassive{Name: "testing"})
	if got := len(base.Passives()); got != 1 {
		t.Errorf("ModuleBase.Passives() = %v, want %v", got, want)
	}
	if err := base.RegisterPassives(&ModulePassive{Name: "testing2"}, &ModulePassive{Name: "testing2"}); err == nil {
		t.Errorf("ModuleBase.RegisterPassives() did not error on duplicate passive registration")
	}
}

func TestModuleBase_RegisterCommand(t *testing.T) {
	want := 1
	base := NewModule(nil, "testing", test.NewTestLogger())
	base.RegisterCommands(&ModuleCommand{Name: "testing"})
	if got := len(base.Commands()); got != 1 {
		t.Errorf("ModuleBase.Commands() = %v, want %v", got, want)
	}
	if err := base.RegisterCommands(&ModuleCommand{Name: "testing"}); err == nil {
		t.Errorf("ModuleBase.RegisterCommand() did not error on duplicate passive registration")
	}
}

func TestModuleBase_RegisterCommands(t *testing.T) {
	want := 1
	base := NewModule(nil, "testing", test.NewTestLogger())
	base.RegisterCommands(&ModuleCommand{Name: "testing"})
	if got := len(base.Commands()); got != 1 {
		t.Errorf("ModuleBase.Commands() = %v, want %v", got, want)
	}
	if err := base.RegisterCommands(&ModuleCommand{Name: "testing2"}, &ModuleCommand{Name: "testing2"}); err == nil {
		t.Errorf("ModuleBase.RegisterCommands() did not error on duplicate passive registration")
	}
}

func TestModuleBase_FindCommandByName(t *testing.T) {
	base := NewModule(nil, "testing", test.NewTestLogger())
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
			args:    args{"test"},
			want:    cmd,
			wantErr: false,
		},
		{
			name:    "negative test 1",
			m:       base,
			args:    args{"m?test"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "negative test 2",
			m:       base,
			args:    args{"testing"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.m.findCommandByName(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("ModuleBase.FindCommandByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ModuleBase.FindCommandByName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModuleBase_FindCommandByTriggers(t *testing.T) {
	base := NewModule(nil, "testing", test.NewTestLogger())
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
			name:    "negative test 1",
			m:       base,
			args:    args{"test"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "negative test 2",
			m:       base,
			args:    args{"m?testing"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.m.findCommandByTriggers(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("ModuleBase.FindCommandByTriggers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ModuleBase.FindCommandByTriggers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModuleBase_FindCommand(t *testing.T) {
	base := NewModule(nil, "testing", test.NewTestLogger())
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
	base := NewModule(nil, "testing", test.NewTestLogger())
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
	m := NewModule(nil, "testing", test.NewTestLogger())
	msg := &discord.DiscordMessage{
		Message:     &discordgo.Message{Type: discordgo.MessageTypeDefault, GuildID: ""},
		MessageType: discord.MessageTypeCreate,
	}

	t.Run("dm ok if allows dms", func(t *testing.T) {
		expected := true
		if got := m.AllowsMessage(msg); got != true {
			t.Errorf("Module.AllowsMessage(msg) = %v, want %v", got, expected)
		}
	})

	t.Run("dm not ok if not allows dms", func(t *testing.T) {
		m.allowDMs = false
		expected := true
		if got := m.AllowsMessage(msg); got != false {
			t.Errorf("Module.AllowsMessage(msg) = %v, want %v", got, expected)
		}
	})

	m.allowDMs = true
	t.Run("ok if good type", func(t *testing.T) {
		msg.MessageType = discord.MessageTypeCreate | discord.MessageTypeUpdate
		expected := true
		if got := m.AllowsMessage(msg); got != true {
			t.Errorf("Module.AllowsMessage(msg) = %v, want %v", got, expected)
		}
	})

	t.Run("not ok if not good type", func(t *testing.T) {
		msg.MessageType = discord.MessageTypeUpdate
		expected := true
		if got := m.AllowsMessage(msg); got != false {
			t.Errorf("Module.AllowsMessage(msg) = %v, want %v", got, expected)
		}
	})
}

func TestModuleCommand_AllowsMessage(t *testing.T) {
	cmd := &ModuleCommand{AllowedTypes: discord.MessageTypeCreate, RequiredPerms: 0, AllowDMs: true}
	msg := &discord.DiscordMessage{
		Message:     &discordgo.Message{Type: discordgo.MessageTypeDefault, GuildID: ""},
		MessageType: discord.MessageTypeCreate,
	}

	t.Run("dm ok if allows dms", func(t *testing.T) {
		expected := true
		if got := cmd.AllowsMessage(msg); got != expected {
			t.Errorf("ModuleCommand.AllowsMessage(msg) = %v, want %v", got, expected)
		}
	})

	t.Run("dm not ok if not allows dms", func(t *testing.T) {
		cmd.AllowDMs = false
		expected := true
		if got := cmd.AllowsMessage(msg); got != false {
			t.Errorf("ModuleCommand.AllowsMessage(msg) = %v, want %v", got, expected)
		}
	})

	cmd.AllowDMs = true
	t.Run("ok if good type", func(t *testing.T) {
		msg.MessageType = discord.MessageTypeCreate | discord.MessageTypeUpdate
		expected := true
		if got := cmd.AllowsMessage(msg); got != true {
			t.Errorf("ModuleCommand.AllowsMessage(msg) = %v, want %v", got, expected)
		}
	})

	t.Run("not ok if not good type", func(t *testing.T) {
		msg.MessageType = discord.MessageTypeUpdate
		expected := true
		if got := cmd.AllowsMessage(msg); got != false {
			t.Errorf("ModuleCommand.AllowsMessage(msg) = %v, want %v", got, expected)
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
		{"user", &ModuleCommand{CooldownScope: User, Name: "test"}, fmt.Sprintf("user:%v:%v", uid, "test")},
		{"channel", &ModuleCommand{CooldownScope: Channel, Name: "test"}, fmt.Sprintf("channel:%v:%v", chid, "test")},
		{"guild", &ModuleCommand{CooldownScope: Guild, Name: "test"}, fmt.Sprintf("guild:%v:%v", gid, "test")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cmd.CooldownKey(msg); got != tt.want {
				t.Errorf("ModuleCommand.CooldownKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
