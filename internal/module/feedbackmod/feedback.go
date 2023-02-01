package feedbackmod

import (
	"fmt"
	"github.com/intrntsrfr/meido/pkg/mio"
)

type FeedbackMod struct {
	*mio.ModuleBase
	bot             *mio.Bot
	bannedUsers     map[string]bool
	feedbackChannel string
	owners          []string
}

func New(b *mio.Bot, ownerIds []string, feedbackCh string) mio.Module {
	return &FeedbackMod{
		ModuleBase:      mio.NewModule("Feedback"),
		bot:             b,
		bannedUsers:     make(map[string]bool),
		feedbackChannel: feedbackCh,
		owners:          ownerIds,
	}
}

func (m *FeedbackMod) Hook() error {
	return nil
}

func (m *FeedbackMod) ToggleBan(msg *mio.DiscordMessage) {
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

func (m *FeedbackMod) LeaveFeedback(msg *mio.DiscordMessage) {
	if msg.LenArgs() <= 1 || msg.Args()[0] != "m?feedback" {
		return
	}

	m.Lock()
	defer m.Unlock()
	banned, ok := m.bannedUsers[msg.Message.Author.ID]
	if ok {
		if banned {
			_, _ = msg.Reply("You're banned from using the feedback feature.")
			return
		}
	}

	_, _ = msg.Discord.Sess.ChannelMessageSend(m.feedbackChannel, fmt.Sprintf(`%v`, msg.RawArgs()[1:]))
	_, _ = msg.Reply("Feedback left")
}
