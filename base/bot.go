package base

import (
	"encoding/json"
	"fmt"
	"github.com/intrntsrfr/meido/database"
	"go.uber.org/zap"
	"log"
	"math/rand"
	"strings"
	"time"
)

// Config is the config struct for the bot.
type Config struct {
	Token            string   `json:"token"`
	ConnectionString string   `json:"connection_string"`
	OwnerIds         []string `json:"owner_ids"`
	DmLogChannels    []string `json:"dm_log_channels"`
	OwoToken         string   `json:"owo_token"`
	YouTubeToken     string   `json:"youtube_key"`
}

// Bot is the main bot struct.
type Bot struct {
	Discord   *Discord
	Config    *Config
	Mods      map[string]Mod
	DB        *database.DB
	Cooldowns CooldownService
	Callbacks CallbackService
	Log       *zap.Logger
}

// NewBot takes in a Config and returns a pointer to a new Bot
func NewBot(config *Config, db *database.DB, log *zap.Logger) *Bot {
	rand.Seed(time.Now().Unix())
	log.Info("new bot")
	return &Bot{
		Discord:   NewDiscord(config.Token),
		Config:    config,
		Mods:      make(map[string]Mod),
		DB:        db,
		Cooldowns: NewCooldownHandler(),
		Callbacks: NewCallbackHandler(),
		Log:       log,
	}
}

// Open sets up the required things the bot needs to run.
// establishes a PSQL connection and starts listening for commands.
func (b *Bot) Open() error {
	log.Println("open and run")
	msgChan, err := b.Discord.Open()
	if err != nil {
		panic(err)
	}

	registerEvents(b.Discord)

	go b.listen(msgChan)
	return nil
}

// Run will start the sessions against Discord and runs it.
func (b *Bot) Run() error {
	return b.Discord.Run()
}

// Close saves all mod states and closes the bot sessions.
func (b *Bot) Close() {
	b.Discord.Close()
}

// RegisterMod takes in a Mod and registers it.
func (b *Bot) RegisterMod(mod Mod) {
	log.Println(fmt.Sprintf("registering mod '%s'", mod.Name()))
	err := mod.Hook()
	if err != nil {
		panic(err)
	}
	b.Mods[mod.Name()] = mod
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

	for _, mod := range b.Mods {
		if m.IsDM() && !mod.AllowDMs() {
			continue
		}
		if m.Type()&mod.AllowedTypes() == 0 {
			continue
		}

		// check if user can use the folder or not
		// may be based on user level, roles, channel etc..

		// it should also be an option to disable passives for certain channels, server or user basis
		// mostly server, channel and maybe role though.
		// probably just server and channels

		for _, pas := range mod.Passives() {
			if m.Type()&pas.AllowedTypes == 0 {
				continue
			}
			go pas.Run(m)
		}

		if m.LenArgs() <= 0 {
			continue
		}

		cmd, found := FindCommand(mod, m.Args())
		if !found {
			continue
		}

		b.processCommand(cmd, m)
	}
}

func (b *Bot) processCommand(cmd *ModCommand, m *DiscordMessage) {
	if !cmd.Enabled {
		return
	}

	if m.IsDM() && !cmd.AllowDMs {
		return
	}

	if m.Type()&cmd.AllowedTypes == 0 {
		return
	}

	if cmd.RequiresOwner && !m.Discord.IsOwner(m) {
		_, _ = m.Reply("owner only lol")
		return
	}

	// check if user can use command or not
	// may be based on user level, roles, channel etc..
	/*
		if m.GuildID() != "" && !b.Perms.Allow(cmd.Name, m.GuildID(), m.ChannelID(), m.Author().ID, m.Member().Roles) {
			return
		}
	*/

	// check if command for channel is on cooldown
	key := ""
	if cmd.CooldownUser {
		key = fmt.Sprintf("%v:%v", m.Message.Author.ID, cmd.Name)
	} else {
		key = fmt.Sprintf("%v:%v", m.Message.ChannelID, cmd.Name)
	}
	if t, ok := b.Cooldowns.Check(key); ok {
		// if on cooldown, we know it's for this command, so we can break out and go next
		cdMsg, err := m.Reply(fmt.Sprintf("on cooldown for another %v", t))
		if err != nil {
			log.Println(err)
			return
		}
		go func() {
			time.AfterFunc(time.Second*2, func() {
				_ = m.Sess.ChannelMessageDelete(cdMsg.ChannelID, cdMsg.ID)
			})
		}()
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

	// run cmd
	go runCommand(cmd.Run, m)
	// log cmd
	go b.logCommand(m, cmd)
	// set cmd on cooldown
	go b.Cooldowns.Set(key, time.Duration(cmd.Cooldown))
}

// if a command causes panic, this will surely keep everything from crashing
func runCommand(f func(*DiscordMessage), m *DiscordMessage) {
	defer func() {
		if r := recover(); r != nil {
			d, err := json.MarshalIndent(m, "", "\t")
			if err != nil {
				return
			}

			log.Println(string(d))
			log.Println()
			log.Println()

			now := time.Now()

			fmt.Println(fmt.Sprintf("!!! RECOVERY NEEDED !!!\ntime: %v\nreason: %v\n\n\n", now.String(), r))

			_, _ = m.Reply("Something terrible happened. Please try again. If that does not work, send a DM to bot dev(s)")
		}
	}()

	f(m)
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
func (b *Bot) logCommand(msg *DiscordMessage, cmd *ModCommand) {
	b.DB.Exec("INSERT INTO command_log VALUES(DEFAULT, $1, $2, $3, $4, $5, $6, $7);",
		cmd.Name, strings.Join(msg.Args(), " "), msg.Message.Author.ID, msg.Message.GuildID,
		msg.Message.ChannelID, msg.Message.ID, time.Now())

	fmt.Println(msg.Shard, msg.Message.Author.String(), msg.Message.Content, msg.TimeReceived.String())
}
