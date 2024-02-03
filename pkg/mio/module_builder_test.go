package mio

import (
	"reflect"
	"testing"
	"time"
)

func TestModuleCommandBuilder(t *testing.T) {
	mod := newTestModule(nil, "test", testLogger())

	cmd := &ModuleCommand{
		Mod:              mod,
		Name:             "testing",
		Description:      "i am testing",
		Triggers:         []string{"test1", "test2"},
		Usage:            ".testing",
		Cooldown:         time.Second,
		CooldownScope:    Channel,
		RequiredPerms:    123,
		RequiresUserType: UserTypeBotOwner,
		CheckBotPerms:    true,
		AllowedTypes:     MessageTypeUpdate,
		AllowDMs:         true,
		IsEnabled:        true,
		Run:              nil,
	}

	cmdBuilder := NewModuleCommandBuilder(mod, "testing").
		WithDescription("i am testing").
		WithTriggers("test1", "test2").
		WithUsage(".testing").
		WithCooldown(time.Second, Channel).
		WithRequiredPerms(123).
		WithRequiresBotOwner().
		WithCheckBotPerms().
		WithAllowedTypes(MessageTypeUpdate).
		WithAllowDMs()

	if built := cmdBuilder.Build(); !reflect.DeepEqual(cmd, built) {
		t.Errorf("Built command is not equal to expected")
	}

	rf := func(*DiscordMessage) {}
	cmdBuilder.WithRunFunc(rf)

	if built := cmdBuilder.Build(); built.Run == nil {
		t.Errorf("Built command run function should not be nil")
	}
}

func TestModulePassiveBuilder(t *testing.T) {
	mod := newTestModule(nil, "test", testLogger())

	cmd := &ModulePassive{
		Mod:          mod,
		Name:         "testing",
		Description:  "i am testing",
		AllowedTypes: MessageTypeUpdate,
		Enabled:      true,
		Run:          nil,
	}

	cmdBuilder := NewModulePassiveBuilder(mod, "testing").
		WithDescription("i am testing").
		WithAllowedTypes(MessageTypeUpdate)

	if built := cmdBuilder.Build(); !reflect.DeepEqual(cmd, built) {
		t.Errorf("Built passive is not equal to expected")
	}

	rf := func(*DiscordMessage) {}
	cmdBuilder.WithRunFunc(rf)

	if built := cmdBuilder.Build(); built.Run == nil {
		t.Errorf("Built passive run function should not be nil")
	}
}
