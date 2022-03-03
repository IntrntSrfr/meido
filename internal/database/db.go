package database

import (
	"github.com/jmoiron/sqlx"
	"time"
)

// this should probably be made to an interface

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

func (db *DB) GetGuild(guildID string) (*Guild, error) {
	guild := &Guild{}
	err := db.Get(&guild, "SELECT * FROM guild WHERE guild_id=$1", guildID)
	return guild, err
}

func (db *DB) GetGuildAutomodLogChannel(guildID string) string {
	var ch string
	_ = db.Get(&ch, "SELECT automod_log_channel FROM guild WHERE guild_id = $1", guildID)
	return ch
}

func (db *DB) GetGuildFishingChannel(guildID string) string {
	var ch string
	_ = db.Get(&ch, "SELECT fishing_channel FROM guild WHERE guild_id = $1", guildID)
	return ch
}

func (db *DB) GetUserRole(guildID, userID string) (*UserRole, error) {
	ur := &UserRole{}
	err := db.Get(ur, "SELECT * FROM user_role WHERE guild_id=$1 AND user_id=$2", guildID, userID)
	return ur, err
}

func (db *DB) GetValidUserWarnCount(guildID, userID string) (int, error) {
	var count int
	err := db.Get(&count, "SELECT COUNT(*) FROM warn WHERE guild_id=$1 AND user_id=$2 AND is_valid", guildID, userID)
	return count, err
}

func (db *DB) GetCommandCount() (int, error) {
	var count int
	err := db.Get(&count, "SELECT COUNT(*) FROM command_log;")
	return count, err
}

func (db *DB) GetFilter(guildID, phrase string) (*Filter, error) {
	f := &Filter{}
	err := db.Get(f, "SELECT * FROM filter WHERE guild_id = $1 AND phrase = $2", guildID, phrase)
	return f, err
}

func (db *DB) GetGuildFilters(guildID string) ([]*Filter, error) {
	var filters []*Filter
	err := db.Select(&filters, "SELECT * FROM filter WHERE guild_id=$1", guildID)
	return filters, err
}

func (db *DB) DeleteGuildFilters(guildID string) {
	_, _ = db.Exec("DELETE FROM filter WHERE guild_id=$1", guildID)
}

func (db *DB) InsertWarn(guildID, userID, reason, givenByID string) error {
	_, err := db.Exec("INSERT INTO warn VALUES(DEFAULT, $1, $2, $3, $4, $5, $6)",
		guildID, userID, reason, givenByID, time.Now(), true)
	return err
}

func (db *DB) ClearActiveUserWarns(guildID, userID, clearedByID string) error {
	_, err := db.Exec("UPDATE warn SET is_valid=false, cleared_by_id=$1, cleared_at=$2 WHERE guild_id=$3 AND user_id=$4 and is_valid",
		clearedByID, time.Now(), guildID, userID)
	return err
}

func (db *DB) InsertNewAquarium(userID string) (*Aquarium, error) {
	aq := &Aquarium{UserID: userID}
	_, err := db.Exec("INSERT INTO aquarium VALUES ($1)", userID)
	return aq, err
}

func (db *DB) GetAquarium(userID string) (*Aquarium, error) {
	aq := new(Aquarium)
	err := db.Get(aq, "SELECT * FROM aquarium WHERE user_id=$1", userID)
	return aq, err
}

func (db *DB) UpdateAquarium(aq *Aquarium) error {
	_, err := db.Exec("UPDATE aquarium SET common=$1, uncommon=$2, rare=$3, super_rare=$4, legendary=$5 WHERE user_id=$6",
		aq.Common, aq.Uncommon, aq.Rare, aq.SuperRare, aq.Legendary, aq.UserID)
	return err
}
