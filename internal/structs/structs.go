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

// CustomRole represents a user role.
type CustomRole struct {
	UID     int    `db:"uid"`
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
	GuildID  string `db:"guild_id"`
	UseWarns bool   `db:"use_warns"`
	MaxWarns int    `db:"max_warns"`

	// described in days, 0 means infinite duration
	WarnDuration        int    `db:"warn_duration"`
	AutomodLogChannelID string `db:"automod_log_channel_id"`
	FishingChannelID    string `db:"fishing_channel_id"`
	AutoRoleID          string `db:"auto_role_id"`
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

type CreatureRarity struct {
	UID         int    `db:"uid"`
	Name        string `db:"name"`
	DisplayName string `db:"display_name"`
	Weight      int    `db:"weight"`
}

type Creature struct {
	UID         int             `db:"uid"`
	Name        string          `db:"name"`
	DisplayName string          `db:"display_name"`
	RarityID    int             `db:"rarity_id"`
	Rarity      *CreatureRarity `db:"-"`
}
