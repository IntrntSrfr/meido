package main

import (
	"encoding/json"
	"github.com/intrntsrfr/meidov2"
	"github.com/intrntsrfr/meidov2/mods/loggermod"
	"github.com/intrntsrfr/meidov2/mods/moderationmod"
	"github.com/intrntsrfr/meidov2/mods/pingmod"
	"github.com/intrntsrfr/meidov2/mods/userrolemod"
	"github.com/intrntsrfr/meidov2/mods/utilitymod"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	file, err := ioutil.ReadFile("./config.json")
	if err != nil {
		panic("Config file not found.\nPlease press enter.")
	}
	var config *meidov2.Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		panic("mangled config file, fix it")
	}

	bot := meidov2.NewBot(config)
	err = bot.Open()
	if err != nil {
		panic(err)
	}

	bot.RegisterMod(pingmod.New(), "ping")
	bot.RegisterMod(loggermod.New(), "logs")
	bot.RegisterMod(utilitymod.New(), "utility")
	bot.RegisterMod(moderationmod.New(), "moderation")
	bot.RegisterMod(userrolemod.New(), "userrole")

	bot.Run()
	defer bot.Close()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
