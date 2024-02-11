package discord

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/pkg/utils"
)

// MessageType represents the 3 types of message events from Discord.
type MessageType int

// MessageType codes.
const (
	MessageTypeCreate MessageType = 1 << iota
	MessageTypeUpdate
	MessageTypeDelete
)

var (
	ErrNotAllowed = errors.New("not allowed, missing permissions")
)

// PermMap is a map that simply converts specific permission bits to the readable version of what they represent.
var PermMap = map[int64]string{
	0:                                         "None",
	discordgo.PermissionSendMessages:          "Send Messages",
	discordgo.PermissionSendTTSMessages:       "Send TTS Messages",
	discordgo.PermissionManageMessages:        "Manage Messages",
	discordgo.PermissionEmbedLinks:            "Embed Links",
	discordgo.PermissionAttachFiles:           "Attach Files",
	discordgo.PermissionReadMessageHistory:    "Read Message History",
	discordgo.PermissionMentionEveryone:       "Mention Everyone",
	discordgo.PermissionUseExternalEmojis:     "Use External Emojis",
	discordgo.PermissionUseSlashCommands:      "Use Slash Commands",
	discordgo.PermissionManageThreads:         "Manage Threads",
	discordgo.PermissionCreatePublicThreads:   "Create Public Threads",
	discordgo.PermissionCreatePrivateThreads:  "Create Private Threads",
	discordgo.PermissionUseExternalStickers:   "Use External Stickers",
	discordgo.PermissionSendMessagesInThreads: "Send Messages In Threads",
	discordgo.PermissionVoicePrioritySpeaker:  "Priority Speaker",
	discordgo.PermissionVoiceStreamVideo:      "Stream Video",
	discordgo.PermissionVoiceConnect:          "Connect",
	discordgo.PermissionVoiceSpeak:            "Speak",
	discordgo.PermissionVoiceMuteMembers:      "Mute Members",
	discordgo.PermissionVoiceDeafenMembers:    "Deafen Members",
	discordgo.PermissionVoiceMoveMembers:      "Move Members",
	discordgo.PermissionVoiceUseVAD:           "Use VAD",
	discordgo.PermissionVoiceRequestToSpeak:   "Request To Speak",
	discordgo.PermissionUseActivities:         "Use Activities",
	discordgo.PermissionChangeNickname:        "Change Nickname",
	discordgo.PermissionManageNicknames:       "Manage Nicknames",
	discordgo.PermissionManageRoles:           "Manage Roles",
	discordgo.PermissionManageWebhooks:        "Manage Webhooks",
	discordgo.PermissionManageEmojis:          "Manage Emojis",
	discordgo.PermissionManageEvents:          "Manage Events",
	discordgo.PermissionCreateInstantInvite:   "Create Instant Invite",
	discordgo.PermissionKickMembers:           "Kick Members",
	discordgo.PermissionBanMembers:            "Ban Members",
	discordgo.PermissionAdministrator:         "Administrator",
	discordgo.PermissionManageChannels:        "Manage Channels",
	discordgo.PermissionManageServer:          "Manage Server",
	discordgo.PermissionAddReactions:          "Add Reactions",
	discordgo.PermissionViewAuditLogs:         "View Audit Logs",
	discordgo.PermissionViewChannel:           "View Channel",
	discordgo.PermissionViewGuildInsights:     "View Guild Insights",
	discordgo.PermissionModerateMembers:       "Moderate Members",
}

// DiscordMessage represents a Discord message sent in a channel, and
// contains fields so that it is easy to work with the data it gives.
type DiscordMessage struct {
	Sess         DiscordSession `json:"-"`
	Discord      *Discord       `json:"-"`
	Message      *discordgo.Message
	MessageType  MessageType
	TimeReceived time.Time
	Shard        int
}

// Reply replies directly to a DiscordMessage
func (m *DiscordMessage) Reply(data string) (*discordgo.Message, error) {
	return m.Sess.ChannelMessageSendComplex(m.ChannelID(), &discordgo.MessageSend{
		Content:         data,
		AllowedMentions: &discordgo.MessageAllowedMentions{},
		Reference: &discordgo.MessageReference{
			MessageID: m.ID(),
			ChannelID: m.ChannelID(),
			GuildID:   m.GuildID(),
		},
	})
}

