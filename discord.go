package meidov2

import (
	"github.com/andersfylling/disgord"
	"time"
)

type Discord struct {
	token  string
	Client *disgord.Client
	//sessions []*disgord.Session
	ownerIds []string

	messageChan chan *DiscordMessage
}

func NewDiscord(token string) *Discord {
	return &Discord{
		token:       token,
		messageChan: make(chan *DiscordMessage, 256),
	}
}

func (d *Discord) Open() (<-chan *DiscordMessage, error) {

	s := disgord.New(disgord.Config{
		BotToken:           d.token,
		LoadMembersQuietly: true,
		Intents:            disgord.AllIntents(disgord.IntentGuildPresences, disgord.IntentGuildMembers),
	})

	d.Client = s

	s.Gateway().MessageCreate(d.onMessageCreate)
	s.Gateway().MessageUpdate(d.onMessageUpdate)
	s.Gateway().MessageDelete(d.onMessageDelete)
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
	return d.Client.Gateway().Connect()
}

func (d *Discord) onMessageCreate(s disgord.Session, m *disgord.MessageCreate) {
	d.messageChan <- &DiscordMessage{
		Sess:         s,
		Discord:      d,
		Message:      m.Message,
		Type:         MessageTypeCreate,
		TimeReceived: time.Now(),
	}
}
func (d *Discord) onMessageUpdate(s disgord.Session, m *disgord.MessageUpdate) {
	d.messageChan <- &DiscordMessage{
		Sess:         s,
		Discord:      d,
		Message:      m.Message,
		Type:         MessageTypeUpdate,
		TimeReceived: time.Now(),
	}
}

func (d *Discord) onMessageDelete(s disgord.Session, m *disgord.MessageDelete) {
	d.messageChan <- &DiscordMessage{
		Sess:    s,
		Discord: d,
		Message: &disgord.Message{
			ID:        m.MessageID,
			ChannelID: m.ChannelID,
			GuildID:   m.GuildID,
		},
		Type:         MessageTypeDelete,
		TimeReceived: time.Now(),
	}
}
