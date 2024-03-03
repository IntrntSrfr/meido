package moderation

import "time"

// Warn represents a warning
type Warn struct {
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

// Filter represents a filtered phrase that the bot should look out for
type Filter struct {
	UID     int    `db:"uid"`
	GuildID string `db:"guild_id"`
	Phrase  string `db:"phrase"`
}
