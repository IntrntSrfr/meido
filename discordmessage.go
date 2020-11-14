package meidov2

import (
	"github.com/bwmarrin/discordgo"
	"sort"
	"strings"
	"time"
)

type MessageType int

const (
	MessageTypeCreate MessageType = iota
	MessageTypeUpdate
	MessageTypeDelete
)

type DiscordMessage struct {
	Sess    *discordgo.Session
	Discord *Discord
	Message *discordgo.Message

	// Partial guild member, use only for guild related stuff
	Author       *discordgo.Member
	Type         MessageType
	TimeReceived time.Time
	Shard        int
}

func (m *DiscordMessage) Reply(data string) (*discordgo.Message, error) {
	return m.Sess.ChannelMessageSend(m.Message.ChannelID, data)
}

func (m *DiscordMessage) ReplyEmbed(embed *discordgo.MessageEmbed) (*discordgo.Message, error) {
	return m.Sess.ChannelMessageSendEmbed(m.Message.ChannelID, embed)
}

func (m *DiscordMessage) Args() []string {
	return strings.Fields(strings.ToLower(m.Message.Content))
}

func (m *DiscordMessage) LenArgs() int {
	return len(m.Args())
}

func (m *DiscordMessage) IsDM() bool {
	return m.Message.Type == discordgo.MessageTypeDefault && m.Message.GuildID == ""
}

func (m *DiscordMessage) HighestRole(gid, uid string) int {

	g, err := m.Sess.State.Guild(gid)
	if err != nil {
		return -1
	}
	mem, err := m.Sess.GuildMember(gid, uid)
	if err != nil {
		return -1
	}

	gRoles := g.Roles

	sort.Sort(RoleByPos(gRoles))

	for _, gr := range gRoles {
		for _, r := range mem.Roles {
			if r == gr.ID {
				return gr.Position
			}
		}
	}

	return -1
}

func (m *DiscordMessage) HighestColor(gid, uid string) int {

	g, err := m.Sess.State.Guild(gid)
	if err != nil {
		return 0
	}

	mem, err := m.Sess.GuildMember(gid, uid)
	if err != nil {
		return 0
	}

	gRoles := g.Roles

	sort.Sort(RoleByPos(gRoles))

	for _, gr := range gRoles {
		for _, r := range mem.Roles {
			if r == gr.ID {
				if gr.Color != 0 {
					return gr.Color
				}
			}
		}
	}

	return 0
}

type RoleByPos []*discordgo.Role

func (a RoleByPos) Len() int           { return len(a) }
func (a RoleByPos) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a RoleByPos) Less(i, j int) bool { return a[i].Position > a[j].Position }
