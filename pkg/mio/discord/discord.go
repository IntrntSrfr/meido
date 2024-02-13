package discord

import (
	"errors"
	"fmt"
	"image"
	"io"
	"runtime/debug"
	"sort"
	"time"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

// Discord represents the part of the bot that deals with interaction with Discord.
type Discord struct {
	token    string
	Sess     DiscordSession
	Sessions []DiscordSession
	shards   int

	messageChan     chan *DiscordMessage
	interactionChan chan *DiscordInteraction
	logger          *zap.Logger
}

type DiscordSession interface {
	Open() error
	Close() error
	ShardID() int
	State() *discordgo.State
	Real() *discordgo.Session

	AddHandler(handler interface{}) func()
	AddHandlerOnce(handler interface{}) func()
	Channel(channelID string, options ...discordgo.RequestOption) (st *discordgo.Channel, err error)
	ChannelFileSend(channelID, name string, r io.Reader, options ...discordgo.RequestOption) (*discordgo.Message, error)
	ChannelMessageDelete(channelID string, messageID string, options ...discordgo.RequestOption) (err error)
	ChannelMessageEdit(channelID string, messageID string, content string, options ...discordgo.RequestOption) (*discordgo.Message, error)
	ChannelMessageEditComplex(m *discordgo.MessageEdit, options ...discordgo.RequestOption) (st *discordgo.Message, err error)
	ChannelMessageEditEmbed(channelID string, messageID string, embed *discordgo.MessageEmbed, options ...discordgo.RequestOption) (*discordgo.Message, error)
	ChannelMessageEditEmbeds(channelID string, messageID string, embeds []*discordgo.MessageEmbed, options ...discordgo.RequestOption) (*discordgo.Message, error)
	ChannelMessagePin(channelID string, messageID string, options ...discordgo.RequestOption) (err error)
	ChannelMessageSend(channelID string, content string, options ...discordgo.RequestOption) (*discordgo.Message, error)
	ChannelMessageSendComplex(channelID string, data *discordgo.MessageSend, options ...discordgo.RequestOption) (st *discordgo.Message, err error)
	ChannelMessageSendEmbed(channelID string, embed *discordgo.MessageEmbed, options ...discordgo.RequestOption) (*discordgo.Message, error)
	ChannelMessageSendEmbedReply(channelID string, embed *discordgo.MessageEmbed, reference *discordgo.MessageReference, options ...discordgo.RequestOption) (*discordgo.Message, error)
	ChannelMessageSendEmbeds(channelID string, embeds []*discordgo.MessageEmbed, options ...discordgo.RequestOption) (*discordgo.Message, error)
	ChannelMessageSendEmbedsReply(channelID string, embeds []*discordgo.MessageEmbed, reference *discordgo.MessageReference, options ...discordgo.RequestOption) (*discordgo.Message, error)
	ChannelMessageSendReply(channelID string, content string, reference *discordgo.MessageReference, options ...discordgo.RequestOption) (*discordgo.Message, error)
	ChannelMessages(channelID string, limit int, beforeID string, afterID string, aroundID string, options ...discordgo.RequestOption) (st []*discordgo.Message, err error)
	ChannelMessagesBulkDelete(channelID string, messages []string, options ...discordgo.RequestOption) (err error)
	ChannelPermissionSet(channelID, targetID string, targetType discordgo.PermissionOverwriteType, allow, deny int64, options ...discordgo.RequestOption) (err error)
	ChannelTyping(channelID string, options ...discordgo.RequestOption) (err error)
	Guild(guildID string, options ...discordgo.RequestOption) (st *discordgo.Guild, err error)
	GuildBanCreate(guildID string, userID string, days int, options ...discordgo.RequestOption) (err error)
	GuildBanCreateWithReason(guildID string, userID string, reason string, days int, options ...discordgo.RequestOption) (err error)
	GuildBanDelete(guildID string, userID string, options ...discordgo.RequestOption) (err error)
	GuildBans(guildID string, limit int, beforeID string, afterID string, options ...discordgo.RequestOption) (st []*discordgo.GuildBan, err error)
	GuildChannels(guildID string, options ...discordgo.RequestOption) (st []*discordgo.Channel, err error)
	GuildIcon(guildID string, options ...discordgo.RequestOption) (img image.Image, err error)
	GuildMember(guildID string, userID string, options ...discordgo.RequestOption) (st *discordgo.Member, err error)
	GuildMemberAdd(guildID string, userID string, data *discordgo.GuildMemberAddParams, options ...discordgo.RequestOption) (err error)
	GuildMemberDelete(guildID string, userID string, options ...discordgo.RequestOption) (err error)
	GuildMemberDeleteWithReason(guildID string, userID string, reason string, options ...discordgo.RequestOption) (err error)
	GuildMemberRoleAdd(guildID string, userID string, roleID string, options ...discordgo.RequestOption) (err error)
	GuildMemberRoleRemove(guildID string, userID string, roleID string, options ...discordgo.RequestOption) (err error)
	GuildMemberTimeout(guildID string, userID string, until *time.Time, options ...discordgo.RequestOption) (err error)
	GuildMembers(guildID string, after string, limit int, options ...discordgo.RequestOption) (st []*discordgo.Member, err error)
	GuildRoleCreate(guildID string, data *discordgo.RoleParams, options ...discordgo.RequestOption) (st *discordgo.Role, err error)
	GuildRoleDelete(guildID string, roleID string, options ...discordgo.RequestOption) (err error)
	GuildRoleEdit(guildID string, roleID string, data *discordgo.RoleParams, options ...discordgo.RequestOption) (st *discordgo.Role, err error)
	GuildRoles(guildID string, options ...discordgo.RequestOption) (st []*discordgo.Role, err error)
	GuildSplash(guildID string, options ...discordgo.RequestOption) (img image.Image, err error)
	RequestGuildMembers(guildID string, query string, limit int, nonce string, presences bool) error
	RequestGuildMembersBatch(guildIDs []string, query string, limit int, nonce string, presences bool) (err error)
	RequestGuildMembersBatchList(guildIDs []string, userIDs []string, limit int, nonce string, presences bool) (err error)
	RequestGuildMembersList(guildID string, userIDs []string, limit int, nonce string, presences bool) error
	User(userID string, options ...discordgo.RequestOption) (st *discordgo.User, err error)
	UserChannelCreate(recipientID string, options ...discordgo.RequestOption) (st *discordgo.Channel, err error)
	UpdateStatusComplex(usd discordgo.UpdateStatusData) (err error)
	InteractionRespond(interaction *discordgo.Interaction, resp *discordgo.InteractionResponse, options ...discordgo.RequestOption) error
}

type SessionWrapper struct {
	*discordgo.Session
}

func (s *SessionWrapper) ShardID() int {
	return s.Session.ShardID
}
func (s *SessionWrapper) State() *discordgo.State {
	return s.Session.State
}

func (s *SessionWrapper) Real() *discordgo.Session {
	return s.Session
}

// NewDiscord takes in a token and creates a Discord object.
func NewDiscord(token string, shards int, logger *zap.Logger) *Discord {
	logger = logger.Named("Discord")
	d := &Discord{
		token:           token,
		shards:          shards,
		messageChan:     make(chan *DiscordMessage, 1),
		interactionChan: make(chan *DiscordInteraction, 1),
		logger:          logger,
	}
	discordgo.Logger = discordgoLogger(logger)
	d.createSessions()
	return d
}

// createSessions populates the Discord object with Sessions and returns a DiscordMessage channel.
func (d *Discord) createSessions() {
	d.Sessions = make([]DiscordSession, d.shards)
	for i := 0; i < d.shards; i++ {
		s, _ := discordgo.New("Bot " + d.token)

		s.State.MaxMessageCount = 100
		s.State.TrackVoice = false
		s.State.TrackPresences = false
		s.ShardCount = d.shards
		s.ShardID = i
		s.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAllWithoutPrivileged | discordgo.IntentsGuildMembers | discordgo.IntentMessageContent)

		s.AddHandler(d.onMessageCreate)
		s.AddHandler(d.onMessageUpdate)
		s.AddHandler(d.onMessageDelete)
		s.AddHandler(d.onInteractionCreate)

		d.Sessions[i] = &SessionWrapper{s}
		d.logger.Info("Added session", zap.Int("sessionID", i))
	}
	d.Sess = d.Sessions[0]
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
			d.logger.Error("Failed to close session", zap.Int("shardID", sess.ShardID()), zap.Error(err))
		}
	}
}

