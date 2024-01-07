package mio

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"sort"
	"time"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

// Discord represents the part of the bot that deals with interaction with Discord.
type Discord struct {
	token    string
	Sess     *discordgo.Session
	Sessions []*discordgo.Session
	ownerIds []string

	messageChan chan *DiscordMessage
	logger      *zap.Logger
}

// NewDiscord takes in a token and creates a Discord object.
func NewDiscord(token string, logger *zap.Logger) *Discord {
	d := &Discord{
		token:       token,
		messageChan: make(chan *DiscordMessage, 256),
		logger:      logger.Named("discord"),
	}
	discordgo.Logger = discordgoLogger(d.logger.Named("discordgo"))
	return d
}

// Open populates the Discord object with Sessions and returns a DiscordMessage channel.
func (d *Discord) Open() error {
	shardCount, err := recommendedShards(d.token)
	if err != nil {
		return err
	}

	d.Sessions = make([]*discordgo.Session, shardCount)
	for i := 0; i < shardCount; i++ {
		s, err := discordgo.New("Bot " + d.token)
		if err != nil {
			return err
		}

		s.State.MaxMessageCount = 100
		s.State.TrackVoice = false
		s.State.TrackPresences = false
		s.ShardCount = shardCount
		s.ShardID = i
		s.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAllWithoutPrivileged | discordgo.IntentsGuildMembers | discordgo.IntentMessageContent)

		s.AddHandler(d.onMessageCreate)
		s.AddHandler(d.onMessageUpdate)
		s.AddHandler(d.onMessageDelete)

		d.Sessions[i] = s
		d.logger.Info("created session", zap.Int("sessionID", i))
	}
	d.Sess = d.Sessions[0]

	return nil
}

// Run opens the Discord sessions.
func (d *Discord) Run() error {
	for _, sess := range d.Sessions {
		if err := sess.Open(); err != nil {
			return err
		}
	}
	return nil
}

// Close closes the Discord sessions
func (d *Discord) Close() {
	for _, sess := range d.Sessions {
		err := sess.Close()
		if err != nil {
			d.logger.Error("failed to close session", zap.Int("shardID", sess.ShardID), zap.Error(err))
		}
	}
}

// recommendedShards asks discord for the recommended shardcount for the bot given the token.
// returns -1 if the request does not go well.
func recommendedShards(token string) (int, error) {
	req, _ := http.NewRequest("GET", "https://discord.com/api/v10/gateway/bot", nil)
	req.Header.Add("Authorization", "Bot "+token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return -1, err
	}
	defer res.Body.Close()

	var resp discordgo.GatewayBotResponse
	err = json.NewDecoder(res.Body).Decode(&resp)
	if err != nil {
		return -1, err
	}

	return resp.Shards, nil
}

func discordgoLogger(l *zap.Logger) func(msgL, caller int, format string, a ...interface{}) {
	return func(msgL, caller int, format string, a ...interface{}) {
		msg := fmt.Sprintf(format, a...)
		switch msgL {
		case discordgo.LogError:
			l.Error(msg)
		case discordgo.LogWarning:
			l.Warn(msg)
		case discordgo.LogInformational:
			l.Info(msg)
		case discordgo.LogDebug:
			l.Debug(msg)
		}
	}
}

// botRecover is the recovery function used in the message create and update handler.
func (d *Discord) botRecover(i interface{}) {
	if r := recover(); r != nil {
		d.logger.Error("recovery needed",
			zap.Any("error", r),
			zap.Any("message", i),
			zap.String("stack trace", string(debug.Stack())),
		)
	}
}

// onMessageCreate is the handler for the *discordgo.MessageCreate event.
// It populates a DiscordMessage object and sends it to Discord.messageChan
func (d *Discord) onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author == nil || m.Message.Author.Bot {
		return
	}
	defer d.botRecover(m)

	if m.GuildID != "" {
		// for some reason the member field might still be nil and no one knows why
		if m.Member == nil {
			return
		}
		m.Member.User = m.Author
		m.Member.GuildID = m.GuildID
	}

	d.messageChan <- &DiscordMessage{
		Sess:         s,
		Discord:      d,
		Message:      m.Message,
		MessageType:  MessageTypeCreate,
		TimeReceived: time.Now(),
		Shard:        s.ShardID,
	}
}

