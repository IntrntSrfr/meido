package main

import (
	"encoding/json"
	"github.com/intrntsrfr/meido/internal/base"
	"github.com/intrntsrfr/meido/internal/mods/utilitymod"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"
)

func main() {

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
	bot := base.NewBot(config)
	err = bot.Open()
	if err != nil {
		panic(err)
	}

	//bot.RegisterMod(pingmod.New("pings"))
	//bot.RegisterMod(fishmod.New("fishing"))
	//bot.RegisterMod(loggermod.New("logs"))
	bot.RegisterMod(utilitymod.New("utility"))
	//bot.RegisterMod(moderationmod.New("moderation"))
	//bot.RegisterMod(userrolemod.New("userrole"))
	//.RegisterMod(searchmod.New("search"))
	//bot.RegisterMod(googlemod.New("google"))
	//bot.RegisterMod(mediaconvertmod.New("mediaconvert"))

	bot.Run()
	defer bot.Close()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	<-sc
}
