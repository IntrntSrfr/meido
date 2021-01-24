package meido

import (
	"github.com/bwmarrin/discordgo"
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
var PermMap = map[int]string{
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
	Sess    *discordgo.Session
	Discord *Discord
	Message *discordgo.Message

	// Partial guild member, use only for guild related stuff
	Author       *discordgo.User
	Member       *discordgo.Member
	Type         MessageType
	TimeReceived time.Time
	Shard        int
}

// Reply replies directly to a DiscordMessage
func (m *DiscordMessage) Reply(data string) (*discordgo.Message, error) {
	return m.Sess.ChannelMessageSend(m.Message.ChannelID, data)
}

// ReplyEmbed replies directly to a DiscordMessage with an embed.
func (m *DiscordMessage) ReplyEmbed(embed *discordgo.MessageEmbed) (*discordgo.Message, error) {
	return m.Sess.ChannelMessageSendEmbed(m.Message.ChannelID, embed)
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

// IsDM returns whether or not the message is a direct message.
func (m *DiscordMessage) IsDM() bool {
	return m.Message.Type == discordgo.MessageTypeDefault && m.Message.GuildID == ""
}

// IsOwner returns whether the author of the message is a bot owner.
func (m *DiscordMessage) IsOwner() bool {
	return m.Discord.IsOwner(m)
}

// HasPermissions returns if a member has certain permissions or not.
func (m *DiscordMessage) HasPermissions(mem *discordgo.Member, channelID string, perm int) bool {
	uPerms, err := m.Discord.UserChannelPermissionsDirect(mem, channelID)
	if err != nil {
		return false
	}
	return uPerms&perm != 0 || uPerms&discordgo.PermissionAdministrator != 0
}
