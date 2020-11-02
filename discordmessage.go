package meidov2

import (
	"context"
	"github.com/andersfylling/disgord"
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
	Discord      *Discord
	Message      *disgord.Message
	Type         MessageType
	TimeReceived time.Time
}

func (m *DiscordMessage) Reply(data interface{}) (*disgord.Message, error) {
	return m.Discord.Client.SendMsg(context.Background(), m.Message.ChannelID, data)
}

func (m *DiscordMessage) Args() []string {
	return strings.Fields(m.Message.Content)
}

func (m *DiscordMessage) LenArgs() int {
	return len(m.Args())
}

func (m *DiscordMessage) HighestRole(gid, uid disgord.Snowflake) int {

	mem, err := m.Discord.Client.GetMember(context.Background(), gid, uid)
	if err != nil {
		return -1
	}

	gRoles, err := m.Discord.Client.GetGuildRoles(context.Background(), gid)
	if err != nil {
		return -1
	}

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

func (m *DiscordMessage) HighestColor(gid, uid disgord.Snowflake) int {

	mem, err := m.Discord.Client.GetMember(context.Background(), gid, uid)
	if err != nil {
		return 0
	}

	gRoles, err := m.Discord.Client.GetGuildRoles(context.Background(), gid)
	if err != nil {
		return 0
	}

	sort.Sort(RoleByPos(gRoles))

	for _, gr := range gRoles {
		for _, r := range mem.Roles {
			if r == gr.ID {
				if gr.Color != 0 {
					return int(gr.Color)
				}
			}
		}
	}

	return 0
}

type RoleByPos []*disgord.Role

func (a RoleByPos) Len() int           { return len(a) }
func (a RoleByPos) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a RoleByPos) Less(i, j int) bool { return a[i].Position > a[j].Position }
