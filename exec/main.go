package main

import (
	"encoding/json"
	"github.com/intrntsrfr/meidov2"
	"github.com/intrntsrfr/meidov2/mods/googlemod"
	"github.com/intrntsrfr/meidov2/mods/loggermod"
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

	bot.RegisterMod(pingmod.New("ping"), "ping")
	bot.RegisterMod(loggermod.New("logs"), "logs")
	bot.RegisterMod(utilitymod.New("utility"), "utility")
	bot.RegisterMod(moderationmod.New("moderation"), "moderation")
	bot.RegisterMod(userrolemod.New("userrole"), "userrole")
	bot.RegisterMod(searchmod.New("search"), "search")
	bot.RegisterMod(googlemod.New("google"), "google")

	bot.Run()
	defer bot.Close()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	<-sc
}
