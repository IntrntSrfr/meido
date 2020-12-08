package meidov2

import (
	"fmt"
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
	CommandLog chan *ExecutedCommand
	DB         *sqlx.DB
	Owo        *owo.Client
	Cooldowns  CooldownCache
}

type CooldownCache struct {
	sync.Mutex
	m map[string]time.Time
}

type PermissionCache struct {
	sync.Mutex
	m map[string][]*PermissionSetting
}
type PermissionSetting struct {
}

type ExecutedCommand struct {
	Msg *DiscordMessage
	Cmd *ModCommand
}

func NewBot(config *Config) *Bot {
	d := NewDiscord(config.Token)

	fmt.Println("new bot")

	return &Bot{
		Discord:    d,
		Config:     config,
		Mods:       make(map[string]Mod),
		CommandLog: make(chan *ExecutedCommand, 256),
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
							splitTrig := strings.Split(trig, " ")

							if m.LenArgs() < len(splitTrig) {
								break
							}
							if strings.Join(m.Args()[:len(splitTrig)], " ") == trig {
								runCmd = true
							}
						}

						if !runCmd {
							return
						}

						if !cmd.Enabled {
							return
						}

						if cmd.RequiresOwner {
							if !m.IsOwner() {
								m.Reply("owner only lol")
								return
							}
						}

						// check if user can use command or not
						// may be based on user level, roles, channel etc..

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
							if !m.HasPermissions(m.Member, m.Message.ChannelID, cmd.RequiredPerms) {
								return
							}
							if cmd.CheckBotPerms {
								m.Discord.HasPermissions(m.Message.ChannelID, cmd.RequiredPerms)
							}
						}

						go cmd.Run(m)
						b.CommandLog <- &ExecutedCommand{Msg: m, Cmd: cmd}

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
				m.Cmd.Name, strings.Join(m.Msg.Args(), " "), m.Msg.Message.Author.ID, m.Msg.Message.GuildID,
				m.Msg.Message.ChannelID, m.Msg.Message.ID, time.Now())

			fmt.Println(m.Msg.Shard, m.Msg.Message.Author.String(), m.Msg.Message.Content, m.Msg.TimeReceived.String())
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
