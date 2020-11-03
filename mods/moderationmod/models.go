package moderationmod

import "time"

type FilterEntry struct {
	UID     int    `db:"uid"`
	GuildID int64  `db:"guild_id"`
	Phrase  string `db:"phrase"`
}

type DiscordGuild struct {
	UID        int   `db:"uid"`
	GuildID    int64 `db:"guild_id"`
	UseStrikes bool  `db:"use_strikes"`
	MaxStrikes int   `db:"max_strikes"`
}

type WarnEntry struct {
	UID         int        `db:"uid"`
	GuildID     int64      `db:"guild_id"`
	UserID      int64      `db:"user_id"`
	Reason      string     `db:"reason"`
	GivenByID   int64      `db:"given_by_id"`
	GivenAt     time.Time  `db:"given_at"`
	IsValid     bool       `db:"is_valid"`
	ClearedByID *int64     `db:"cleared_by_id"`
	ClearedAt   *time.Time `db:"cleared_at"`
}
