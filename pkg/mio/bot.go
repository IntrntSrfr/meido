package mio

import (
	"context"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/pkg/mio"
	"github.com/intrntsrfr/meido/pkg/mio/discord"
	mutils "github.com/intrntsrfr/meido/pkg/mio/utils"
	"github.com/intrntsrfr/meido/pkg/utils"
)

type Bot struct {
	sync.Mutex
	Discord *discord.Discord
	Config  *utils.Config

	*ModuleManager
	EventHandler *EventHandler
	Callbacks    *mutils.CallbackManager
	Cooldowns    *mutils.CooldownManager
	*mio.EventBus

	Logger mio.Logger
}

func (b *Bot) Run(ctx context.Context) error {
	b.Logger.Info("Starting up...")
	go b.EventHandler.Listen(ctx)
	if err := b.Discord.Run(); err != nil {
		return err
	}
	if err := b.setApplicationCommands(); err != nil {
		return err
	}
	b.Logger.Info("Running")
	return nil
}

func (b *Bot) Close() {
	b.Logger.Info("Shutting down")
	b.Discord.Close()
}

func (b *Bot) setApplicationCommands() error {
	var allCommands []*discordgo.ApplicationCommand
	for _, m := range b.Modules {
		allCommands = append(allCommands, m.ApplicationCommandStructs()...)
	}

	created, err := b.Discord.Sess.ApplicationCommandBulkOverwrite(b.Discord.Sess.State().User.ID, "", allCommands)
	if err != nil {
		b.logger.Error("could not overwrite commands", "error", err)
		return err
	}
	for _, c := range created {
		b.logger.Info("Created/updated command", "name", c.Name, "type", uint8(c.Type))
	}
	return nil
}

func (b *Bot) IsOwner(userID string) bool {
	for _, id := range b.Config.GetStringSlice("owner_ids") {
		if id == userID {
			return true
		}
	}
	return false
}

func readyHandler(logger mio.Logger) func(s *discordgo.Session, r *discordgo.Ready) {
	return func(s *discordgo.Session, r *discordgo.Ready) {
		logger.Info("Event: ready",
			"shard", s.ShardID,
			"user", r.User.String(),
			"server count", len(r.Guilds),
		)
	}
}

func guildJoinHandler(logger mio.Logger) func(s *discordgo.Session, g *discordgo.GuildCreate) {
	return func(s *discordgo.Session, g *discordgo.GuildCreate) {
		_ = s.RequestGuildMembers(g.ID, "", 0, "", false)
		logger.Info("Event: guild join",
			"name", g.Guild.Name,
			"member count", g.MemberCount,
			"members available", len(g.Members),
		)
	}
}

func guildLeaveHandler(logger mio.Logger) func(s *discordgo.Session, g *discordgo.GuildDelete) {
	return func(s *discordgo.Session, g *discordgo.GuildDelete) {
		if !g.Unavailable {
			return
		}
		logger.Info("Event: guild leave",
			"id", g.ID,
		)
	}
}

func memberChunkHandler(logger mio.Logger) func(s *discordgo.Session, g *discordgo.GuildMembersChunk) {
	return func(s *discordgo.Session, g *discordgo.GuildMembersChunk) {
		if g.ChunkIndex == g.ChunkCount-1 {
			// I don't know if this will work with several shards
			guild, err := s.Guild(g.GuildID)
			if err != nil {
				return
			}
			logger.Info("Event: guild members chunk",
				"name", guild.Name,
			)
		}
	}
}
