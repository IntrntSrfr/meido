package meidov2

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"time"
)

type Discord struct {
	token    string
	Client   *discordgo.Session
	sessions []*discordgo.Session
	ownerIds []string

	messageChan chan *DiscordMessage
}

func NewDiscord(token string) *Discord {
	return &Discord{
		token:       token,
		messageChan: make(chan *DiscordMessage, 256),
	}
}

type Log struct {
}

func (l *Log) Debug(v ...interface{}) {
	fmt.Println(v)
}
func (l *Log) Info(v ...interface{}) {
	fmt.Println(v)
}
func (l *Log) Error(v ...interface{}) {
	fmt.Println(v)
}

func (d *Discord) Open() (<-chan *DiscordMessage, error) {

	s, err := discordgo.New("Bot " + d.token)
	if err != nil {
		return nil, err
	}

	s.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAllWithoutPrivileged)
	d.Client = s

	s.AddHandler(d.onMessageCreate)
	s.AddHandler(d.onMessageUpdate)
	s.AddHandler(d.onMessageDelete)

	/*
		err := s.Connect(context.Background())
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
	*/
	//go d.listen()

	return d.messageChan, nil
}

func (d *Discord) Run() error {
	return d.Client.Open()
}

func (d *Discord) onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	d.messageChan <- &DiscordMessage{
		Sess:         s,
		Discord:      d,
		Message:      m.Message,
		Type:         MessageTypeCreate,
		TimeReceived: time.Now(),
	}
}
func (d *Discord) onMessageUpdate(s *discordgo.Session, m *discordgo.MessageUpdate) {
	d.messageChan <- &DiscordMessage{
		Sess:         s,
		Discord:      d,
		Message:      m.Message,
		Type:         MessageTypeUpdate,
		TimeReceived: time.Now(),
	}
}

func (d *Discord) onMessageDelete(s *discordgo.Session, m *discordgo.MessageDelete) {
	d.messageChan <- &DiscordMessage{
		Sess:         s,
		Discord:      d,
		Message:      m.Message,
		Type:         MessageTypeDelete,
		TimeReceived: time.Now(),
	}
}
