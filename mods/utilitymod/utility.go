package utilitymod

import (
	"context"
	"fmt"
	"github.com/andersfylling/disgord"
	"github.com/intrntsrfr/meidov2"
	"strconv"
)

type UtilityMod struct {
	cl       chan *meidov2.DiscordMessage
	commands []func(msg *meidov2.DiscordMessage)
}

func New() meidov2.Mod {
	return &UtilityMod{
		//cl: make(chan *meidov2.DiscordMessage),
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

	b.Discord.Client.On(disgord.EvtReady, func(s disgord.Session, r *disgord.Ready) {
		s.UpdateStatus(&disgord.UpdateStatusPayload{
			Game: &disgord.Activity{
				Type: disgord.ActivityTypeGame,
				Name: "BEING REWORKED, WILL WORK AGAIN SOON",
			},
		})
	})

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
		if len(msg.Message.Mentions) >= 1 {
			targetUser = msg.Message.Mentions[0]
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
		targetUser, err = msg.Discord.Client.GetUser(context.Background(), msg.Message.Author.ID)
		if err != nil {
			return
		}
	}

	if targetUser == nil {
		return
	}

	if targetUser.Avatar == "" {
		msg.Reply(&disgord.Embed{
			Color:       0xC80000,
			Description: fmt.Sprintf("%v has no avatar set.", targetUser.Tag()),
		})
	} else {
		msg.Reply(&disgord.Embed{
			Color: msg.HighestColor(msg.Message.GuildID, targetUser.ID),
			Title: targetUser.Tag(),
			Image: &disgord.EmbedImage{URL: AvatarURL(targetUser, 1024)},
		})
	}
}

func AvatarURL(u *disgord.User, size int) string {
	a, _ := u.AvatarURL(size, true)
	return a
}
