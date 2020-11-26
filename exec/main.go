package main

import (
	"encoding/json"
	"github.com/intrntsrfr/meidov2"
	"github.com/intrntsrfr/meidov2/mods/googlemod"
	"github.com/intrntsrfr/meidov2/mods/helpmod"
	"github.com/intrntsrfr/meidov2/mods/loggermod"
	"github.com/intrntsrfr/meidov2/mods/mediaconvertmod"
	"github.com/intrntsrfr/meidov2/mods/moderationmod"
	"github.com/intrntsrfr/meidov2/mods/pingmod"
	"github.com/intrntsrfr/meidov2/mods/searchmod"
	"github.com/intrntsrfr/meidov2/mods/userrolemod"
	"github.com/intrntsrfr/meidov2/mods/utilitymod"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "net/http/pprof"
)

func main() {

	go http.ListenAndServe("localhost:8070", nil)

	file, err := ioutil.ReadFile("./config.json")
	if err != nil {
		panic("config file not found")
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

	bot.RegisterMod(pingmod.New("pinger"))
	bot.RegisterMod(loggermod.New("logs"))
	bot.RegisterMod(utilitymod.New("utility"))
	bot.RegisterMod(moderationmod.New("moderation"))
	bot.RegisterMod(userrolemod.New("userrole"))
	bot.RegisterMod(searchmod.New("search"))
	bot.RegisterMod(googlemod.New("google"))
	bot.RegisterMod(helpmod.New("helper"))
	bot.RegisterMod(mediaconvertmod.New("mediaconvert"))

	bot.Run()
	defer bot.Close()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	<-sc
}
