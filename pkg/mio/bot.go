package mio

import (
	"encoding/json"
	"fmt"
	"github.com/intrntsrfr/meido/internal/database"
	"github.com/intrntsrfr/meido/internal/structs"
	"math/rand"
	"strings"
	"time"

	"go.uber.org/zap"
)

// Config is the config struct for the bot
type Config struct {
	Token             string   `json:"token"`
	ConnectionString  string   `json:"connection_string"`
	OwnerIds          []string `json:"owner_ids"`
	DmLogChannels     []string `json:"dm_log_channels"`
	OwoToken          string   `json:"owo_token"`
	YouTubeToken      string   `json:"youtube_key"`
	OpenWeatherApiKey string   `json:"open_weather_api_key"`
}

// Bot is the main bot struct.
type Bot struct {
	Discord   *Discord
	Config    *Config
	Modules   map[string]Module
	DB        database.DB
	Cooldowns CooldownService
	Callbacks CallbackService
	Log       *zap.Logger
}

// NewBot takes in a Config and returns a pointer to a new Bot
func NewBot(config *Config, db database.DB, log *zap.Logger) *Bot {
	log.Info("new bot")
	rand.Seed(time.Now().Unix())

	return &Bot{
		Discord:   NewDiscord(config.Token),
		Config:    config,
		Modules:   make(map[string]Module),
		DB:        db,
		Cooldowns: NewCooldownHandler(),
		Callbacks: NewCallbackHandler(),
		Log:       log,
	}
}

// Open will connect to Discord and register eventhandlers
func (b *Bot) Open() error {
	b.Log.Info("setting up bot")
	err := b.Discord.Open()
	if err != nil {
		panic(err)
	}

	registerEvents(b.Discord)
	return nil
}

// Run will start the sessions against Discord and runs it.
func (b *Bot) Run() error {
	b.Log.Info("starting bot")
	go b.listen(b.Discord.messageChan)
	return b.Discord.Run()
}

// Close saves all mod states and closes the bot sessions.
func (b *Bot) Close() {
	b.Log.Info("stopping bot")
	b.Discord.Close()
}

// RegisterModule takes in a Module and registers it.
func (b *Bot) RegisterModule(mod Module) {
	b.Log.Info("adding module", zap.String("name", mod.Name()))
	err := mod.Hook()
	if err != nil {
		panic(err)
	}
	b.Modules[mod.Name()] = mod
}

// listen is the main command handler. It will listen for messages and execute commands accordingly.
func (b *Bot) listen(msg <-chan *DiscordMessage) {
	for {
		m := <-msg
		go b.deliverCallbacks(m)
		go b.processMessage(m)
	}
}

func (b *Bot) processMessage(m *DiscordMessage) {
	if m.Message.Author == nil || m.Message.Author.Bot {
		return
	}
	for _, mod := range b.Modules {
		if m.IsDM() && !mod.AllowDMs() {
			continue
		}
		if m.Type()&mod.AllowedTypes() == 0 {
			continue
		}

		// run all passives if they allow the message type
		for _, pas := range mod.Passives() {
			if m.Type()&pas.AllowedTypes == pas.AllowedTypes {
				go pas.Run(m)
			}
		}

		// if there is no text, there can be no command
		if m.LenArgs() <= 0 {
			continue
		}

		if cmd, found := FindCommand(mod, m.Args()); found {
			b.processCommand(cmd, m)
		}
	}
}

func (b *Bot) processCommand(cmd *ModuleCommand, m *DiscordMessage) {
	if !cmd.Enabled {
		return
	}

	if m.IsDM() && !cmd.AllowDMs {
		return
	}

	if m.Type()&cmd.AllowedTypes == 0 {
		return
	}

	if cmd.RequiresOwner && !m.Discord.IsBotOwner(m) {
		_, _ = m.Reply("This command is owner only")
		return
	}

	// check if cooldown is for user or channel
	key := fmt.Sprintf("%v:%v", m.Message.ChannelID, cmd.Name)
	if cmd.CooldownUser {
		key = fmt.Sprintf("%v:%v", m.Message.Author.ID, cmd.Name)
	}
	if t, ok := b.Cooldowns.Check(key); ok {
		// if on cooldown, we know it's for this command, so we can break out and go next
		_, _ = m.ReplyAndDelete(fmt.Sprintf("This command is on cooldown for another %v", t), time.Second*2)
		return
	}

	//check for perms
	if cmd.RequiredPerms != 0 {
		if allow, err := m.HasPermissions(cmd.RequiredPerms); err != nil || !allow {
			return
		}
		if cmd.CheckBotPerms {
			if botAllow, err := m.Discord.HasPermissions(m.Message.ChannelID, cmd.RequiredPerms); err != nil || !botAllow {
				return
			}
		}
	}

	// log cmd
	b.logCommand(m, cmd)
	// set cmd on cooldown
	b.Cooldowns.Set(key, time.Duration(cmd.Cooldown))
	// run cmd
	b.runCommand(cmd, m)
}

// if a command causes panic, this will surely keep everything from crashing
func (b *Bot) runCommand(cmd *ModuleCommand, m *DiscordMessage) {
	defer func() {
		if r := recover(); r != nil {
			d, err := json.MarshalIndent(m, "", "\t")
			if err != nil {
				return
			}

			b.Log.Error("recovery needed", zap.String("message JSON", string(d)))

			//log.Println(string(d))
			//now := time.Now()
			//fmt.Println(fmt.Sprintf("!!! RECOVERY NEEDED !!!\ntime: %v\nreason: %v\n\n\n", now.String(), r))

			_, _ = m.Reply("Something terrible happened. Please try again. If that does not work, send a DM to bot dev(s)")
		}
	}()

	cmd.Run(m)
}

func (b *Bot) deliverCallbacks(msg *DiscordMessage) {
	if msg.Type()&MessageTypeCreate == 0 {
		return
	}

	key := fmt.Sprintf("%v:%v", msg.ChannelID(), msg.AuthorID())
	ch, err := b.Callbacks.Get(key)
	if err != nil {
		return
	}
	ch <- msg
}

// logCommand logs an executed command
func (b *Bot) logCommand(msg *DiscordMessage, cmd *ModuleCommand) {
	b.Log.Info("new command", zap.String("author", fmt.Sprintf("%v | %v", msg.Author(), msg.AuthorID())), zap.String("content", msg.RawContent()))
	// fmt.Println(msg.Shard, msg.Message.Author, msg.Message.Content, msg.TimeReceived.String())

	if err := b.DB.CreateCommandLogEntry(&structs.CommandLogEntry{
		Command:   cmd.Name,
		Args:      strings.Join(msg.Args(), " "),
		UserID:    msg.AuthorID(),
		GuildID:   msg.GuildID(),
		ChannelID: msg.ChannelID(),
		MessageID: msg.Message.ID,
		SentAt:    time.Now(),
	}); err != nil {
		b.Log.Error("error logging command", zap.Error(err))
	}
}
