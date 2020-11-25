package meidov2

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/owo"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"strings"
	"sync"
	"time"
)

type Config struct {
	Token            string   `json:"token"`
	OwoToken         string   `json:"owo_token"`
	ConnectionString string   `json:"connection_string"`
	DmLogChannels    []string `json:"dm_log_channels"`
	OwnerIds         []string `json:"owner_ids"`
	YouTubeKey       string   `json:"youtube_key"`
}

type Bot struct {
	Discord    *Discord
	Config     *Config
	Mods       map[string]Mod
	CommandLog chan *DiscordMessage
	DB         *sqlx.DB
	Owo        *owo.Client
	Cooldowns  CooldownCache
}

type CooldownCache struct {
	sync.Mutex
	m map[string]time.Time
}

func NewBot(config *Config) *Bot {
	d := NewDiscord(config.Token)

	fmt.Println("new bot")

	return &Bot{
		Discord:    d,
		Config:     config,
		Mods:       make(map[string]Mod),
		CommandLog: make(chan *DiscordMessage, 256),
		Cooldowns:  CooldownCache{m: make(map[string]time.Time)},
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
	b.DB = psql
	fmt.Println("psql connection established")

	b.Owo = owo.NewClient(b.Config.OwoToken)
	fmt.Println("owo client created")

	go b.listen(msgChan)
	go b.logCommands()
	return nil
}

func (b *Bot) Run() error {
	return b.Discord.Run()
}

func (b *Bot) Close() {
	b.Discord.Close()
}

func (b *Bot) RegisterMod(mod Mod) {
	fmt.Println(fmt.Sprintf("registering mod '%s'", mod.Name()))
	err := mod.Hook(b)
	if err != nil {
		panic(err)
	}
	b.Mods[mod.Name()] = mod
}

func (b *Bot) listen(msg <-chan *DiscordMessage) {
	for {
		select {
		case m := <-msg:
			if m.Message.Author == nil || m.Message.Author.Bot {
				continue
			}

			for _, mod := range b.Mods {
				if m.IsDM() && !mod.AllowDMs() {
					continue
				}
				if m.Type&mod.AllowedTypes() == 0 {
					continue
				}

				for _, pas := range mod.Passives() {
					if m.Type&pas.AllowedTypes == 0 {
						continue
					}
					go pas.Run(m)
				}

				if m.LenArgs() <= 0 {
					continue
				}

				for _, cmd := range mod.Commands() {

					go func(cmd *ModCommand) {
						if m.IsDM() && !cmd.AllowDMs {
							return
						}
						if m.Type&cmd.AllowedTypes == 0 {
							return
						}

						runCmd := false
						for _, trig := range cmd.Triggers {
							if m.Args()[0] == trig {
								runCmd = true
							}
						}
						if !runCmd {
							return
						}

						if !cmd.Enabled {
							return
						}

						// add some check later to see if command isnt disabled for user

						// check if command for channel is on cooldown
						key := fmt.Sprintf("%v:%v", m.Message.ChannelID, cmd.Name)
						if t, ok := b.isOnCooldown(key); ok {
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

							return
						}

						//check for perms
						if cmd.RequiredPerms != 0 {
							uPerms, err := m.Discord.UserChannelPermissions(m.Member, m.Message.ChannelID)
							if err != nil {
								return
							}
							if uPerms&cmd.RequiredPerms == 0 || uPerms&discordgo.PermissionAdministrator == 0 {
								return
							}

							botPerms, err := m.Discord.Sess.State.UserChannelPermissions(m.Sess.State.User.ID, m.Message.ChannelID)
							if err != nil {
								return
							}
							if botPerms&discordgo.PermissionBanMembers == 0 && botPerms&discordgo.PermissionAdministrator == 0 {
								return
							}
						}

						go cmd.Run(m)

						// set cmd on cooldown
						go b.setOnCooldown(key, time.Duration(cmd.Cooldown))
					}(cmd)

				}
			}
		}
	}
}

func (b *Bot) logCommands() {
	for {
		select {
		case m := <-b.CommandLog:

			b.DB.Exec("INSERT INTO commandlog VALUES(DEFAULT, $1, $2, $3, $4, $5, $6, $7);",
				m.Args()[0], strings.Join(m.Args(), " "), m.Message.Author.ID, m.Message.GuildID,
				m.Message.ChannelID, m.Message.ID, time.Now())

			fmt.Println(m.Shard, m.Message.Author.String(), m.Message.Content, m.TimeReceived.String())
		}
	}
}

// returns if its on cooldown, and the time for the cooldown if any
func (b *Bot) isOnCooldown(key string) (time.Time, bool) {
	b.Cooldowns.Lock()
	defer b.Cooldowns.Unlock()
	t, ok := b.Cooldowns.m[key]
	return t, ok
}

func (b *Bot) setOnCooldown(key string, dur time.Duration) {

	b.Cooldowns.Lock()
	b.Cooldowns.m[key] = time.Now().Add(time.Second * dur)
	b.Cooldowns.Unlock()

	go func() {
		time.AfterFunc(time.Second*dur, func() {
			b.Cooldowns.Lock()
			delete(b.Cooldowns.m, key)
			b.Cooldowns.Unlock()
		})
	}()
}
