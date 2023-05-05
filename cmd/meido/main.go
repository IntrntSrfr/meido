package main

import (
	"github.com/intrntsrfr/meido/internal/database"
	"github.com/intrntsrfr/meido/internal/structs"
	"github.com/intrntsrfr/meido/pkg/mio"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	logger, _ := zap.NewProduction()
	logger = logger.Named("meido")

	cfg, err := structs.LoadConfig()
	if err != nil {
		panic(err)
	}

	db, err := database.NewPSQLDatabase(cfg.GetString("connection_string"))
	if err != nil {
		panic(err)
	}

	//owoClient := owo.NewClient(config.OwoToken)
	//gptClient := gogpt.NewClient(config.OpenAIToken)

	bot := mio.NewBot(cfg, db, logger)
	err = bot.Open(true)
	if err != nil {
		panic(err)
	}

	//bot.RegisterModule(administration.New(bot, logger))
	//bot.RegisterModule(testing.New(bot, logger))
	//bot.RegisterModule(fun.New(bot, logger))
	//bot.RegisterModule(fishmod.New())
	//bot.RegisterModule(utility.New(bot, db, logger))
	//bot.RegisterModule(moderation.New(bot, db, logger))
	//bot.RegisterModule(customrole.New(bot, db, logger))
	//bot.RegisterModule(search.New(bot, logger))
	//bot.RegisterModule(mediaconvertmod.New())
	//bot.RegisterModule(aimod.New(gptClient, config.GPT3Engine))

	//bot.AddEventHandler("command_ran", testHandler(db, logger))

	err = bot.Run()
	if err != nil {
		panic(err)
	}
	defer bot.Close()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	<-sc
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
