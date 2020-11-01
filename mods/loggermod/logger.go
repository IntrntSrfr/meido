package loggermod

import (
	"github.com/intrntsrfr/meidov2"
)

type LoggerMod struct {
	cl       chan *meidov2.DiscordMessage
	commands []func(msg *meidov2.DiscordMessage)
}

func New() meidov2.Mod {
	return &LoggerMod{
		cl: make(chan *meidov2.DiscordMessage, 256),
	}
}

func (m *LoggerMod) Save() error {
	return nil
}

func (m *LoggerMod) Load() error {
	return nil
}

func (m *LoggerMod) Settings(msg *meidov2.DiscordMessage) {

}

func (m *LoggerMod) Help(msg *meidov2.DiscordMessage) {

}

func (m *LoggerMod) Hook(b *meidov2.Bot, cl chan *meidov2.DiscordMessage) error {
	m.cl = cl
	/*
		b.Discord.Client.On(disgord.EvtGuildMemberAdd, func(s disgord.Session, mem *disgord.GuildMemberAdd) {
			fmt.Println(mem.Member.String())
		})
	*/
	return nil
}

func (m *LoggerMod) Message(msg *meidov2.DiscordMessage) {
	for _, c := range m.commands {
		go c(msg)
	}
}
