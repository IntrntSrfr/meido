package meidov2

import (
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"net/http"
	"time"
)

type Discord struct {
	token    string
	Sess     *discordgo.Session
	Sessions []*discordgo.Session
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
	req, _ := http.NewRequest("GET", "https://discord.com/api/v8/gateway/bot", nil)
	req.Header.Add("Authorization", "Bot "+d.token)
	res, err := http.DefaultClient.Do(req)
	defer res.Body.Close()
	if err != nil {
		panic(err)
	}
	resp := &discordgo.GatewayBotResponse{}
	err = json.NewDecoder(res.Body).Decode(&resp)
	if err != nil {
		panic(err)
	}

	shardCount := resp.Shards
	d.Sessions = make([]*discordgo.Session, shardCount)

	for i := 0; i < shardCount; i++ {
		s, err := discordgo.New("Bot " + d.token)
		if err != nil {
			return nil, err
		}

		s.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAllWithoutPrivileged)
		s.State.TrackVoice = false
		s.State.TrackPresences = false
		s.ShardCount = shardCount
		s.ShardID = i

		s.AddHandler(d.onMessageCreate)
		s.AddHandler(d.onMessageUpdate)
		s.AddHandler(d.onMessageDelete)

		d.Sessions[i] = s
		fmt.Println("created session:", i)
	}
	d.Sess = d.Sessions[0]

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
	for _, sess := range d.Sessions {
		sess.Open()
	}
	return nil
}

func (d *Discord) Close() {
	for _, sess := range d.Sessions {
		sess.Close()
	}
}

func (d *Discord) onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	d.messageChan <- &DiscordMessage{
		Sess:         s,
		Discord:      d,
		Message:      m.Message,
		Type:         MessageTypeCreate,
		TimeReceived: time.Now(),
		Shard:        s.ShardID,
	}
}

func (d *Discord) onMessageUpdate(s *discordgo.Session, m *discordgo.MessageUpdate) {
	d.messageChan <- &DiscordMessage{
		Sess:         s,
		Discord:      d,
		Message:      m.Message,
		Type:         MessageTypeUpdate,
		TimeReceived: time.Now(),
		Shard:        s.ShardID,
	}
}

func (d *Discord) onMessageDelete(s *discordgo.Session, m *discordgo.MessageDelete) {
	d.messageChan <- &DiscordMessage{
		Sess:         s,
		Discord:      d,
		Message:      m.Message,
		Type:         MessageTypeDelete,
		TimeReceived: time.Now(),
		Shard:        s.ShardID,
	}
}
