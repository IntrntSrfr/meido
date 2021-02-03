package moderationmod

import "time"

// FilterEntry represents a filtered phrase that the bot should look out for
type FilterEntry struct {
	UID     int    `db:"uid"`
	GuildID string `db:"guild_id"`
	Phrase  string `db:"phrase"`
}

// DiscordGuild represents a servers settings.
type DiscordGuild struct {
	UID      int    `db:"uid"`
	GuildID  string `db:"guild_id"`
	UseWarns bool   `db:"use_warns"`
	MaxWarns int    `db:"max_warns"`

	// described in days, 0 means infinite duration
	WarnDuration int `db:"warn_duration"`
}

// WarnEntry represents a warning
type WarnEntry struct {
	UID         int        `db:"uid"`
	GuildID     string     `db:"guild_id"`
	UserID      string     `db:"user_id"`
	Reason      string     `db:"reason"`
	GivenByID   string     `db:"given_by_id"`
	GivenAt     time.Time  `db:"given_at"`
	IsValid     bool       `db:"is_valid"`
	ClearedByID *string    `db:"cleared_by_id"`
	ClearedAt   *time.Time `db:"cleared_at"`
}
