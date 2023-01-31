package main

import (
	"encoding/json"
	"github.com/intrntsrfr/meido/internal/database"
	"github.com/intrntsrfr/meido/internal/service/search"
	"github.com/intrntsrfr/meido/pkg/mio"
	"os"
	"os/signal"
	"syscall"

	"github.com/intrntsrfr/meido/internal/module/searchmod"
	"github.com/intrntsrfr/meido/internal/module/testmod"
	"go.uber.org/zap"

	_ "github.com/lib/pq"
)

func main() {
	logger, _ := zap.NewProduction()

	file, err := os.ReadFile("./config.json")
	if err != nil {
		panic("config file not found")
	}
	var config *mio.Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		panic("mangled config file, fix it")
	}

	db, err := database.NewPSQLDatabase(config.ConnectionString)
	if err != nil {
		panic(err)
	}

	//owoClient := owo.NewClient(config.OwoToken)
	searchService := search.NewSearchService(config.YouTubeToken, config.OpenWeatherApiKey)
	//gptClient := gogpt.NewClient(config.OpenAIToken)

	bot := mio.NewBot(config, db, logger.Named("meido"))
	err = bot.Open()
	if err != nil {
		panic(err)
	}

	bot.RegisterMod(testmod.New())
	//bot.RegisterMod(fishmod.New())
	//bot.RegisterMod(loggermod.New(config.DmLogChannels))
	//bot.RegisterMod(utilitymod.New(bot, db))
	//bot.RegisterMod(moderationmod.New(bot, db, logger.Named("moderation")))
	//bot.RegisterMod(userrolemod.New(bot, db, owoClient, logger.Named("userrole")))
	bot.RegisterMod(searchmod.New(bot, searchService))
	//bot.RegisterMod(mediaconvertmod.New())
	//bot.RegisterMod(aimod.New(gptClient, config.GPT3Engine))

	err = bot.Run()
	if err != nil {
		panic(err)
	}
	defer bot.Close()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	<-sc
}
