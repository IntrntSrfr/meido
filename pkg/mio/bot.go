package mio

import (
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"math/rand"
	"runtime/debug"
	"sync"
	"time"

	"github.com/intrntsrfr/meido/internal/database"
	"go.uber.org/zap"
)

// Bot is the main bot struct.
type Bot struct {
	sync.Mutex
	Discord   *Discord
	Config    Configurable
	Modules   map[string]Module
	DB        database.DB
	Cooldowns CooldownService
	Callbacks CallbackService
	Log       *zap.Logger
	handlers  map[string][]func(interface{})
}

// NewBot takes in a Config and returns a pointer to a new Bot
func NewBot(config Configurable, db database.DB, log *zap.Logger) *Bot {
	log.Info("new bot")
	rand.Seed(time.Now().Unix())

	return &Bot{
		Discord:   NewDiscord(config.GetString("token")),
		Config:    config,
		Modules:   make(map[string]Module),
		DB:        db,
		Cooldowns: NewCooldownHandler(),
		Callbacks: NewCallbackHandler(),
		Log:       log,
		handlers:  make(map[string][]func(interface{})),
	}
}

// Open will connect to Discord and register event handlers
func (b *Bot) Open(useDefHandlers bool) error {
	b.Log.Info("setting up bot")
	err := b.Discord.Open()
	if err != nil {
		panic(err)
	}
	if useDefHandlers {
		b.Discord.AddEventHandler(ready)
		b.Discord.AddEventHandler(guildJoin)
		b.Discord.AddEventHandler(guildLeave)
		b.Discord.AddEventHandler(memberChunk)
	}
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

func (b *Bot) AddEventHandler(event string, handler func(interface{})) {
	b.Lock()
	defer b.Unlock()
	b.handlers[event] = append(b.handlers[event], handler)
}

func (b *Bot) emit(event string, data interface{}) {
	b.Lock()
	defer b.Unlock()
	if cbs, ok := b.handlers[event]; ok {
		for _, cb := range cbs {
			go cb(data)
		}
	}
}

// listen is the main command handler. It will listen for messages and execute commands accordingly.
func (b *Bot) listen(msg <-chan *DiscordMessage) {
	for {
		m := <-msg
		go b.deliverCallbacks(m)
		go b.processMessage(m)
	}
}

func (b *Bot) processMessage(msg *DiscordMessage) {
	if msg.Message.Author == nil || msg.Message.Author.Bot {
		return
	}
	for _, mod := range b.Modules {
		if msg.IsDM() && !mod.AllowDMs() {
			continue
		}
		if msg.Type()&mod.AllowedTypes() == 0 {
			continue
		}

		// run all passives if they allow the message type
		for _, pas := range mod.Passives() {
			if msg.Type()&pas.AllowedTypes == pas.AllowedTypes {
				go pas.Run(msg)
			}
		}

		// if there is no text, there can be no command
		if msg.LenArgs() <= 0 {
			continue
		}

		if cmd, found := FindCommand(mod, msg.Args()); found {
			b.processCommand(cmd, msg)
		}
	}
}

func (b *Bot) processCommand(cmd *ModuleCommand, msg *DiscordMessage) {
	if !cmd.Enabled {
		return
	}

	if msg.IsDM() && !cmd.AllowDMs {
		return
	}

	if msg.Type()&cmd.AllowedTypes == 0 {
		return
	}

	if cmd.RequiresOwner && !msg.Discord.IsBotOwner(msg) {
		_, _ = msg.Reply("This command is owner only")
		return
	}

	// check if cooldown is for user or channel
	key := fmt.Sprintf("%v:%v", msg.Message.ChannelID, cmd.Name)
	if cmd.CooldownUser {
		key = fmt.Sprintf("%v:%v", msg.Message.Author.ID, cmd.Name)
	}
	if t, ok := b.Cooldowns.Check(key); ok {
		// if on cooldown, we know it's for this command, so we can break out and go next
		_, _ = msg.ReplyAndDelete(fmt.Sprintf("This command is on cooldown for another %v", t), time.Second*2)
		return
	}

	//check for perms
	if cmd.RequiredPerms != 0 {
		if allow, err := msg.HasPermissions(cmd.RequiredPerms); err != nil || !allow {
			return
		}
		if cmd.CheckBotPerms {
			if botAllow, err := msg.Discord.HasPermissions(msg.Message.ChannelID, cmd.RequiredPerms); err != nil || !botAllow {
				return
			}
		}
	}

	// set cmd on cooldown
	b.Cooldowns.Set(key, time.Duration(cmd.Cooldown))
	// run cmd
	b.runCommand(cmd, msg)
}

// if a command causes panic, this will surely keep everything from crashing
func (b *Bot) runCommand(cmd *ModuleCommand, msg *DiscordMessage) {
	defer func() {
		if r := recover(); r != nil {
			d, err := json.MarshalIndent(msg, "", "\t")
			if err != nil {
				return
			}
			b.Log.Error("recovery needed", zap.String("message JSON", string(d)))
			b.emit("command_panicked", &CommandPanicked{cmd, msg, string(debug.Stack())})
			_, _ = msg.Reply("Something terrible happened. Please try again. If that does not work, send a DM to bot dev(s)")
		}
	}()

	cmd.Run(msg)
	b.emit("command_ran", &CommandRan{cmd, msg})
	b.Log.Info("new command", zap.String("author", fmt.Sprintf("%v | %v", msg.Author(), msg.AuthorID())), zap.String("content", msg.RawContent()))
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

func ready(s *discordgo.Session, r *discordgo.Ready) {
	fmt.Println("shard:", s.ShardID)
	fmt.Println("user:", r.User.String())
	fmt.Println("servers:", len(r.Guilds))
}

func guildJoin(s *discordgo.Session, g *discordgo.GuildCreate) {
	_ = s.RequestGuildMembers(g.ID, "", 0, false)
	fmt.Println("loading: ", g.Guild.Name, g.MemberCount, len(g.Members))
}

func guildLeave(s *discordgo.Session, g *discordgo.GuildDelete) {
	if !g.Unavailable {
		return
	}
	fmt.Println(fmt.Sprintf("Removed from guild (%v)", g.ID))
}

func memberChunk(s *discordgo.Session, g *discordgo.GuildMembersChunk) {
	if g.ChunkIndex == g.ChunkCount-1 {
		// I don't know if this will work with several shards
		guild, err := s.Guild(g.GuildID)
		if err != nil {
			return
		}
		fmt.Println("finished loading " + guild.Name)
	}
}
