package mio

import (
	"fmt"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

// Bot is the main bot struct.
type Bot struct {
	sync.Mutex
	Discord   *Discord
	Config    Configurable
	Modules   map[string]Module
	Cooldowns CooldownService
	Callbacks CallbackService
	Log       *zap.Logger
	handlers  map[string][]func(interface{})
}

// NewBot takes in a Config and returns a pointer to a new Bot
func NewBot(config Configurable, log *zap.Logger) *Bot {
	log.Info("new bot")
	return &Bot{
		Discord:   NewDiscord(config.GetString("token"), config.GetInt("shards"), log),
		Config:    config,
		Modules:   make(map[string]Module),
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
		b.Discord.AddEventHandler(readyHandler(b))
		b.Discord.AddEventHandler(guildJoinHandler(b))
		b.Discord.AddEventHandler(guildLeaveHandler(b))
		b.Discord.AddEventHandler(memberChunkHandler(b))
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
		b.Log.Error("could not register module", zap.Error(err))
		return
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
		if len(msg.Args()) <= 0 {
			continue
		}

		if cmd, err := mod.FindCommandByTriggers(msg.RawContent()); err == nil {
			b.processCommand(cmd, msg)
		}
		/*
			if cmd, found := FindCommand(mod, msg.Args()); found {
				b.processCommand(cmd, msg)
			}
		*/
	}
}

func (b *Bot) processCommand(cmd *ModuleCommand, msg *DiscordMessage) {
	if !cmd.IsEnabled {
		return
	}
	if msg.IsDM() && !cmd.AllowDMs {
		return
	}
	if msg.Type()&cmd.AllowedTypes == 0 {
		return
	}
	if cmd.RequiresUserType == UserTypeBotOwner && !b.IsOwner(msg.AuthorID()) {
		_, _ = msg.Reply("This command is owner only")
		return
	}

	var key string
	switch cmd.CooldownScope {
	case User:
		key = fmt.Sprintf("%v:%v", msg.AuthorID(), cmd.Name)
	case Channel:
		key = fmt.Sprintf("%v:%v", msg.ChannelID(), cmd.Name)
	case Guild:
		key = fmt.Sprintf("%v:%v", msg.GuildID(), cmd.Name)
	}

	if t, ok := b.Cooldowns.Check(key); ok {
		// if on cooldown, we know it's for this command, so we can break out and go next
		_, _ = msg.ReplyAndDelete(fmt.Sprintf("This command is on cooldown for another %v", t), time.Second*2)
		return
	}

	//check for perms
	if cmd.RequiredPerms != 0 {
		if allow, err := msg.AuthorHasPermissions(cmd.RequiredPerms); err != nil || !allow {
			return
		}
		if cmd.CheckBotPerms {
			if botAllow, err := msg.Discord.BotHasPermissions(msg.ChannelID(), cmd.RequiredPerms); err != nil || !botAllow {
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
			b.Log.Error("recovery needed", zap.Any("error", r))
			b.emit("command_panicked", &CommandPanicked{cmd, msg, string(debug.Stack())})
			_, _ = msg.Reply("Something terrible happened. Please try again. If that does not work, send a DM to bot dev(s)")
		}
	}()

	cmd.Run(msg)
	b.emit("command_ran", &CommandRan{cmd, msg})
	b.Log.Info("new command",
		zap.String("id", msg.ID()),
		zap.String("content", msg.RawContent()),
		zap.String("author ID", msg.AuthorID()),
		zap.String("author username", msg.Author().String()),
	)
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

func (b *Bot) FindModule(name string) (Module, error) {
	for _, m := range b.Modules {
		if strings.EqualFold(m.Name(), name) {
			return m, nil
		}
	}
	return nil, ErrModuleNotFound
}

func (b *Bot) FindCommand(name string) (*ModuleCommand, error) {
	for _, m := range b.Modules {
		if cmd, err := m.FindCommand(name); err == nil {
			return cmd, nil
		}
	}
	return nil, ErrCommandNotFound
}

func (b *Bot) FindPassive(name string) (*ModulePassive, error) {
	for _, m := range b.Modules {
		if pas, err := m.FindPassive(name); err == nil {
			return pas, nil
		}
	}
	return nil, ErrPassiveNotFound
}

func (b *Bot) IsOwner(userID string) bool {
	for _, id := range b.Config.GetStringSlice("owner_ids") {
		if id == userID {
			return true
		}
	}
	return false
}

func readyHandler(b *Bot) func(s *discordgo.Session, r *discordgo.Ready) {
	return func(s *discordgo.Session, r *discordgo.Ready) {
		b.Log.Info("ready",
			zap.Int("shard", s.ShardID),
			zap.String("user", r.User.String()),
			zap.Int("server count", len(r.Guilds)),
		)
	}
}

func guildJoinHandler(b *Bot) func(s *discordgo.Session, g *discordgo.GuildCreate) {
	return func(s *discordgo.Session, g *discordgo.GuildCreate) {
		_ = s.RequestGuildMembers(g.ID, "", 0, "", false)
		b.Log.Info("started loading guild",
			zap.String("name", g.Guild.Name),
			zap.Int("member count", g.MemberCount),
			zap.Int("members available", len(g.Members)),
		)
	}
}

func guildLeaveHandler(b *Bot) func(s *discordgo.Session, g *discordgo.GuildDelete) {
	return func(s *discordgo.Session, g *discordgo.GuildDelete) {
		if !g.Unavailable {
			return
		}
		b.Log.Info("removed from guild",
			zap.String("id", g.ID),
		)
	}
}

func memberChunkHandler(b *Bot) func(s *discordgo.Session, g *discordgo.GuildMembersChunk) {
	return func(s *discordgo.Session, g *discordgo.GuildMembersChunk) {
		if g.ChunkIndex == g.ChunkCount-1 {
			// I don't know if this will work with several shards
			guild, err := s.Guild(g.GuildID)
			if err != nil {
				return
			}
			b.Log.Info("finished loading guild",
				zap.String("name", guild.Name),
			)
		}
	}
}