// ReplyAndDelete sends a message to a channel, then deletes it after a duration d
func (m *DiscordMessage) ReplyAndDelete(data string, d time.Duration) (*discordgo.Message, error) {
	r, err := m.Reply(data)
	if err != nil {
		return nil, err
	}
	go func() {
		time.AfterFunc(d, func() {
			_ = m.Sess.ChannelMessageDelete(m.ChannelID(), r.ID)
		})
	}()
	return r, nil
}

// ReplyEmbed replies directly to a DiscordMessage with an embed.
func (m *DiscordMessage) ReplyEmbed(embed *discordgo.MessageEmbed) (*discordgo.Message, error) {
	return m.Sess.ChannelMessageSendComplex(m.ChannelID(), &discordgo.MessageSend{
		Embed:           embed,
		AllowedMentions: &discordgo.MessageAllowedMentions{},
		Reference: &discordgo.MessageReference{
			MessageID: m.ID(),
			ChannelID: m.ChannelID(),
			GuildID:   m.GuildID(),
		},
	})
}

func (m *DiscordMessage) ReplyComplex(data *discordgo.MessageSend) (*discordgo.Message, error) {
	return m.Sess.ChannelMessageSendComplex(m.ChannelID(), data)
}

func (m *DiscordMessage) Type() MessageType {
	return m.MessageType
}

func (m *DiscordMessage) Delete() error {
	return m.Sess.ChannelMessageDelete(m.ChannelID(), m.ID())
}

// Args returns the split content of a DiscordMessage in lowercase.
func (m *DiscordMessage) Args() []string {
	return strings.Fields(strings.ToLower(m.Message.Content))
}

// RawArgs returns the raw split content of a DiscordMessage.
func (m *DiscordMessage) RawArgs() []string {
	return strings.Fields(m.Message.Content)
}

// RawContent returns the raw content of a DiscordMessage.
func (m *DiscordMessage) RawContent() string {
	return m.Message.Content
}

// IsDM returns whether the message is a direct message.
func (m *DiscordMessage) IsDM() bool {
	return m.Message.Type == discordgo.MessageTypeDefault && m.Message.GuildID == ""
}

// AuthorHasPermissions returns if a member has certain permissions or not.
func (m *DiscordMessage) AuthorHasPermissions(perm int64) (bool, error) {
	uPerms, err := m.Sess.State().MessagePermissions(m.Message)
	if err != nil {
		return false, err
	}
	return uPerms&perm != 0 || uPerms&discordgo.PermissionAdministrator != 0, nil
}

func (m *DiscordMessage) CallbackKey() string {
	return fmt.Sprintf("%v:%v", m.ChannelID(), m.AuthorID())
}

// BotHasPermissions returns if a member has certain permissions or not.
func (m *DiscordMessage) BotHasPermissions(perm int64) (bool, error) {
	return m.Discord.BotHasPermissions(m.ChannelID(), perm)
}

func (m *DiscordMessage) IsBot() bool {
	return m.Message != nil && m.Author() != nil && m.Author().Bot
}

func (m *DiscordMessage) Author() *discordgo.User {
	return m.Message.Author
}

func (m *DiscordMessage) Member() *discordgo.Member {
	return m.Message.Member
}

func (m *DiscordMessage) ID() string {
	return m.Message.ID
}

func (m *DiscordMessage) AuthorID() string {
	if m.Message.Author == nil {
		return ""
	}
	return m.Message.Author.ID
}

func (m *DiscordMessage) GuildID() string {
	return m.Message.GuildID
}

func (m *DiscordMessage) ChannelID() string {
	return m.Message.ChannelID
}

func (m *DiscordMessage) Mentions() []*discordgo.User {
	return m.Message.Mentions
}

func (m *DiscordMessage) MentionRoles() []string {
	return m.Message.MentionRoles
}

