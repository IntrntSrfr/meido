package base

import (
	"errors"
	"fmt"
	"github.com/intrntsrfr/meido/internal/services/cooldowns"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/intrntsrfr/owo"
	"github.com/jmoiron/sqlx"
)

// Config is the config struct for the bot.
type Config struct {
	Token            string   `json:"token"`
	OwoToken         string   `json:"owo_token"`
	ConnectionString string   `json:"connection_string"`
	DmLogChannels    []string `json:"dm_log_channels"`
	OwnerIds         []string `json:"owner_ids"`
	YouTubeKey       string   `json:"youtube_key"`
}

// Bot is the main bot struct.
type Bot struct {
	Discord   *Discord
	Config    *Config
	Mods      map[string]Mod
	DB        *sqlx.DB
	Owo       *owo.Client
	Cooldowns *cooldowns.CooldownHandler
	Callbacks *CallbackCache
	Perms     *PermissionHandler
}

type CallbackCache struct {
	sync.Mutex
	ch map[string]chan *DiscordMessage
}

// NewBot takes in a Config and returns a pointer to a new Bot
func NewBot(config *Config) *Bot {
	d := NewDiscord(config.Token)

	fmt.Println("new bot")

	if _, err := os.Stat("./data"); err != nil {
		if err := os.Mkdir("./data", os.ModePerm); err != nil {
			panic(err)
		}
	}

	return &Bot{
		Discord:   d,
		Config:    config,
		Mods:      make(map[string]Mod),
		Cooldowns: cooldowns.NewCooldownHandler(),
		Callbacks: &CallbackCache{ch: make(map[string]chan *DiscordMessage)},
		Perms:     NewPermissionHandler(),
	}
}

// Open sets up the required things the bot needs to run.
// establishes a PSQL connection and starts listening for commands.
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
	b.DB = psql
	fmt.Println("psql connection established")

	b.Owo = owo.NewClient(b.Config.OwoToken)
	fmt.Println("owo client created")

	go b.listen(msgChan)
	return nil
}

// Run will start the sessions against Discord and runs it.
func (b *Bot) Run() error {
	return b.Discord.Run()
}

// Close saves all mod states and closes the bot sessions.
func (b *Bot) Close() {

	for _, mod := range b.Mods {
		err := mod.Save()
		if err != nil {
			fmt.Println(fmt.Sprintf("Error saving %v: %v", mod.Name(), err))
		}
	}

	b.Discord.Close()
}

// RegisterMod takes in a Mod and registers it.
func (b *Bot) RegisterMod(mod Mod) {
	fmt.Println(fmt.Sprintf("registering mod '%s'", mod.Name()))
	err := mod.Hook(b)
	if err != nil {
		panic(err)
	}
	b.Mods[mod.Name()] = mod
}

// MakeCallback returns a channel for future messages and stores it using channel and user id
func (b *Bot) MakeCallback(channelID, userID string) (chan *DiscordMessage, error) {
	key := fmt.Sprintf("%v:%v", channelID, userID)

	ch := make(chan *DiscordMessage)
	b.Callbacks.Lock()
	defer b.Callbacks.Unlock()
	if _, ok := b.Callbacks.ch[key]; ok {
		return nil, errors.New("a menu already exists")
	}

	b.Callbacks.ch[key] = ch
	return ch, nil
}

// CloseCallback closes a callback
func (b *Bot) CloseCallback(channelID, userID string) {
	key := fmt.Sprintf("%v:%v", channelID, userID)

	b.Callbacks.Lock()
	defer b.Callbacks.Unlock()
	close(b.Callbacks.ch[key])
	delete(b.Callbacks.ch, key)
}

// listen is the main command handler. It will listen for messages and execute commands accordingly.
func (b *Bot) listen(msg <-chan *DiscordMessage) {
	for {
		m := <-msg
		go b.processMessage(m)
		go b.deliverCallbacks(m)
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
		if m.Type&mod.AllowedTypes() == 0 {
			continue
		}

		// check if user can use the folder or not
		// may be based on user level, roles, channel etc..

		// it should also be an option to disable passives for certain channels, server or user basis
		// mostly server, channel and maybe role though.
		// probably just server and channels

		for _, pas := range mod.Passives() {
			if m.Type&pas.AllowedTypes == 0 {
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

	if m.Type&cmd.AllowedTypes == 0 {
		return
	}

	if cmd.RequiresOwner && !m.Discord.IsOwner(m) {
		m.Reply("owner only lol")
		return
	}

	// check if user can use command or not
	// may be based on user level, roles, channel etc..

	if m.GuildID() != "" && !b.Perms.Allow(cmd.Name, m.GuildID(), m.ChannelID(), m.Author().ID, m.Member().Roles) {
		return
	}

	// check if command for channel is on cooldown
	key := ""
	if cmd.CooldownUser {
		key = fmt.Sprintf("%v:%v", m.Message.Author.ID, cmd.Name)
	} else {
		key = fmt.Sprintf("%v:%v", m.Message.ChannelID, cmd.Name)
	}
	if t, ok := b.Cooldowns.IsOnCooldown(key); ok {
		// if on cooldown, we know its for this command so we can break out and go next
		cdMsg, err := m.Reply(fmt.Sprintf("on cooldown for another %v", time.Until(t)))
		if err != nil {
			return
		}
		go func() {
			time.AfterFunc(time.Second*2, func() {
				m.Sess.ChannelMessageDelete(cdMsg.ChannelID, cdMsg.ID)
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

	//go cmd.Run(m)
	// log cmd
	go b.logCommand(m, cmd)
	// set cmd on cooldown
	go b.Cooldowns.SetOnCooldown(key, time.Duration(cmd.Cooldown))
}

func runCommand(f func(*DiscordMessage), m *DiscordMessage) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r)
		}
	}()

	f(m)
}

func (b *Bot) deliverCallbacks(msg *DiscordMessage) {
	if msg.Type&MessageTypeCreate == 0 {
		return
	}

	key := fmt.Sprintf("%v:%v", msg.ChannelID(), msg.Author().ID)

	b.Callbacks.Lock()
	defer b.Callbacks.Unlock()
	ch, ok := b.Callbacks.ch[key]
	if !ok {
		return
	}

	ch <- msg
}

// logCommand logs an executed command
func (b *Bot) logCommand(msg *DiscordMessage, cmd *ModCommand) {
	b.DB.Exec("INSERT INTO commandlog VALUES(DEFAULT, $1, $2, $3, $4, $5, $6, $7);",
		cmd.Name, strings.Join(msg.Args(), " "), msg.Message.Author.ID, msg.Message.GuildID,
		msg.Message.ChannelID, msg.Message.ID, time.Now())

	fmt.Println(msg.Shard, msg.Message.Author.String(), msg.Message.Content, msg.TimeReceived.String())
}
