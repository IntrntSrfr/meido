package main

import (
	"github.com/intrntsrfr/meido/internal/database"
	"github.com/intrntsrfr/meido/internal/module/administration"
	"github.com/intrntsrfr/meido/internal/module/customrole"
	"github.com/intrntsrfr/meido/internal/module/fun"
	"github.com/intrntsrfr/meido/internal/module/moderation"
	"github.com/intrntsrfr/meido/internal/module/utility"
	"github.com/intrntsrfr/meido/internal/structs"
	"github.com/intrntsrfr/meido/pkg/mio"
	"os"
	"os/signal"
	"syscall"

	"github.com/intrntsrfr/meido/internal/module/search"
	"github.com/intrntsrfr/meido/internal/module/testing"
	"go.uber.org/zap"
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
	err = bot.Open()
	if err != nil {
		panic(err)
	}

	bot.RegisterModule(administration.New(bot, logger))
	bot.RegisterModule(testing.New(bot, logger))
	bot.RegisterModule(fun.New(bot, logger))
	//bot.RegisterModule(fishmod.New())
	bot.RegisterModule(utility.New(bot, db, logger))
	bot.RegisterModule(moderation.New(bot, db, logger))
	bot.RegisterModule(customrole.New(bot, db, logger))
	bot.RegisterModule(search.New(bot, logger))
	//bot.RegisterModule(mediaconvertmod.New())
	//bot.RegisterModule(aimod.New(gptClient, config.GPT3Engine))

	err = bot.Run()
	if err != nil {
		panic(err)
	}
	defer bot.Close()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	<-sc
}
