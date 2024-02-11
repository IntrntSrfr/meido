package discord

import (
	"github.com/intrntsrfr/meido/pkg/mio/discord/mocks"
	"github.com/intrntsrfr/meido/pkg/mio/test"
	"github.com/intrntsrfr/meido/pkg/utils"
)

func NewTestDiscord(conf *utils.Config, sess DiscordSession) *Discord {
	if conf == nil {
		conf = test.NewTestConfig()
	}
	if sess == nil {
		sess = mocks.NewDiscordSession(conf.GetString("token"), conf.GetInt("shards"))
	}
	d := NewDiscord(conf.GetString("token"), conf.GetInt("shards"), test.NewTestLogger())
	d.Sess = sess
	d.Sessions = []DiscordSession{d.Sess}
	return d
}