// onMessageUpdate is the handler for the *discordgo.MessageUpdate event.
// It populates a DiscordMessage object and sends it to Discord.messageChan
func (d *Discord) onMessageUpdate(s *discordgo.Session, m *discordgo.MessageUpdate) {
	if m.Author == nil || m.Message.Author.Bot {
		return
	}
	defer d.botRecover(m)

	if m.GuildID != "" {
		if m.Member == nil {
			return
		}
		m.Member.User = m.Author
		m.Member.GuildID = m.GuildID
	}

	d.messageChan <- &DiscordMessage{
		Sess:         s,
		Discord:      d,
		Message:      m.Message,
		MessageType:  MessageTypeUpdate,
		TimeReceived: time.Now(),
		Shard:        s.ShardID,
	}
}

// onMessageDelete is the handler for the *discordgo.MessageDelete event.
// It populates a DiscordMessage object and sends it to Discord.messageChan
func (d *Discord) onMessageDelete(s *discordgo.Session, m *discordgo.MessageDelete) {
	d.messageChan <- &DiscordMessage{
		Sess:         s,
		Discord:      d,
		Message:      m.Message,
		MessageType:  MessageTypeDelete,
		TimeReceived: time.Now(),
		Shard:        s.ShardID,
	}
}

var (
	ErrMissingArgs = errors.New("missing one or more required arguments")
)

func (d *Discord) BotUser() *discordgo.User {
	return d.Sess.State.User
}

// UserChannelPermissions finds member permissions the usual way, using just the IDs.
func (d *Discord) UserChannelPermissions(userID, channelID string) (int64, error) {
	var (
		err         error
		permissions int64
	)
	for _, s := range d.Sessions {
		permissions, err = s.State.UserChannelPermissions(userID, channelID)
		if err == nil {
			return permissions, nil
		}
	}
	return permissions, err
}

// BotHasPermissions finds if the bot user has permissions in a channel.
func (d *Discord) BotHasPermissions(channelID string, perm int64) (bool, error) {
	uPerms, err := d.UserChannelPermissions(d.Sess.State.User.ID, channelID)
	if err != nil {
		return false, err
	}
	return uPerms&(perm|discordgo.PermissionAdministrator) != 0, nil
}

// HasPermissions finds if the bot user has permissions in a channel.
func (d *Discord) HasPermissions(channelID, userID string, perm int64) (bool, error) {
	uPerms, err := d.UserChannelPermissions(userID, channelID)
	if err != nil {
		return false, err
	}
	return uPerms&(perm|discordgo.PermissionAdministrator) != 0, nil
}

// HighestRole finds the highest role a user has in the guild hierarchy.
func (d *Discord) HighestRole(gid, uid string) *discordgo.Role {
	g, err := d.Guild(gid)
	if err != nil {
		return nil
	}
	mem, err := d.Member(gid, uid)
	if err != nil {
		return nil
	}

	gRoles := g.Roles
	sort.Sort(RoleByPos(gRoles))
	for _, gr := range gRoles {
		for _, r := range mem.Roles {
			if r == gr.ID {
				return gr
			}
		}
	}
	return nil
}

// HighestRolePosition gets the highest role position a user has in the guild hierarchy.
func (d *Discord) HighestRolePosition(gid, uid string) int {
	role := d.HighestRole(gid, uid)
	if role == nil {
		return -1
	}
	return role.Position
}

// HighestColor finds the role color for the top-most role where the color is non-zero.
func (d *Discord) HighestColor(gid, uid string) int {
	g, err := d.Guild(gid)
	if err != nil {
		return 0
	}

	mem, err := d.Member(gid, uid)
	if err != nil {
		return 0
	}

	gRoles := g.Roles
	sort.Sort(RoleByPos(gRoles))
	for _, gr := range gRoles {
		for _, r := range mem.Roles {
			if r == gr.ID {
				if gr.Color != 0 {
					return gr.Color
				}
			}
		}
	}
	return 0
}

// RoleByPos is used to sort discordgo.Roles so that it accurately represents how the hierarchy looks.
type RoleByPos []*discordgo.Role

func (a RoleByPos) Len() int           { return len(a) }
func (a RoleByPos) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a RoleByPos) Less(i, j int) bool { return a[i].Position > a[j].Position }

// AddEventHandler adds an event handler to each discord session the bot holds
func (d *Discord) AddEventHandler(h interface{}) {
	for _, s := range d.Sessions {
		s.AddHandler(h)
	}
}

