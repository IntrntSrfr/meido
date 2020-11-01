package feedbackmod

import (
	"context"
	"fmt"
	"github.com/andersfylling/disgord"
	"github.com/intrntsrfr/meidov2"
	"strconv"
	"sync"
)

type FeedbackMod struct {
	cl          chan *meidov2.DiscordMessage
	commands    []func(msg *meidov2.DiscordMessage)
	bannedUsers map[disgord.Snowflake]bool
	sync.Mutex
	feedbackChannel disgord.Snowflake
}

func New() meidov2.Mod {
	return &FeedbackMod{
		cl:              make(chan *meidov2.DiscordMessage, 256),
		bannedUsers:     make(map[disgord.Snowflake]bool),
		feedbackChannel: disgord.Snowflake(497106582144942101),
	}
}

func (m *FeedbackMod) Save() error {
	return nil
}

func (m *FeedbackMod) Load() error {
	return nil
}

func (m *FeedbackMod) Settings(msg *meidov2.DiscordMessage) {

}
func (m *FeedbackMod) Help(msg *meidov2.DiscordMessage) {

}

func (m *FeedbackMod) Hook(b *meidov2.Bot, cl chan *meidov2.DiscordMessage) error {
	m.cl = cl

	b.Discord.Client.On(disgord.EvtReady, func(s disgord.Session, r *disgord.Ready) {
		fmt.Println(r.User.String())
	})

	m.commands = append(m.commands, m.LeaveFeedback)

	return nil
}

func (m *FeedbackMod) Message(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() == 0 {
		return
	}
	for _, c := range m.commands {
		go c(msg)
	}
}

func (m *FeedbackMod) ToggleBan(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() <= 1 || msg.Args()[0] != "m?togglefeedback" || msg.DiscordMessage.Author.ID != disgord.Snowflake(163454407999094786) {
		return
	}
	memId, err := strconv.Atoi(msg.Args()[1])
	if err != nil {
		return
	}

	m.Lock()
	defer m.Unlock()
	b, ok := m.bannedUsers[disgord.Snowflake(memId)]
	if ok {
		if b {
			m.bannedUsers[disgord.Snowflake(memId)] = false
			// send unbanned mesage
		} else {
			m.bannedUsers[disgord.Snowflake(memId)] = true
			// send banned message
		}
	} else {
		m.bannedUsers[disgord.Snowflake(memId)] = true
		// send banned message
	}
}

func (m *FeedbackMod) LeaveFeedback(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() <= 1 || msg.Args()[0] != "m?feedback" {
		return
	}

	m.cl <- msg

	m.Lock()
	defer m.Unlock()
	banned, ok := m.bannedUsers[msg.DiscordMessage.Author.ID]
	if ok {
		if banned {
			msg.Discord.Client.SendMsg(context.Background(), msg.DiscordMessage.ChannelID, "You're banned from using the feedback feature.")
			return
		}
	}

	msg.Discord.Client.SendMsg(context.Background(), m.feedbackChannel, fmt.Sprintf(`%v`, msg.Args()[1:]))
	msg.Discord.Client.SendMsg(context.Background(), msg.DiscordMessage.ChannelID, "Feedback left")
}
