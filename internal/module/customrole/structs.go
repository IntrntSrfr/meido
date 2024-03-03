package customrole

type CustomRole struct {
	UID     int    `db:"uid"`
	GuildID string `db:"guild_id"`
	RoleID  string `db:"role_id"`
	UserID  string `db:"user_id"`
}
