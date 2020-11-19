package moderationmod

import "time"

type FilterEntry struct {
	UID     int    `db:"uid"`
	GuildID string `db:"guild_id"`
	Phrase  string `db:"phrase"`
}

type DiscordGuild struct {
	UID        int    `db:"uid"`
	GuildID    string `db:"guild_id"`
	UseStrikes bool   `db:"use_strikes"`
	MaxStrikes int    `db:"max_strikes"`
}

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
