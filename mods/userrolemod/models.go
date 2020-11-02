package userrolemod

type Userrole struct {
	UID     int
	GuildID int64 `db:"guild_id"`
	RoleID  int64 `db:"role_id"`
	UserID  int64 `db:"user_id"`
}
