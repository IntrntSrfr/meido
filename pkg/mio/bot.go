package mio

import (
	"context"
	"sync"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

type Bot struct {
	sync.Mutex
	Discord *Discord
	Config  Configurable

	*ModuleManager
	MessageProcessor *MessageProcessor
	Callbacks        *CallbackManager
	*EventManager

	Log *zap.Logger
}

func NewBot(config Configurable, log *zap.Logger) *Bot {
	log.Info("new bot")

	bot := &Bot{
		Config: config,
		Log:    log,
	}

	bot.EventManager = NewEventManager()
	bot.Callbacks = NewCallbackManager()
	bot.ModuleManager = NewModuleManager(log)
	bot.Discord = NewDiscord(config.GetString("token"), config.GetInt("shards"), log)
	bot.MessageProcessor = NewMessageProcessor(bot, log)

	return bot
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
	go b.MessageProcessor.ListenMessages(ctx)
	return b.Discord.Run()
}

func (b *Bot) Close() {
	b.Log.Info("stopping bot")
	b.Discord.Close()
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
