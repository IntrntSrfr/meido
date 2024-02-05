package mio

import (
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

func readyHandler(b *Bot) func(s *discordgo.Session, r *discordgo.Ready) {
	return func(s *discordgo.Session, r *discordgo.Ready) {
		b.Logger.Info("Event: ready",
			zap.Int("shard", s.ShardID),
			zap.String("user", r.User.String()),
			zap.Int("server count", len(r.Guilds)),
		)
	}
}

func guildJoinHandler(b *Bot) func(s *discordgo.Session, g *discordgo.GuildCreate) {
	return func(s *discordgo.Session, g *discordgo.GuildCreate) {
		_ = s.RequestGuildMembers(g.ID, "", 0, "", false)
		b.Logger.Info("Event: guild join",
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
		b.Logger.Info("Event: guild leave",
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
			b.Logger.Info("Event: guild members chunk",
				zap.String("name", guild.Name),
			)
		}
	}
}
