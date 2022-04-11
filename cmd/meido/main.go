package main

import (
	"encoding/json"
	"github.com/intrntsrfr/meido/base"
	"github.com/intrntsrfr/meido/database"
	"github.com/intrntsrfr/meido/internal/mods/googlemod"
	"github.com/intrntsrfr/meido/internal/mods/loggermod"
	"github.com/intrntsrfr/meido/internal/mods/mediaconvertmod"
	"github.com/intrntsrfr/meido/internal/mods/moderationmod"
	"github.com/intrntsrfr/meido/internal/mods/searchmod"
	"github.com/intrntsrfr/meido/internal/mods/testmod"
	"github.com/intrntsrfr/meido/internal/mods/userrolemod"
	"github.com/intrntsrfr/meido/internal/mods/utilitymod"
	"github.com/intrntsrfr/owo"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"
)

func main() {

	logger, _ := zap.NewProduction()

	f, err := os.Create("./error_log.dat")
	if err != nil {
		panic("cannot create error log file")
	}
	defer f.Close()
	log.SetOutput(f)

	file, err := ioutil.ReadFile("./config.json")
	if err != nil {
		panic("config file not found")
	}
	var config *base.Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		panic("mangled config file, fix it")
	}

	psql, err := sqlx.Connect("postgres", config.ConnectionString)
	if err != nil {
		panic(err)
	}

	db := database.New(psql)
	owoClient := owo.NewClient(config.OwoToken)

	bot := base.NewBot(config, db, logger.Named("meido"))
	err = bot.Open()
	if err != nil {
		panic(err)
	}

	bot.RegisterMod(testmod.New())
	//bot.RegisterMod(fishmod.New())
	bot.RegisterMod(loggermod.New(config.DmLogChannels))
	bot.RegisterMod(utilitymod.New(bot, db))
	bot.RegisterMod(moderationmod.New(bot, db, logger.Named("moderation")))
	bot.RegisterMod(userrolemod.New(bot, db, owoClient))
	bot.RegisterMod(searchmod.New(config.YouTubeToken))
	bot.RegisterMod(googlemod.New(bot))
	bot.RegisterMod(mediaconvertmod.New())

	bot.Run()
	defer bot.Close()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	<-sc
}
