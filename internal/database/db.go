package database

import "github.com/jmoiron/sqlx"

type DB struct {
	*sqlx.DB
}

func New(db *sqlx.DB) (*DB, error) {
	return &DB{db}, nil
}

func (db *DB) Populate() {
	// run all the schemas for the db startup here
	// as well as migrations i suppose
}

func (db *DB) Guild(guildID, userID string) (*DiscordGuild, error) {
	var guild *DiscordGuild
	err := db.Get(&guild, "SELECT COUNT(*) FROM warns WHERE guild_id=$1 AND user_id=$2 AND is_valid", guildID, userID)
	return guild, err
}

func (db *DB) UserRole(guildID, userID string) (*Userrole, error) {
	var ur *Userrole
	err := db.Get(ur, "SELECT * FROM userroles WHERE guild_id=$1 AND user_id=$2", guildID, userID)
	return ur, err
}

func (db *DB) WarnCount(guildID, userID string) (int, error) {
	var count int
	err := db.Get(&count, "SELECT COUNT(*) FROM warns WHERE guild_id=$1 AND user_id=$2 AND is_valid", guildID, userID)
	return count, err
}

func (db *DB) CommandCount() (int, error) {
	var count int
	err := db.Get(&count, "SELECT COUNT(*) FROM commandlog;")
	return count, err
}
