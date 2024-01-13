package mocks

import (
	"errors"
	"image"
	"io"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

type DiscordSessionMock struct {
	token      string
	shardID    int
	shardCount int
	isOpened   bool
	identify   discordgo.Identify
	state      *discordgo.State

	handlersMu   sync.RWMutex
	handlers     map[string][]*discordgo.EventHandler
	onceHandlers map[string][]*discordgo.EventHandler
}

func NewDiscordSession(token string) *DiscordSessionMock {
	s := &DiscordSessionMock{
		token:        token,
		shardID:      0,
		shardCount:   1,
		isOpened:     false,
		state:        discordgo.NewState(),
		handlers:     make(map[string][]*discordgo.EventHandler, 0),
		onceHandlers: make(map[string][]*discordgo.EventHandler, 0),
	}
	return s
}

func (s *DiscordSessionMock) Open() error {
	if s.isOpened {
		return errors.New("session is already open")
	}
	s.isOpened = true
	return nil
}

func (s *DiscordSessionMock) Close() error {
	return nil
}

func (s *DiscordSessionMock) ShardID() int {
	return s.shardID
}

func (s *DiscordSessionMock) State() *discordgo.State {
	return s.state
}

func (s *DiscordSessionMock) AddHandler(handler interface{}) func() {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) AddHandlerOnce(handler interface{}) func() {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) Channel(channelID string, options ...discordgo.RequestOption) (st *discordgo.Channel, err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) ChannelFileSend(channelID string, name string, r io.Reader, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) ChannelMessageDelete(channelID string, messageID string, options ...discordgo.RequestOption) (err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) ChannelMessageEdit(channelID string, messageID string, content string, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) ChannelMessageEditComplex(m *discordgo.MessageEdit, options ...discordgo.RequestOption) (st *discordgo.Message, err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) ChannelMessageEditEmbed(channelID string, messageID string, embed *discordgo.MessageEmbed, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) ChannelMessageEditEmbeds(channelID string, messageID string, embeds []*discordgo.MessageEmbed, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) ChannelMessagePin(channelID string, messageID string, options ...discordgo.RequestOption) (err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) ChannelMessageSend(channelID string, content string, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) ChannelMessageSendComplex(channelID string, data *discordgo.MessageSend, options ...discordgo.RequestOption) (st *discordgo.Message, err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) ChannelMessageSendEmbed(channelID string, embed *discordgo.MessageEmbed, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) ChannelMessageSendEmbedReply(channelID string, embed *discordgo.MessageEmbed, reference *discordgo.MessageReference, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) ChannelMessageSendEmbeds(channelID string, embeds []*discordgo.MessageEmbed, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) ChannelMessageSendEmbedsReply(channelID string, embeds []*discordgo.MessageEmbed, reference *discordgo.MessageReference, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) ChannelMessageSendReply(channelID string, content string, reference *discordgo.MessageReference, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) ChannelMessages(channelID string, limit int, beforeID string, afterID string, aroundID string, options ...discordgo.RequestOption) (st []*discordgo.Message, err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) ChannelMessagesBulkDelete(channelID string, messages []string, options ...discordgo.RequestOption) (err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) ChannelPermissionSet(channelID, targetID string, targetType discordgo.PermissionOverwriteType, allow, deny int64, options ...discordgo.RequestOption) (err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) ChannelTyping(channelID string, options ...discordgo.RequestOption) (err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) Guild(guildID string, options ...discordgo.RequestOption) (st *discordgo.Guild, err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) GuildBanCreate(guildID string, userID string, days int, options ...discordgo.RequestOption) (err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) GuildBanCreateWithReason(guildID string, userID string, reason string, days int, options ...discordgo.RequestOption) (err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) GuildBanDelete(guildID string, userID string, options ...discordgo.RequestOption) (err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) GuildBans(guildID string, limit int, beforeID string, afterID string, options ...discordgo.RequestOption) (st []*discordgo.GuildBan, err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) GuildChannels(guildID string, options ...discordgo.RequestOption) (st []*discordgo.Channel, err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) GuildIcon(guildID string, options ...discordgo.RequestOption) (img image.Image, err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) GuildMember(guildID string, userID string, options ...discordgo.RequestOption) (st *discordgo.Member, err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) GuildMemberAdd(guildID string, userID string, data *discordgo.GuildMemberAddParams, options ...discordgo.RequestOption) (err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) GuildMemberDelete(guildID string, userID string, options ...discordgo.RequestOption) (err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) GuildMemberDeleteWithReason(guildID string, userID string, reason string, options ...discordgo.RequestOption) (err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) GuildMemberRoleAdd(guildID string, userID string, roleID string, options ...discordgo.RequestOption) (err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) GuildMemberRoleRemove(guildID string, userID string, roleID string, options ...discordgo.RequestOption) (err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) GuildMemberTimeout(guildID string, userID string, until *time.Time, options ...discordgo.RequestOption) (err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) GuildMembers(guildID string, after string, limit int, options ...discordgo.RequestOption) (st []*discordgo.Member, err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) GuildRoleCreate(guildID string, data *discordgo.RoleParams, options ...discordgo.RequestOption) (st *discordgo.Role, err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) GuildRoleDelete(guildID string, roleID string, options ...discordgo.RequestOption) (err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) GuildRoleEdit(guildID string, roleID string, data *discordgo.RoleParams, options ...discordgo.RequestOption) (st *discordgo.Role, err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) GuildRoles(guildID string, options ...discordgo.RequestOption) (st []*discordgo.Role, err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) GuildSplash(guildID string, options ...discordgo.RequestOption) (img image.Image, err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) RequestGuildMembers(guildID string, query string, limit int, nonce string, presences bool) error {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) RequestGuildMembersBatch(guildIDs []string, query string, limit int, nonce string, presences bool) (err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) RequestGuildMembersBatchList(guildIDs []string, userIDs []string, limit int, nonce string, presences bool) (err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) RequestGuildMembersList(guildID string, userIDs []string, limit int, nonce string, presences bool) error {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) User(userID string, options ...discordgo.RequestOption) (st *discordgo.User, err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) UserChannelCreate(recipientID string, options ...discordgo.RequestOption) (st *discordgo.Channel, err error) {
	panic("not implemented") // TODO: Implement
}

func (s *DiscordSessionMock) UpdateStatusComplex(usd discordgo.UpdateStatusData) (err error) {
	panic("not implemented") // TODO: Implement
}
