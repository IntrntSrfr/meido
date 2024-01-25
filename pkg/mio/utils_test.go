package mio

import (
	"errors"

	"github.com/intrntsrfr/meido/pkg/mio/mocks"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func testConfig() Configurable {
	conf := NewConfig()
	conf.Set("shards", 1)
	conf.Set("token", "asdf")
	return conf
}

func testLogger() *zap.Logger {
	loggerConfig := zap.NewDevelopmentConfig()
	loggerConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	loggerConfig.OutputPaths = []string{}
	loggerConfig.ErrorOutputPaths = []string{}
	logger, _ := loggerConfig.Build()
	return logger.Named("test")
}

func testBot() *Bot {
	b := NewBot(testConfig(), testLogger())
	// more stuff
	return b
}

func newTestModule(bot *Bot, name string, log *zap.Logger) *testModule {
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

func testDiscord(sess DiscordSession) *Discord {
	if sess == nil {
		sess = mocks.NewDiscordSession("Bot asdf")
	}
	d := NewDiscord("Bot asdf", 1, testLogger())
	d.Sess = sess
	d.Sessions = []DiscordSession{d.Sess}
	return d
}

func testCommand(mod Module) *ModuleCommand {
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
		AllowedTypes:     MessageTypeCreate,
		AllowDMs:         false,
		IsEnabled:        true,
		Run:              testCommandRun,
	}
}

func testCommandRun(msg *DiscordMessage) {

}

func testPassive(mod Module) *ModulePassive {
	return &ModulePassive{
		Mod:          mod,
		Name:         "test",
		Description:  "testing",
		AllowedTypes: MessageTypeCreate,
		Run:          testPassiveRun,
	}
}

func testPassiveRun(msg *DiscordMessage) {

}
