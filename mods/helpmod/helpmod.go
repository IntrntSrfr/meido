package helpmod

import (
	"github.com/intrntsrfr/meidov2"
)

type HelpMod struct {
	cl       chan *meidov2.DiscordMessage
	commands []func(msg *meidov2.DiscordMessage)
	b        *meidov2.Bot
}

func New() meidov2.Mod {
	return &HelpMod{}
}

func (m *HelpMod) Save() error {
	return nil
}

func (m *HelpMod) Load() error {
	return nil
}

func (m *HelpMod) Settings(msg *meidov2.DiscordMessage) {

}
func (m *HelpMod) Help(msg *meidov2.DiscordMessage) {

}
func (m *HelpMod) Commands() map[string]meidov2.ModCommand {
	return nil
}

func (m *HelpMod) Hook(b *meidov2.Bot) error {
	m.cl = b.CommandLog
	m.b = b

	return nil
}

func (m *HelpMod) Message(msg *meidov2.DiscordMessage) {
	if msg.Type != meidov2.MessageTypeCreate {
		return
	}
	for _, c := range m.commands {
		go c(msg)
	}
}
