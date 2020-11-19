package feedbackmod

/*
import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meidov2"
	"sync"
)

type FeedbackMod struct {
	cl          chan *meidov2.DiscordMessage
	commands    []func(msg *meidov2.DiscordMessage)
	bannedUsers map[string]bool
	sync.Mutex
	feedbackChannel string
	owners          []string
}

func New() meidov2.Mod {
	return &FeedbackMod{
		bannedUsers:     make(map[string]bool),
		feedbackChannel: "497106582144942101",
		owners:          []string{},
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
func (m *FeedbackMod) Commands() map[string]meidov2.ModCommand {
	return nil
}

func (m *FeedbackMod) Hook(b *meidov2.Bot) error {
	m.cl = b.CommandLog
	m.owners = b.Config.OwnerIds

	b.Discord.Sess.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
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
		if msg.Message.Author.ID == id {
			owner = true
		}
	}
	if !owner {
		return
	}

	m.cl <- msg

	m.Lock()
	defer m.Unlock()
	b, ok := m.bannedUsers[msg.Args()[1]]
	if ok {
		if b {
			m.bannedUsers[msg.Args()[1]] = false
			// send unbanned mesage
		} else {
			m.bannedUsers[msg.Args()[1]] = true
			// send banned message
		}
	} else {
		m.bannedUsers[msg.Args()[1]] = true
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

	msg.Discord.Sess.ChannelMessageSend(m.feedbackChannel, fmt.Sprintf(`%v`, msg.Args()[1:]))
	msg.Reply("Feedback left")
}
*/
