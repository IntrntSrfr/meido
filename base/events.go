package base

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
)

func registerEvents(d *Discord) {
	d.AddEventHandler(ready)
	d.AddEventHandler(guildJoin)
	d.AddEventHandler(guildLeave)
	d.AddEventHandler(memberChunk)
}

func ready(s *discordgo.Session, r *discordgo.Ready) {
	fmt.Println("shard:", s.ShardID)
	fmt.Println("user:", r.User.String())
	fmt.Println("servers:", len(r.Guilds))
}

func guildJoin(s *discordgo.Session, g *discordgo.GuildCreate) {
	s.RequestGuildMembers(g.ID, "", 0, false)
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