// AddEventHandlerOnce adds an event handler to the main session that will be fired once.
func (d *Discord) AddEventHandlerOnce(h interface{}) {
	for _, s := range d.Sessions {
		s.AddHandlerOnce(h)
	}
}

// Guilds returns all the guild objects the bot has in its sessions.
func (d *Discord) Guilds() []*discordgo.Guild {
	var guilds []*discordgo.Guild
	for _, sess := range d.Sessions {
		guilds = append(guilds, sess.State.Guilds...)
	}
	return guilds
}

// GuildCount returns the amount of guilds shared between all sessions.
func (d *Discord) GuildCount() int {
	var amount int
	for _, sess := range d.Sessions {
		amount += len(sess.State.Guilds)
	}
	return amount
}

// Guild takes in an ID and returns a discordgo.Guild if one with that ID exists.
func (d *Discord) Guild(guildID string) (*discordgo.Guild, error) {
	if guildID == "" {
		return nil, ErrMissingArgs
	}
	var err error
	var guild *discordgo.Guild
	for _, sess := range d.Sessions {
		guild, err = sess.State.Guild(guildID)
		if err == nil {
			return guild, nil
		}
	}
	return nil, err
}

// Channel takes in an ID and returns a discordgo.Channel if one with that ID exists.
func (d *Discord) Channel(channelID string) (*discordgo.Channel, error) {
	if channelID == "" {
		return nil, ErrMissingArgs
	}
	var err error
	var channel *discordgo.Channel
	for _, sess := range d.Sessions {
		channel, err = sess.State.Channel(channelID)
		if err == nil {
			return channel, nil
		}
	}
	return nil, err
}

// Member takes in a guild ID and a user ID and returns a discordgo.Member if one with such ID exists.
func (d *Discord) Member(guildID, userID string) (*discordgo.Member, error) {
	if guildID == "" || userID == "" {
		return nil, ErrMissingArgs
	}
	var err error
	var mem *discordgo.Member
	for _, sess := range d.Sessions {
		mem, err = sess.State.Member(guildID, userID)
		if err == nil {
			return mem, nil
		}
	}
	mem, err = d.Sess.GuildMember(guildID, userID)
	if err != nil {
		return nil, err
	}
	return mem, nil
}

// Role takes in a guild ID and a role ID and returns a discordgo.Role if one with such IDs exists.
func (d *Discord) Role(guildID, roleID string) (*discordgo.Role, error) {
	if guildID == "" || roleID == "" {
		return nil, ErrMissingArgs
	}
	var err error
	var role *discordgo.Role
	for _, sess := range d.Sessions {
		role, err = sess.State.Role(guildID, roleID)
		if err == nil {
			return role, nil
		}
	}
	return nil, err
}

func (d *Discord) GuildRoleByNameOrID(guildID, name, id string) (*discordgo.Role, error) {
	if guildID == "" || (name == "" && id == "") {
		return nil, ErrMissingArgs
	}
	g, err := d.Guild(guildID)
	if err != nil {
		return nil, err
	}
	for _, role := range g.Roles {
		if role.Name == name || role.ID == id {
			return role, nil
		}
	}
	return nil, discordgo.ErrStateNotFound
}

// IsBotOwner returns whether the author of a DiscordMessage is a bot owner by checking
// the IDs in the ownerIDs in the Discord struct.
func (d *Discord) IsBotOwner(msg *DiscordMessage) bool {
	for _, id := range d.ownerIds {
		if msg.Author().ID == id {
			return true
		}
	}
	return false
}

// StartTyping makes the bot show as 'typing...' in a channel.
func (d *Discord) StartTyping(channelID string) error {
	return d.Sess.ChannelTyping(channelID)
}

func (d *Discord) SendMessage(channelID, content string) (*discordgo.Message, error) {
	return d.Sess.ChannelMessageSend(channelID, content)
}

func (d *Discord) UpdateStatus(status string, activityType discordgo.ActivityType) {
	d.Sess.UpdateStatusComplex(discordgo.UpdateStatusData{
		Status: status,
		Activities: []*discordgo.Activity{
			{
				Type: activityType,
			},
		},
	})
}

func (d *Discord) IsOwner(userID string) bool {
	for _, id := range d.ownerIds {
		if id == userID {
			return true
		}
	}
	return false
}
