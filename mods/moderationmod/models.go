package moderationmod

type FilterEntry struct {
	UID     int    `db:"uid"`
	GuildID int64  `db:"guild_id"`
	Phrase  string `db:"phrase"`
}