func (d *Discord) Messages() chan *DiscordMessage {
	return d.messageChan
}

func (d *Discord) Interactions() chan *DiscordInteraction {
	return d.interactionChan
}

func discordgoLogger(logger *zap.Logger) func(msgL, caller int, format string, a ...interface{}) {
	logger = logger.Named("DiscordGo")
	return func(msgL, caller int, format string, a ...interface{}) {
		msg := fmt.Sprintf(format, a...)
		switch msgL {
		case discordgo.LogError:
			logger.Error(msg)
		case discordgo.LogWarning:
			logger.Warn(msg)
		case discordgo.LogInformational:
			logger.Info(msg)
		case discordgo.LogDebug:
			logger.Debug(msg)
		}
	}
}

// botRecover is the recovery function used in the message create and update handler.
func (d *Discord) botRecover(i interface{}) {
	if r := recover(); r != nil {
		d.logger.Warn("Recovery needed",
			zap.Any("error", r),
			zap.Any("message", i),
			zap.String("stack trace", string(debug.Stack())),
		)
	}
}

func (d *Discord) onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Message == nil || m.Author == nil || m.Author.Bot {
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
		Sess:         d.Sess,
		Discord:      d,
		Message:      m.Message,
		MessageType:  MessageTypeCreate,
		TimeReceived: time.Now(),
		Shard:        s.ShardID,
	}
}

