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
	if m.Message.Author == nil || m.Message.Author.Bot {
		return
	}
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
	if m.Message.Author == nil || m.Message.Author.Bot {
		return
	}
	d.messageChan <- &DiscordMessage{
		Sess:         s,
		Discord:      d,
		Message:      m.Message,
		Author:       m.Member,
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

func (s *Discord) UserChannelPermissions(m *discordgo.Member, channelID string) (apermissions int, err error) {

	channel, err := s.Sess.State.Channel(channelID)
	if err != nil {
		return
	}

	guild, err := s.Sess.State.Guild(channel.GuildID)
	if err != nil {
		return
	}

	if m.User.ID == guild.OwnerID {
		apermissions = discordgo.PermissionAll
		return
	}

	return memberPermissions(guild, channel, m), nil
}

// Calculates the permissions for a member.
// https://support.discord.com/hc/en-us/articles/206141927-How-is-the-permission-hierarchy-structured-
func memberPermissions(guild *discordgo.Guild, channel *discordgo.Channel, member *discordgo.Member) (apermissions int) {
	userID := member.User.ID

	if userID == guild.OwnerID {
		apermissions = discordgo.PermissionAll
		return
	}

	for _, role := range guild.Roles {
		if role.ID == guild.ID {
			apermissions |= role.Permissions
			break
		}
	}

	for _, role := range guild.Roles {
		for _, roleID := range member.Roles {
			if role.ID == roleID {
				apermissions |= role.Permissions
				break
			}
		}
	}

	if apermissions&discordgo.PermissionAdministrator == discordgo.PermissionAdministrator {
		apermissions |= discordgo.PermissionAll
	}

	// Apply @everyone overrides from the channel.
	for _, overwrite := range channel.PermissionOverwrites {
		if guild.ID == overwrite.ID {
			apermissions &= ^overwrite.Deny
			apermissions |= overwrite.Allow
			break
		}
	}

	denies := 0
	allows := 0

	// Member overwrites can override role overrides, so do two passes
	for _, overwrite := range channel.PermissionOverwrites {
		for _, roleID := range member.Roles {
			if overwrite.Type == "role" && roleID == overwrite.ID {
				denies |= overwrite.Deny
				allows |= overwrite.Allow
				break
			}
		}
	}

	apermissions &= ^denies
	apermissions |= allows

	for _, overwrite := range channel.PermissionOverwrites {
		if overwrite.Type == "member" && overwrite.ID == userID {
			apermissions &= ^overwrite.Deny
			apermissions |= overwrite.Allow
			break
		}
	}

	if apermissions&discordgo.PermissionAdministrator == discordgo.PermissionAdministrator {
		apermissions |= discordgo.PermissionAllChannel
	}

	return apermissions
}
