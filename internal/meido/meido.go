package meido

import (
	"github.com/intrntsrfr/meido/internal/database"
	"github.com/intrntsrfr/meido/pkg/mio"
	"go.uber.org/zap"
)

type Meido struct {
	Bot    *mio.Bot
	logger *zap.Logger
}

func New(config mio.Configurable, db database.DB, log *zap.Logger) *Meido {
	meido := &Meido{
		Bot:    mio.NewBot(config, db, log.Named("mio")),
		logger: log,
	}

	//meido.Bot.RegisterModule(administration.New(bot, logger))
	//meido.Bot.RegisterModule(testing.New(bot, logger))
	//meido.Bot.RegisterModule(fun.New(bot, logger))
	//meido.Bot.RegisterModule(fishmod.New())
	//meido.Bot.RegisterModule(utility.New(bot, db, logger))
	//meido.Bot.RegisterModule(moderation.New(bot, db, logger))
	//meido.Bot.RegisterModule(customrole.New(bot, db, logger))
	//meido.Bot.RegisterModule(search.New(bot, logger))
	//meido.Bot.RegisterModule(mediaconvertmod.New())
	//meido.Bot.RegisterModule(aimod.New(gptClient, config.GPT3Engine))

	//meido.Bot.AddEventHandler("command_ran", testHandler(db, log.Named("testHandler")))
	return meido
}

func (m *Meido) Run(useDefHandlers bool) error {
	if err := m.Bot.Open(useDefHandlers); err != nil {
		return err
	}
	if err := m.Bot.Run(); err != nil {
		return err
	}
	return nil
}

func (m *Meido) Close() {
	m.Bot.Close()
}

/*
func testHandler(db database.DB, log *zap.Logger) func(interface{}) {
	return func(i interface{}) {
		cmd, ok := i.(*mio.ModuleCommand)
		if !ok {
			return
		}

		if err := db.CreateCommandLogEntry(&structs.CommandLogEntry{
			Command:   cmd.Name,
			Args:      strings.Join(msg.Args(), " "),
			UserID:    msg.AuthorID(),
			GuildID:   msg.GuildID(),
			ChannelID: msg.ChannelID(),
			MessageID: msg.Message.ID,
			SentAt:    time.Now(),
		}); err != nil {
			log.Error("error logging command", zap.Error(err))
		}
	}
}
*/
