package mio

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

type Bot struct {
	sync.Mutex
	Discord   *Discord
	Config    Configurable
	Modules   map[string]Module
	Cooldowns CooldownService
	Callbacks CallbackService
	Log       *zap.Logger
	eventCh   chan *BotEventData
}

type BotEvent string

const (
	BotEventCommandRan      BotEvent = "command_ran"
	BotEventCommandPanicked BotEvent = "command_panicked"
)

type BotEventData struct {
	Type BotEvent
	Data interface{}
}

type CommandRan struct {
	Command *ModuleCommand
	Message *DiscordMessage
}

type CommandPanicked struct {
	Command    *ModuleCommand
	Message    *DiscordMessage
	StackTrace string
}

func NewBot(config Configurable, log *zap.Logger) *Bot {
	log.Info("new bot")
	return &Bot{
		Discord:   NewDiscord(config.GetString("token"), config.GetInt("shards"), log),
		Config:    config,
		Modules:   make(map[string]Module),
		Cooldowns: NewCooldownHandler(),
		Callbacks: NewCallbackHandler(),
		Log:       log,
		eventCh:   make(chan *BotEventData),
	}
}

func (b *Bot) Open(useDefHandlers bool) error {
	b.Log.Info("setting up bot")
	err := b.Discord.Open()
	if err != nil {
		return err
	}
	if useDefHandlers {
		b.Discord.AddEventHandler(readyHandler(b))
		b.Discord.AddEventHandler(guildJoinHandler(b))
		b.Discord.AddEventHandler(guildLeaveHandler(b))
		b.Discord.AddEventHandler(memberChunkHandler(b))
	}
	return nil
}

func (b *Bot) Run(ctx context.Context) error {
	b.Log.Info("starting bot")
	go b.listenMessages(ctx)
	return b.Discord.Run()
}

func (b *Bot) Close() {
	b.Log.Info("stopping bot")
	b.Discord.Close()
}

func (b *Bot) RegisterModule(mod Module) {
	b.Log.Info("adding module", zap.String("name", mod.Name()))
	err := mod.Hook()
	if err != nil {
		b.Log.Error("could not register module", zap.Error(err))
		return
	}
	b.Modules[mod.Name()] = mod
}

func (b *Bot) emit(event BotEvent, data interface{}) {
	b.eventCh <- &BotEventData{Type: event, Data: data}
}

func (b *Bot) EventChannel() chan *BotEventData {
	return b.eventCh
}

func (b *Bot) listenMessages(ctx context.Context) {
	for {
		select {
		case msg := <-b.Discord.messageChan:
			go b.deliverCallbacks(msg)
			go b.processMessage(msg)
		case <-ctx.Done():
			return
		}
	}
}

func (b *Bot) processMessage(msg *DiscordMessage) {
	for _, mod := range b.Modules {
		if !mod.AllowsMessage(msg) {
			return
		}

		for _, pas := range mod.Passives() {
			if !pas.Enabled || msg.Type()&pas.AllowedTypes == 0 {
				continue
			}
			go pas.Run(msg)
		}

		if len(msg.Args()) <= 0 {
			continue
		}

		if cmd, err := mod.FindCommandByTriggers(msg.RawContent()); err == nil {
			b.processCommand(cmd, msg)
		}
	}
}

func (b *Bot) processCommand(cmd *ModuleCommand, msg *DiscordMessage) {
	if !cmd.IsEnabled || !cmd.AllowsMessage(msg) {
		return
	}

	if cmd.RequiresUserType == UserTypeBotOwner && !b.IsOwner(msg.AuthorID()) {
		_, _ = msg.Reply("This command is owner only")
		return
	}

	cdKey := cmd.CooldownKey(msg)
	if t, ok := b.Cooldowns.Check(cdKey); ok {
		_, _ = msg.ReplyAndDelete(fmt.Sprintf("This command is on cooldown for another %v", t), time.Second*2)
		return
	}
	b.Cooldowns.Set(cdKey, time.Duration(cmd.Cooldown))
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
	if msg.Type() != MessageTypeCreate {
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
