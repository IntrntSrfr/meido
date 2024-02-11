package discord

import (
	"github.com/intrntsrfr/meido/pkg/mio/discord/mocks"
	"github.com/intrntsrfr/meido/pkg/mio/test"
	"github.com/intrntsrfr/meido/pkg/utils"
	"go.uber.org/zap"
)

func NewTestDiscord(conf *utils.Config, sess DiscordSession, logger *zap.Logger) *Discord {
	if conf == nil {
		conf = test.NewTestConfig()
	}
	if sess == nil {
		sess = mocks.NewDiscordSession(conf.GetString("token"), conf.GetInt("shards"))
	}
	if logger == nil {
		logger = test.NewTestLogger()
	}
	d := NewDiscord(conf.GetString("token"), conf.GetInt("shards"), logger)
	d.Sess = sess
	d.Sessions = []DiscordSession{d.Sess}
	return d
}
