package utilitymod

import (
	"context"
	"fmt"
	"github.com/andersfylling/disgord"
	"github.com/intrntsrfr/meidov2"
	"sort"
	"strconv"
)

type UtilityMod struct {
	cl       chan *meidov2.DiscordMessage
	commands []func(msg *meidov2.DiscordMessage)
}

func New() meidov2.Mod {
	return &UtilityMod{
		cl: make(chan *meidov2.DiscordMessage, 256),
	}
}

func (m *UtilityMod) Save() error {
	return nil
}

func (m *UtilityMod) Load() error {
	return nil
}

func (m *UtilityMod) Settings(msg *meidov2.DiscordMessage) {

}

func (m *UtilityMod) Help(msg *meidov2.DiscordMessage) {

}

func (m *UtilityMod) Hook(b *meidov2.Bot, cl chan *meidov2.DiscordMessage) error {
	m.cl = cl

	m.commands = append(m.commands, m.Avatar)

	return nil
}

func (m *UtilityMod) Message(msg *meidov2.DiscordMessage) {
	for _, c := range m.commands {
		go c(msg)
	}
}

func (m *UtilityMod) Avatar(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() == 0 || msg.Args()[0] != ">av" {
		return
	}

	m.cl <- msg

	var targetUser *disgord.User
	var err error

	if msg.LenArgs() > 1 {
		if len(msg.DiscordMessage.Mentions) >= 1 {
			targetUser = msg.DiscordMessage.Mentions[0]
		} else {
			id, err := strconv.Atoi(msg.Args()[1])
			if err != nil {
				return
			}
			targetUser, err = msg.Discord.Client.GetUser(context.Background(), disgord.Snowflake(id))
			if err != nil {
				return
			}
		}
	} else {
		targetUser, err = msg.Discord.Client.GetUser(context.Background(), msg.DiscordMessage.Author.ID)
		if err != nil {
			return
		}
	}

	if targetUser == nil {
		return
	}

	if targetUser.Avatar == "" {
		msg.Discord.Client.SendMsg(context.Background(), msg.DiscordMessage.ChannelID, &disgord.Embed{
			Color:       0xC80000,
			Description: fmt.Sprintf("%v has no avatar set.", targetUser.Tag()),
		})
	} else {
		msg.Discord.Client.SendMsg(context.Background(), msg.DiscordMessage.ChannelID, &disgord.Embed{
			Color: HighestColor(msg.Discord.Client, msg.DiscordMessage.GuildID, targetUser.ID),
			Title: targetUser.Tag(),
			Image: &disgord.EmbedImage{URL: AvatarURL(targetUser, 1024)},
		})
	}
}

func AvatarURL(u *disgord.User, size int) string {
	a, _ := u.AvatarURL(size, true)
	return a
}

func HighestColor(s disgord.Session, gid, uid disgord.Snowflake) int {

	mem, err := s.GetMember(context.Background(), gid, uid)
	if err != nil {
		return 0
	}

	gRoles, err := s.GetGuildRoles(context.Background(), gid)
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
