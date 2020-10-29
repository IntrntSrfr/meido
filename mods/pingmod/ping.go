package pingmod

import (
	"context"
	"fmt"
	"github.com/andersfylling/disgord"
	"github.com/intrntsrfr/meidov2"
	"time"
)

type PingMod struct {
	cl chan *meidov2.DiscordMessage
	commands []func(msg *meidov2.DiscordMessage)
}

func New() meidov2.Mod {
	return &PingMod{
		cl: make(chan *meidov2.DiscordMessage, 256),
	}
}

func (m *PingMod) Hook(b *meidov2.Bot, cl chan *meidov2.DiscordMessage) error {
	m.cl = cl

	b.Discord.Client.On(disgord.EvtReady, func(s disgord.Session, r *disgord.Ready) {
		fmt.Println(r.User.String())
	})

	m.commands = append(m.commands, m.PingCommand)

	return nil
}

func (m *PingMod) Message(msg *meidov2.DiscordMessage) {
	for _, c := range m.commands{
		go c(msg)
	}
}

func (m *PingMod) PingCommand(msg *meidov2.DiscordMessage) {
	if msg.DiscordMessage.Content != "m?ping" {
		return
	}

	m.cl<-msg

	startTime := time.Now()

	first, err := msg.DiscordMessage.Reply(context.Background(), msg.Discord.Client, "Ping")
	if err != nil {
		return
	}

	now := time.Now()
	discordLatency := now.Sub(startTime)
	botLatency := now.Sub(msg.TimeReceived)

	msg.Discord.Client.SetMsgContent(context.Background(), first.ChannelID, first.ID,
		fmt.Sprintf("Pong!\nDiscord delay: %s\nBot delay: %s",
			discordLatency, botLatency))
}
