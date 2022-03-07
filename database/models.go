package database

import "time"

// UserRole represents a user role.
type UserRole struct {
	UID     int
	GuildID string `db:"guild_id"`
	RoleID  string `db:"role_id"`
	UserID  string `db:"user_id"`
}

// Filter represents a filtered phrase that the bot should look out for
type Filter struct {
	UID     int    `db:"uid"`
	GuildID string `db:"guild_id"`
	Phrase  string `db:"phrase"`
}

// Guild represents a server and its information.
type Guild struct {
	UID      int    `db:"uid"`
	GuildID  string `db:"guild_id"`
	UseWarns bool   `db:"use_warns"`
	MaxWarns int    `db:"max_warns"`

	// described in days, 0 means infinite duration
	WarnDuration int `db:"warn_duration"`
	//AutoRole     string `db:"autorole"`

	AutomodLogChannel string `db:"automod_log_channel"`
	FishingChannel    string `db:"fishing_channel"`
}

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

// Aquarium represents a users aquarium
type Aquarium struct {
	UserID    string `db:"user_id"`
	Common    int    `db:"common"`
	Uncommon  int    `db:"uncommon"`
	Rare      int    `db:"rare"`
	SuperRare int    `db:"super_rare"`
	Legendary int    `db:"legendary"`
}

// AutoRole represents a users aquarium
type AutoRole struct {
	ID      int    `db:"id"`
	GuildID string `db:"guild_id"`
	RoleID  string `db:"role_id"`
	Enabled bool   `db:"enabled"`
}