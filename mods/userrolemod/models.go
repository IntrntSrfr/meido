package userrolemod

// Userrole represents a user role.
type Userrole struct {
	UID     int
	GuildID string `db:"guild_id"`
	RoleID  string `db:"role_id"`
	UserID  string `db:"user_id"`
}
