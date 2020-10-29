package meidov2

import (
	"github.com/andersfylling/disgord"
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
	TimeReceived time.Time
}
