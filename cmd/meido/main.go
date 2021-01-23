package main

import (
	"encoding/json"
	"github.com/intrntsrfr/meido"
	"github.com/intrntsrfr/meido/mods/googlemod"
	"github.com/intrntsrfr/meido/mods/helpmod"
	"github.com/intrntsrfr/meido/mods/loggermod"
	"github.com/intrntsrfr/meido/mods/mediaconvertmod"
	"github.com/intrntsrfr/meido/mods/moderationmod"
	"github.com/intrntsrfr/meido/mods/pingmod"
	"github.com/intrntsrfr/meido/mods/searchmod"
	"github.com/intrntsrfr/meido/mods/userrolemod"
	"github.com/intrntsrfr/meido/mods/utilitymod"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"
	_ "net/http/pprof"
)

func main() {

	go http.ListenAndServe("localhost:8070", nil)

	file, err := ioutil.ReadFile("./config.json")
	if err != nil {
		panic("config file not found")
	}
	var config *meido.Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		panic("mangled config file, fix it")
	}
	bot := meido.NewBot(config)
	err = bot.Open()
	if err != nil {
		panic(err)
	}

	bot.RegisterMod(pingmod.New("pings"))
	bot.RegisterMod(loggermod.New("logs"))
	bot.RegisterMod(utilitymod.New("utility"))
	bot.RegisterMod(moderationmod.New("moderation"))
	bot.RegisterMod(userrolemod.New("userrole"))
	bot.RegisterMod(searchmod.New("search"))
	bot.RegisterMod(googlemod.New("google"))
	bot.RegisterMod(helpmod.New("assist"))
	bot.RegisterMod(mediaconvertmod.New("mediaconvert"))
	//bot.RegisterMod(autorolemod.New("autorole"))

	bot.Run()
	defer bot.Close()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	<-sc
}
