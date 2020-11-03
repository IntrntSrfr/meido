package meidov2

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"strings"
	"time"
)

type Config struct {
	Token            string `json:"token"`
	OwoToken         string `json:"owo_token"`
	ConnectionString string `json:"connection_string"`
	DmLogChannels    []int  `json:"dm_log_channels"`
	OwnerIds         []int  `json:"owner_ids"`
}

type Bot struct {
	Discord    *Discord
	Config     *Config
	Mods       map[string]Mod
	commandLog chan *DiscordMessage
	db         *sqlx.DB
}

func NewBot(config *Config) *Bot {
	d := NewDiscord(config.Token)

	fmt.Println("new bot")

	return &Bot{
		Discord:    d,
		Config:     config,
		Mods:       make(map[string]Mod),
		commandLog: make(chan *DiscordMessage, 256),
	}
}

func (b *Bot) Open() error {
	msgChan, err := b.Discord.Open()
	if err != nil {
		panic(err)
	}

	fmt.Println("open and run")

	psql, err := sqlx.Connect("postgres", b.Config.ConnectionString)
	if err != nil {
		panic(err)
	}
	b.db = psql
	fmt.Println("psql connection established")

	go b.listen(msgChan)
	go b.logCommands()
	return nil
}

func (b *Bot) Run() error {
	return b.Discord.Run()
}

func (b *Bot) Close() {
	b.Discord.Client.Disconnect()
}

func (b *Bot) RegisterMod(mod Mod, name string) {
	fmt.Println(fmt.Sprintf("registering mod '%s'", name))
	err := mod.Hook(b, b.db, b.commandLog)
	if err != nil {
		fmt.Println("could not attach mod:", name)
		return
	}
	b.Mods[name] = mod
}

func (b *Bot) listen(msg <-chan *DiscordMessage) {
	for {
		select {
		case m := <-msg:
			if m.Message.Author.Bot {
				continue
			}
			for _, mod := range b.Mods {
				go mod.Message(m)
			}
		}
	}
}

func (b *Bot) logCommands() {
	for {
		select {
		case m := <-b.commandLog:
			b.db.Exec("INSERT INTO commandlog VALUES(DEFAULT, $1, $2, $3, $4, $5, $6,$7);",
				m.Args()[0], strings.Join(m.Args(), " "), m.Message.Author.ID, m.Message.GuildID,
				m.Message.ChannelID, m.Message.ID, time.Now())
			fmt.Println(m.Message.Author.String(), m.Message.Content, m.TimeReceived.String())
		}
	}
}
