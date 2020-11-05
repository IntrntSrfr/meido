package feedbackmod

import (
	"fmt"
	"github.com/andersfylling/disgord"
	"github.com/intrntsrfr/meidov2"
	"github.com/jmoiron/sqlx"
	"strconv"
	"sync"
)

type FeedbackMod struct {
	cl          chan *meidov2.DiscordMessage
	commands    []func(msg *meidov2.DiscordMessage)
	bannedUsers map[disgord.Snowflake]bool
	sync.Mutex
	feedbackChannel disgord.Snowflake
	owners          []int
}

func New() meidov2.Mod {
	return &FeedbackMod{
		//cl:              make(chan *meidov2.DiscordMessage),
		bannedUsers:     make(map[disgord.Snowflake]bool),
		feedbackChannel: disgord.Snowflake(497106582144942101),
		owners:          []int{},
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
func (m *FeedbackMod) Commands() []meidov2.ModCommand {
	return nil
}

func (m *FeedbackMod) Hook(b *meidov2.Bot, db *sqlx.DB, cl chan *meidov2.DiscordMessage) error {
	m.cl = cl

	m.owners = b.Config.OwnerIds

	b.Discord.Client.On(disgord.EvtReady, func(s disgord.Session, r *disgord.Ready) {
		fmt.Println(r.User.String())
	})

	m.commands = append(m.commands, m.LeaveFeedback)

	return nil
}

func (m *FeedbackMod) Message(msg *meidov2.DiscordMessage) {
	if msg.Type != meidov2.MessageTypeCreate {
		return
	}
	if msg.LenArgs() == 0 {
		return
	}
	for _, c := range m.commands {
		go c(msg)
	}
}

func (m *FeedbackMod) ToggleBan(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() <= 1 || msg.Args()[0] != "m?togglefeedback" {
		return
	}

	owner := false
	for _, id := range m.owners {
		if msg.Message.Author.ID == disgord.Snowflake(id) {
			owner = true
		}
	}
	if !owner {
		return
	}

	m.cl <- msg

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
	banned, ok := m.bannedUsers[msg.Message.Author.ID]
	if ok {
		if banned {
			msg.Reply("You're banned from using the feedback feature.")
			return
		}
	}

	msg.Discord.Client.SendMsg(m.feedbackChannel, fmt.Sprintf(`%v`, msg.Args()[1:]))
	msg.Reply("Feedback left")
}
