package bot

import (
	"reflect"
	"testing"
	"time"

	"github.com/intrntsrfr/meido/pkg/mio/discord"
	"github.com/intrntsrfr/meido/pkg/mio/test"
)

func TestModuleCommandBuilder(t *testing.T) {
	mod := NewTestModule(nil, "test", test.NewTestLogger())

	cmd := &ModuleCommand{
		Mod:              mod,
		Name:             "testing",
		Description:      "i am testing",
		Triggers:         []string{"test1", "test2"},
		Usage:            ".testing",
		Cooldown:         time.Second,
		CooldownScope:    CooldownScopeChannel,
		RequiredPerms:    123,
		RequiresUserType: UserTypeBotOwner,
		CheckBotPerms:    true,
		AllowedTypes:     discord.MessageTypeUpdate,
		AllowDMs:         true,
		Enabled:          true,
		Run:              nil,
	}

	cmdBuilder := NewModuleCommandBuilder(mod, "testing").
		WithDescription("i am testing").
		WithTriggers("test1", "test2").
		WithUsage(".testing").
		WithCooldown(time.Second, CooldownScopeChannel).
		WithRequiredPerms(123).
		WithRequiresBotOwner().
		WithCheckBotPerms().
		WithAllowedTypes(discord.MessageTypeUpdate).
		WithAllowDMs()

	if built := cmdBuilder.Build(); !reflect.DeepEqual(cmd, built) {
		t.Errorf("Built command is not equal to expected")
	}

	rf := func(*discord.DiscordMessage) {}
	cmdBuilder.WithRunFunc(rf)

	if built := cmdBuilder.Build(); built.Run == nil {
		t.Errorf("Built command run function should not be nil")
	}
}

func TestModulePassiveBuilder(t *testing.T) {
	mod := NewTestModule(nil, "test", test.NewTestLogger())

	cmd := &ModulePassive{
		Mod:          mod,
		Name:         "testing",
		Description:  "i am testing",
		AllowedTypes: discord.MessageTypeUpdate,
		Enabled:      true,
		Run:          nil,
	}

	cmdBuilder := NewModulePassiveBuilder(mod, "testing").
		WithDescription("i am testing").
		WithAllowedTypes(discord.MessageTypeUpdate)

	if built := cmdBuilder.Build(); !reflect.DeepEqual(cmd, built) {
		t.Errorf("Built passive is not equal to expected")
	}

	rf := func(*discord.DiscordMessage) {}
	cmdBuilder.WithRunFunc(rf)

	if built := cmdBuilder.Build(); built.Run == nil {
		t.Errorf("Built passive run function should not be nil")
	}
}

/*
func TestModuleApplicationCommandBuilder(t *testing.T) {
	mod := NewTestModule(nil, "test", test.NewTestLogger())

	cmd := &ModuleApplicationCommand{
		Mod: mod,
		ApplicationCommand: &discordgo.ApplicationCommand{
			Type:        discordgo.ChatApplicationCommand,
			Name:        "test",
			Description: "testing",
			Options: []*discordgo.ApplicationCommandOption{{
				Name: "test",
			}},
		},
		Enabled: true,
		Run:     nil,
	}

	rf := func(*discord.DiscordApplicationCommand) {}
	bld := NewModuleApplicationCommandBuilder(mod, "test").
		Type(discordgo.ChatApplicationCommand).
		Description("testing").
		AddOption(&discordgo.ApplicationCommandOption{
			Name: "test",
		}).
		Cooldown(time.Second, CooldownScopeChannel).
		NoDM().
		Permissions(1).
		CheckBotPerms().
		Run(rf)

	if built := bld.Build(); !reflect.DeepEqual(cmd, built) {
		t.Errorf("Build is not as expected: %v, want %v", built, cmd)
	}

	if built := bld.Build(); built.Run == nil {
		t.Errorf("run function should not be nil")
	}
}
*/
