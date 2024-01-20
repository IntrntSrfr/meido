package structs

import "time"

// CommandLogEntry represents an entry in the command log
type CommandLogEntry struct {
	UID       int       `db:"uid"`
	Command   string    `db:"command"`
	Args      string    `db:"args"`
	UserID    string    `db:"user_id"`
	GuildID   string    `db:"guild_id"`
	ChannelID string    `db:"channel_id"`
	MessageID string    `db:"message_id"`
	SentAt    time.Time `db:"sent_at"`
}

// Guild represents a server and its information.
type Guild struct {
	GuildID  string `db:"guild_id"`
	UseWarns bool   `db:"use_warns"`
	MaxWarns int    `db:"max_warns"`

	// described in days, 0 means infinite duration
	WarnDuration        int    `db:"warn_duration"`
	AutomodLogChannelID string `db:"automod_log_channel_id"`
	FishingChannelID    string `db:"fishing_channel_id"`
	AutoRoleID          string `db:"auto_role_id"`
}
