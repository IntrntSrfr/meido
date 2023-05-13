package mio

import (
	"reflect"
	"testing"
)

func TestNewModule(t *testing.T) {
	want := "testing"
	base := NewModule(nil, "testing", nil)
	if got := base.Name(); got != want {
		t.Errorf("ModuleBase.New() did not produce correct name; got = %v, want %v", got, want)
	}
}

func TestModuleBase_Name(t *testing.T) {
	want := "testing"
	base := NewModule(nil, "testing", nil)
	if got := base.Name(); got != "testing" {
		t.Errorf("ModuleBase.Name() = %v, want %v", got, want)
	}
}

func TestModuleBase_Passives(t *testing.T) {
	want := 1
	base := NewModule(nil, "testing", nil)
	base.RegisterPassive(&ModulePassive{Name: "testing"})
	if got := len(base.Passives()); got != 1 {
		t.Errorf("ModuleBase.Passives() = %v, want %v", got, want)
	}
}

func TestModuleBase_Commands(t *testing.T) {
	want := 1
	base := NewModule(nil, "testing", nil)
	base.RegisterCommand(&ModuleCommand{Name: "testing"})
	if got := len(base.Commands()); got != 1 {
		t.Errorf("ModuleBase.Commands() = %v, want %v", got, want)
	}
}

func TestModuleBase_AllowedTypes(t *testing.T) {
	want := MessageTypeCreate | MessageTypeUpdate
	base := &ModuleBase{allowedTypes: MessageTypeCreate}
	if got := base.AllowedTypes(); got&MessageTypeCreate != MessageTypeCreate {
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
	base := NewModule(nil, "testing", nil)
	base.RegisterPassive(&ModulePassive{Name: "testing"})
	if got := len(base.Passives()); got != 1 {
		t.Errorf("ModuleBase.Passives() = %v, want %v", got, want)
	}
	if err := base.RegisterPassive(&ModulePassive{Name: "testing"}); err == nil {
		t.Errorf("ModuleBase.RegisterPassive() did not error on duplicate passive registration")
	}
}

func TestModuleBase_RegisterPassives(t *testing.T) {
	want := 1
	base := NewModule(nil, "testing", nil)
	base.RegisterPassives([]*ModulePassive{{Name: "testing"}})
	if got := len(base.Passives()); got != 1 {
		t.Errorf("ModuleBase.Passives() = %v, want %v", got, want)
	}
	if err := base.RegisterPassives([]*ModulePassive{{Name: "testing2"}, {Name: "testing2"}}); err == nil {
		t.Errorf("ModuleBase.RegisterPassives() did not error on duplicate passive registration")
	}
}

func TestModuleBase_RegisterCommand(t *testing.T) {
	want := 1
	base := NewModule(nil, "testing", nil)
	base.RegisterCommand(&ModuleCommand{Name: "testing"})
	if got := len(base.Commands()); got != 1 {
		t.Errorf("ModuleBase.Commands() = %v, want %v", got, want)
	}
	if err := base.RegisterCommand(&ModuleCommand{Name: "testing"}); err == nil {
		t.Errorf("ModuleBase.RegisterCommand() did not error on duplicate passive registration")
	}
}

func TestModuleBase_RegisterCommands(t *testing.T) {
	want := 1
	base := NewModule(nil, "testing", nil)
	base.RegisterCommands([]*ModuleCommand{{Name: "testing"}})
	if got := len(base.Commands()); got != 1 {
		t.Errorf("ModuleBase.Commands() = %v, want %v", got, want)
	}
	if err := base.RegisterCommands([]*ModuleCommand{{Name: "testing2"}, {Name: "testing2"}}); err == nil {
		t.Errorf("ModuleBase.RegisterCommands() did not error on duplicate passive registration")
	}
}

func TestModuleBase_FindCommandByName(t *testing.T) {
	base := NewModule(nil, "testing", nil)
	cmd := &ModuleCommand{
		Name:     "test",
		Triggers: []string{"m?test", "m?settings test"},
	}
	base.RegisterCommand(cmd)

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
			got, err := tt.m.FindCommandByName(tt.args.name)
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
	base := NewModule(nil, "testing", nil)
	cmd := &ModuleCommand{
		Name:     "test",
		Triggers: []string{"m?test", "m?settings test"},
	}
	base.RegisterCommand(cmd)

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
			got, err := tt.m.FindCommandByTriggers(tt.args.name)
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
	base := NewModule(nil, "testing", nil)
	cmd := &ModuleCommand{
		Name:     "test",
		Triggers: []string{"m?test", "m?settings test"},
	}
	base.RegisterCommand(cmd)

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
	base := NewModule(nil, "testing", nil)
	cmd := &ModulePassive{
		Name: "test",
	}
	base.RegisterPassive(cmd)

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
