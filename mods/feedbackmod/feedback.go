package feedbackmod

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meidov2"
	"sync"
)

type FeedbackMod struct {
	sync.Mutex
	name string
	//cl              chan *meidov2.DiscordMessage
	commands        map[string]*meidov2.ModCommand // func(msg *meidov2.DiscordMessage)
	bannedUsers     map[string]bool
	feedbackChannel string
	owners          []string
	allowedTypes    meidov2.MessageType
	allowDMs        bool
}

func New(n string) meidov2.Mod {
	return &FeedbackMod{
		name:         n,
		commands:     make(map[string]*meidov2.ModCommand),
		allowedTypes: meidov2.MessageTypeCreate,
		allowDMs:     true,
	}
}

func (m *FeedbackMod) Name() string {
	return m.name
}
func (m *FeedbackMod) Save() error {
	return nil
}
func (m *FeedbackMod) Load() error {
	return nil
}
func (m *FeedbackMod) Passives() []*meidov2.ModPassive {
	return []*meidov2.ModPassive{}
}
func (m *FeedbackMod) Commands() map[string]*meidov2.ModCommand {
	return m.commands
}
func (m *FeedbackMod) AllowedTypes() meidov2.MessageType {
	return m.allowedTypes
}
func (m *FeedbackMod) AllowDMs() bool {
	return m.allowDMs
}
func (m *FeedbackMod) Hook(b *meidov2.Bot) error {
	//m.cl = b.CommandLog
	m.owners = b.Config.OwnerIds

	b.Discord.Sess.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		fmt.Println(r.User.String())
	})

	return nil
}
func (m *FeedbackMod) RegisterCommand(cmd *meidov2.ModCommand) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.commands[cmd.Name]; ok {
		panic(fmt.Sprintf("command '%v' already exists in %v", cmd.Name, m.Name()))
	}
	m.commands[cmd.Name] = cmd
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

	//m.cl <- msg

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

	//m.cl <- msg

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
