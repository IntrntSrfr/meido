package main

import (
	"encoding/json"
	"github.com/intrntsrfr/meidov2"
	"github.com/intrntsrfr/meidov2/mods/pingmod"
	"io/ioutil"
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

	bot.Run()
	defer bot.Close()

	lol := make(chan interface{})
	<-lol

}
