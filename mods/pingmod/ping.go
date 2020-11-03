package pingmod

import (
	"context"
	"fmt"
	"github.com/andersfylling/disgord"
	"github.com/intrntsrfr/meidov2"
	"github.com/jmoiron/sqlx"
	"time"
)

type PingMod struct {
	cl       chan *meidov2.DiscordMessage
	commands []func(msg *meidov2.DiscordMessage)
}

func New() meidov2.Mod {
	return &PingMod{
		//cl: make(chan *meidov2.DiscordMessage),
	}
}

func (m *PingMod) Save() error {
	return nil
}

func (m *PingMod) Load() error {
	return nil
}

func (m *PingMod) Settings(msg *meidov2.DiscordMessage) {

}
func (m *PingMod) Help(msg *meidov2.DiscordMessage) {

}
func (m *PingMod) Commands() []meidov2.ModCommand {
	return nil
}

func (m *PingMod) Hook(b *meidov2.Bot, _ *sqlx.DB, cl chan *meidov2.DiscordMessage) error {
	m.cl = cl

	b.Discord.Client.On(disgord.EvtReady, func(s disgord.Session, r *disgord.Ready) {
		fmt.Println(len(r.Guilds))
		fmt.Println(r.User.String())
	})

	m.commands = append(m.commands, m.PingCommand)
	//m.commands = append(m.commands, m.check)

	return nil
}

func (m *PingMod) Message(msg *meidov2.DiscordMessage) {
	if msg.Type != meidov2.MessageTypeCreate {
		return
	}
	for _, c := range m.commands {
		go c(msg)
	}
}

func (m *PingMod) PingCommand(msg *meidov2.DiscordMessage) {
	if msg.Message.Content != "m?ping" {
		return
	}

	m.cl <- msg

	startTime := time.Now()

	first, err := msg.Message.Reply(context.Background(), msg.Discord.Client, "Ping")
	if err != nil {
		return
	}

	now := time.Now()
	discordLatency := now.Sub(startTime)
	botLatency := now.Sub(msg.TimeReceived)

	msg.Discord.Client.SetMsgContent(context.Background(), first.ChannelID, first.ID,
		fmt.Sprintf("Test Pong!\nDiscord delay: %s\nBot delay: %s",
			discordLatency, botLatency))
}