func (m *DiscordMessage) Attachments() []*discordgo.MessageAttachment {
	return m.Message.Attachments
}

func (m *DiscordMessage) StartTyping() {
	m.Sess.ChannelTyping(m.ChannelID())
}

func (m *DiscordMessage) Ban(userID, reason string, days int) error {
	allowed, err := m.AuthorHasPermissions(discordgo.PermissionBanMembers)
	if err != nil {
		return err
	}
	botAllowed, err := m.BotHasPermissions(discordgo.PermissionBanMembers)
	if err != nil {
		return err
	}

	if !allowed || !botAllowed {
		return ErrNotAllowed
	}

	return m.Sess.GuildBanCreate(m.GuildID(), userID, days, discordgo.WithAuditLogReason(reason))
}

func (m *DiscordMessage) Unban(userID string) error {
	allowed, err := m.AuthorHasPermissions(discordgo.PermissionBanMembers)
	if err != nil {
		return err
	}
	botAllowed, err := m.BotHasPermissions(discordgo.PermissionBanMembers)
	if err != nil {
		return err
	}

	if !allowed || !botAllowed {
		return ErrNotAllowed
	}
	return m.Sess.GuildBanDelete(m.GuildID(), userID)
}

func (m *DiscordMessage) MemberRoleRemove(userID, roleID string) error {
	allowed, err := m.AuthorHasPermissions(discordgo.PermissionManageRoles)
	if err != nil {
		return err
	}
	botAllowed, err := m.BotHasPermissions(discordgo.PermissionManageRoles)
	if err != nil {
		return err
	}

	if !allowed || !botAllowed {
		return ErrNotAllowed
	}

	return m.Sess.GuildMemberRoleRemove(m.GuildID(), userID, roleID)
}

func (m *DiscordMessage) MemberRoleAdd(userID, roleID string) error {
	allowed, err := m.AuthorHasPermissions(discordgo.PermissionManageRoles)
	if err != nil {
		return err
	}
	botAllowed, err := m.BotHasPermissions(discordgo.PermissionManageRoles)
	if err != nil {
		return err
	}

	if !allowed || !botAllowed {
		return ErrNotAllowed
	}

	return m.Sess.GuildMemberRoleAdd(m.GuildID(), userID, roleID)
}

func (m *DiscordMessage) GetMemberAtArg(index int) (*discordgo.Member, error) {
	if len(m.Args()) <= index {
		return nil, errors.New("index out of range")
	}
	str := m.Args()[index]
	userID := utils.TrimUserID(str)
	if !utils.IsNumber(userID) {
		return nil, fmt.Errorf("%s could not be parsed as a number", userID)
	}
	return m.Discord.Member(m.GuildID(), userID)
}

func (m *DiscordMessage) GetUserAtArg(index int) (*discordgo.User, error) {
	if len(m.Args()) <= index {
		return nil, errors.New("index out of range")
	}
	str := m.Args()[index]
	userID := utils.TrimUserID(str)
	if !utils.IsNumber(userID) {
		return nil, fmt.Errorf("%s could not be parsed as a number", userID)
	}
	return m.Sess.User(userID)
}

func (m *DiscordMessage) GetMemberOrUserAtArg(index int) (*discordgo.User, error) {
	member, err := m.GetMemberAtArg(index)
	if err != nil {
		return m.GetUserAtArg(index)
	}
	return member.User, nil
}

// TargetRoleIsLowest compares the bot user, author, and a target member, and returns whether the
// targetMember is below both the bot and author in the role hierarchy
func (m *DiscordMessage) TargetRoleIsLowest(targetMember *discordgo.Member) bool {
	topUserRole := m.Discord.HighestRolePosition(m.GuildID(), m.AuthorID())
	topTargetRole := m.Discord.HighestRolePosition(m.GuildID(), targetMember.User.ID)
	topBotRole := m.Discord.HighestRolePosition(m.GuildID(), m.Sess.State().User.ID)
	return topUserRole > topTargetRole && topBotRole > topTargetRole
}
