package meidov2

import (
	"github.com/andersfylling/disgord"
	"strings"
	"time"
)

type MessageType int

const (
	MessageTypeCreate MessageType = iota
	MessageTypeUpdate
	MessageTypeDelete
)

type DiscordMessage struct {
	Discord        *Discord
	DiscordMessage *disgord.Message
	Type           MessageType
	TimeReceived   time.Time
}

func (m *DiscordMessage) Args() []string {
	return strings.Fields(m.DiscordMessage.Content)
}

func (m *DiscordMessage) LenArgs() int {
	return len(m.Args())
}
