package database

import (
	"github.com/jmoiron/sqlx"
	"time"
)

type DB struct {
	*sqlx.DB
}

func New(db *sqlx.DB) *DB {
	return &DB{db}
}

func (db *DB) Populate() {
	// run all the schemas for the db startup here
	// as well as migrations I suppose
}

func (db *DB) Guild(guildID string) (*DiscordGuild, error) {
	var guild *DiscordGuild
	err := db.Get(&guild, "SELECT * FROM guild WHERE guild_id=$1", guildID)
	return guild, err
}

func (db *DB) UserRole(guildID, userID string) (*Userrole, error) {
	var ur *Userrole
	err := db.Get(ur, "SELECT * FROM user_role WHERE guild_id=$1 AND user_id=$2", guildID, userID)
	return ur, err
}

func (db *DB) GetValidWarnCount(guildID, userID string) (int, error) {
	var count int
	err := db.Get(&count, "SELECT COUNT(*) FROM warns WHERE guild_id=$1 AND user_id=$2 AND is_valid", guildID, userID)
	return count, err
}

func (db *DB) CommandCount() (int, error) {
	var count int
	err := db.Get(&count, "SELECT COUNT(*) FROM command_log;")
	return count, err
}

func (db *DB) GetFilter(guildID, phrase string) (*FilterEntry, error) {
	var f *FilterEntry
	err := db.Get(f, "SELECT * FROM filter WHERE guild_id = $1 AND phrase = $2", guildID, phrase)
	return f, err
}

func (db *DB) GetGuildFilters(guildID string) ([]*FilterEntry, error) {
	var filters []*FilterEntry
	err := db.Select(&filters, "SELECT * FROM filter WHERE guild_id=$1", guildID)
	return filters, err
}

func (db *DB) DeleteGuildFilters(guildID string) {
	_, _ = db.Exec("DELETE FROM filters WHERE guild_id=$1", guildID)
}

func (db *DB) InsertWarn(guildID, userID, reason, givenByID string) error {
	_, err := db.Exec("INSERT INTO warns VALUES(DEFAULT, $1, $2, $3, $4, $5, $6)",
		guildID, userID, reason, givenByID, time.Now(), true)
	return err
}

func (db *DB) ClearActiveWarns(guildID, userID, clearedByID string) error {
	_, err := db.Exec("UPDATE warns SET is_valid=false, cleared_by_id=$1, cleared_at=$2 WHERE guild_id=$3 AND user_id=$4 and is_valid",
		clearedByID, time.Now(), guildID, userID)
	return err
}