func (d *Discord) onMessageUpdate(s *discordgo.Session, m *discordgo.MessageUpdate) {
	if m.Message == nil || m.Author == nil || m.Author.Bot {
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
		Sess:         d.Sess,
		Discord:      d,
		Message:      m.Message,
		MessageType:  MessageTypeUpdate,
		TimeReceived: time.Now(),
		Shard:        s.ShardID,
	}
}

func (d *Discord) onMessageDelete(s *discordgo.Session, m *discordgo.MessageDelete) {
	d.messageChan <- &DiscordMessage{
		Sess:         d.Sess,
		Discord:      d,
		Message:      m.Message,
		MessageType:  MessageTypeDelete,
		TimeReceived: time.Now(),
		Shard:        s.ShardID,
	}
}

func (d *Discord) onInteractionCreate(s *discordgo.Session, m *discordgo.InteractionCreate) {
	if m.GuildID != "" && m.Member == nil {
		return
	}
	if m.GuildID == "" && m.User == nil {
		return
	}

	d.interactionChan <- &DiscordInteraction{
		Sess:         d.Sess,
		Discord:      d,
		Interaction:  m.Interaction,
		TimeReceived: time.Now(),
		Shard:        s.ShardID,
	}
}

var (
	ErrMissingArgs = errors.New("missing one or more required arguments")
)

func (d *Discord) BotUser() *discordgo.User {
	return d.Sess.State().User
}

// UserChannelPermissions finds member permissions the usual way, using just the IDs.
func (d *Discord) UserChannelPermissions(userID, channelID string) (int64, error) {
	var (
		err         error
		permissions int64
	)
	for _, s := range d.Sessions {
		permissions, err = s.State().UserChannelPermissions(userID, channelID)
		if err == nil {
			return permissions, nil
		}
	}
	return permissions, err
}

// BotHasPermissions finds if the bot user has permissions in a channel.
func (d *Discord) BotHasPermissions(channelID string, perm int64) (bool, error) {
	uPerms, err := d.UserChannelPermissions(d.Sess.State().User.ID, channelID)
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
		guilds = append(guilds, sess.State().Guilds...)
	}
	return guilds
}

// GuildCount returns the amount of guilds shared between all sessions.
func (d *Discord) GuildCount() int {
	var amount int
	for _, sess := range d.Sessions {
		amount += len(sess.State().Guilds)
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
		guild, err = sess.State().Guild(guildID)
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
		channel, err = sess.State().Channel(channelID)
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
		mem, err = sess.State().Member(guildID, userID)
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
		role, err = sess.State().Role(guildID, roleID)
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
