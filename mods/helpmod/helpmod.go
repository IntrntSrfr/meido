package helpmod

import (
	"github.com/intrntsrfr/meidov2"
	"github.com/jmoiron/sqlx"
)

type HelpMod struct {
	cl       chan *meidov2.DiscordMessage
	commands []func(msg *meidov2.DiscordMessage)
}

func New() meidov2.Mod {
	return &HelpMod{
		//cl: make(chan *meidov2.DiscordMessage),
	}
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
func (m *HelpMod) Commands() []meidov2.ModCommand {
	return nil
}

func (m *HelpMod) Hook(b *meidov2.Bot, _ *sqlx.DB, cl chan *meidov2.DiscordMessage) error {
	m.cl = cl

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
