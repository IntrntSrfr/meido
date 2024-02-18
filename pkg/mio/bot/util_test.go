package bot

import (
	"errors"

	"github.com/intrntsrfr/meido/pkg/mio/discord"
	"github.com/intrntsrfr/meido/pkg/mio/test"
	"go.uber.org/zap"
)

func NewTestBot() *Bot {
	bot := NewBotBuilder(test.NewTestConfig(), test.NewTestLogger()).
		WithDiscord(discord.NewTestDiscord(nil, nil, nil)).
		Build()
	bot.UseDefaultHandlers()
	return bot
}

func NewTestModule(bot *Bot, name string, log *zap.Logger) *testModule {
	return &testModule{ModuleBase: *NewModule(bot, name, log)}
}

type testModule struct {
	ModuleBase
	hookShouldFail bool
}

func (m *testModule) Hook() error {
	if m.hookShouldFail {
		return errors.New("Something terrible has happened")
	}
	return nil
}

func NewTestCommand(mod Module) *ModuleCommand {
	return &ModuleCommand{
		Mod:              mod,
		Name:             "test",
		Description:      "testing",
		Triggers:         []string{".test"},
		Usage:            ".test",
		Cooldown:         0,
		CooldownScope:    Channel,
		RequiredPerms:    0,
		CheckBotPerms:    false,
		RequiresUserType: UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         false,
		Enabled:          true,
		Run:              testCommandRun,
	}
}

func testCommandRun(msg *discord.DiscordMessage) {

}

func NewTestPassive(mod Module) *ModulePassive {
	return &ModulePassive{
		Mod:          mod,
		Name:         "test",
		Description:  "testing",
		AllowedTypes: discord.MessageTypeCreate,
		Enabled:      true,
		Run:          testPassiveRun,
	}
}

func testPassiveRun(msg *discord.DiscordMessage) {

}

func NewTestApplicationCommand(mod Module) *ModuleApplicationCommand {
	return &ModuleApplicationCommand{
		Mod:           mod,
		Name:          "test",
		Description:   "testing",
		Cooldown:      0,
		CooldownScope: Channel,
		CheckBotPerms: false,
		AllowDMs:      false,
		Enabled:       true,
		Run:           testApplicationCommandRun,
	}
}

func testApplicationCommandRun(msg *discord.DiscordApplicationCommand) {

}
