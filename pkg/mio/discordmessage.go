package mio

import (
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/pkg/utils"
	"strings"
	"time"
)

// MessageType represents the 3 types of message events from Discord.
type MessageType int

// MessageType codes.
const (
	MessageTypeCreate MessageType = 1 << iota
	MessageTypeUpdate
	MessageTypeDelete
)

// PermMap is a map that simply converts specific permission bits to the readable version of what they represent.
var PermMap = map[int64]string{
	0:                                        "None",
	discordgo.PermissionCreateInstantInvite:  "Create Instant Invite",
	discordgo.PermissionKickMembers:          "Kick Members",
	discordgo.PermissionBanMembers:           "Ban Members",
	discordgo.PermissionAdministrator:        "Administrator",
	discordgo.PermissionManageChannels:       "Manage Channels",
	discordgo.PermissionManageServer:         "Manage Server",
	discordgo.PermissionAddReactions:         "Add Reactions",
	discordgo.PermissionViewAuditLogs:        "View Audit Log",
	discordgo.PermissionVoicePrioritySpeaker: "Priority Speaker",
	discordgo.PermissionViewChannel:          "View Channel",
	discordgo.PermissionSendMessages:         "Send Messages",
	discordgo.PermissionSendTTSMessages:      "Send TTS Messages",
	discordgo.PermissionManageMessages:       "Manage Messages",
	discordgo.PermissionEmbedLinks:           "Embed Links",
	discordgo.PermissionAttachFiles:          "Attach Files",
	discordgo.PermissionReadMessageHistory:   "Read Message History",
	discordgo.PermissionMentionEveryone:      "Mention Everyone",
	discordgo.PermissionUseExternalEmojis:    "Use External Emojis",
	discordgo.PermissionVoiceConnect:         "Connect",
	discordgo.PermissionVoiceSpeak:           "Speak",
	discordgo.PermissionVoiceMuteMembers:     "Mute Members",
	discordgo.PermissionVoiceDeafenMembers:   "Deafen Members",
	discordgo.PermissionVoiceMoveMembers:     "Move Members",
	discordgo.PermissionVoiceUseVAD:          "Use VAD",
	discordgo.PermissionChangeNickname:       "Change Nickname",
	discordgo.PermissionManageNicknames:      "Manage Nicknames",
	discordgo.PermissionManageRoles:          "Manage Roles",
	discordgo.PermissionManageWebhooks:       "Manage Webhooks",
	discordgo.PermissionManageEmojis:         "Manage Emojis",
}

// DiscordMessage represents a Discord message sent in a channel, and contains fields so that it is easy to
// work with the data it gives.
type DiscordMessage struct {
	Sess         *discordgo.Session `json:"-"`
	Discord      *Discord           `json:"-"`
	Message      *discordgo.Message
	MessageType  MessageType
	TimeReceived time.Time
	Shard        int
}

// Reply replies directly to a DiscordMessage
func (m *DiscordMessage) Reply(data string) (*discordgo.Message, error) {
	return m.Sess.ChannelMessageSend(m.Message.ChannelID, data)
}

// ReplyAndDelete sends a message to a channel, then deletes it after a duration d
func (m *DiscordMessage) ReplyAndDelete(data string, d time.Duration) (*discordgo.Message, error) {
	r, err := m.Sess.ChannelMessageSend(m.Message.ChannelID, data)
	if err != nil {
		return nil, err
	}
	go func() {
		time.AfterFunc(d, func() {
			_ = m.Sess.ChannelMessageDelete(m.Message.ChannelID, r.ID)
		})
	}()
	return r, nil
}

// ReplyEmbed replies directly to a DiscordMessage with an embed.
func (m *DiscordMessage) ReplyEmbed(embed *discordgo.MessageEmbed) (*discordgo.Message, error) {
	return m.Sess.ChannelMessageSendEmbed(m.Message.ChannelID, embed)
}

func (m *DiscordMessage) Type() MessageType {
	return m.MessageType
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

// LenArgs returns the length of Args
func (m *DiscordMessage) LenArgs() int {
	return len(m.Args())
}

// IsDM returns whether the message is a direct message.
func (m *DiscordMessage) IsDM() bool {
	return m.Message.Type == discordgo.MessageTypeDefault && m.Message.GuildID == ""
}

// HasPermissions returns if a member has certain permissions or not.
func (m *DiscordMessage) HasPermissions(perm int64) (bool, error) {
	uPerms, err := m.Sess.State.MessagePermissions(m.Message)
	//uPerms, err := m.Discord.UserChannelPermissionsDirect(mem, channelID)
	if err != nil {
		return false, err
	}
	return uPerms&perm != 0 || uPerms&discordgo.PermissionAdministrator != 0, nil
}

func (m *DiscordMessage) IsBot() bool {
	return m.Message.Author.Bot
}

func (m *DiscordMessage) Author() *discordgo.User {
	return m.Message.Author
}

func (m *DiscordMessage) Member() *discordgo.Member {
	return m.Message.Member
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

func (m *DiscordMessage) GetMemberAtArg(index int) (*discordgo.Member, error) {
	if len(m.Args()) <= index {
		return nil, errors.New("index out of range")
	}
	str := m.Args()[index]
	userID := utils.TrimUserID(str)
	if !utils.IsNumber(userID) {
		return nil, errors.New(fmt.Sprintf("%s could not be parsed as a number", userID))
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
		return nil, errors.New(fmt.Sprintf("%s could not be parsed as a number", userID))
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
