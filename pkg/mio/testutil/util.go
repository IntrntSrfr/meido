package testutil

/*
func NewTestConfig() *utils.Config {
	conf := utils.NewConfig()
	conf.Set("shards", 1)
	conf.Set("token", "asdf")
	return conf
}

func NewTestDiscord(conf *utils.Config, sess mio.DiscordSession, logger mio.Logger) *mio.Discord {
	if conf == nil {
		conf = NewTestConfig()
	}
	if sess == nil {
		sess = mocks.NewDiscordSession(conf.GetString("token"), conf.GetInt("shards"))
	}
	if logger == nil {
		logger = mio.NewLogger(io.Discard)
	}
	d := mio.NewDiscord(conf.GetString("token"), conf.GetInt("shards"), logger)
	d.Sess = sess
	d.Sessions = []mio.DiscordSession{d.Sess}
	return d
}
*/
