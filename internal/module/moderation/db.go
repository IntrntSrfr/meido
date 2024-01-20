package moderation

import (
	"time"

	"github.com/intrntsrfr/meido/internal/database"
)

type IModerationDB interface {
	database.DB
	IFilterDB
	IWarnDB
}

type ModerationDB struct {
	database.DB
	IFilterDB
	IWarnDB
}

type IFilterDB interface {
	CreateGuildFilter(guildID, phrase string) error
	GetGuildFilterByPhrase(guildID, phrase string) (*Filter, error)
	GetGuildFilters(guildID string) ([]*Filter, error)
	DeleteGuildFilter(filterID int) error
	DeleteGuildFilters(guildID string) error
}

type FilterDB struct {
	database.DB
}

func (db *FilterDB) CreateGuildFilter(guildID, phrase string) error {
	_, err := db.GetConn().Exec("INSERT INTO filter VALUES (DEFAULT, $1, $2)", guildID, phrase)
	return err
}

func (db *FilterDB) GetGuildFilterByPhrase(guildID, phrase string) (*Filter, error) {
	var filter Filter
	err := db.GetConn().Get(&filter, "SELECT * FROM filter WHERE guild_id = $1 AND phrase = $2", guildID, phrase)
	return &filter, err
}

func (db *FilterDB) GetGuildFilters(guildID string) ([]*Filter, error) {
	var filters []*Filter
	err := db.GetConn().Select(&filters, "SELECT * FROM filter WHERE guild_id=$1", guildID)
	return filters, err
}
func (db *FilterDB) DeleteGuildFilter(filterID int) error {
	_, err := db.GetConn().Exec("DELETE FROM filter WHERE uid=$1", filterID)
	return err
}

func (db *FilterDB) DeleteGuildFilters(guildID string) error {
	_, err := db.GetConn().Exec("DELETE FROM filter WHERE guild_id=$1", guildID)
	return err
}

type IWarnDB interface {
	CreateMemberWarn(guildID, userID, reason, authorID string) error
	GetGuildWarns(guildID string) ([]*Warn, error)
	GetGuildWarnsIfActive(guildID string) ([]*Warn, error)
	ClearActiveUserWarns(guildID, userID, clearedByID string) error
	GetMemberWarns(guildID, userID string) ([]*Warn, error)
	GetMemberWarnsIfActive(guildID, userID string) ([]*Warn, error)
	UpdateMemberWarn(warn *Warn) error
}

type WarnDB struct {
	database.DB
}

func (db *WarnDB) CreateMemberWarn(guildID, userID, reason, authorID string) error {
	_, err := db.GetConn().Exec("INSERT INTO warn VALUES(DEFAULT, $1, $2, $3, $4, $5, $6)",
		guildID, userID, reason, authorID, time.Now(), true)
	return err
}

func (db *WarnDB) GetGuildWarnsIfActive(guildID string) ([]*Warn, error) {
	var warns []*Warn
	err := db.GetConn().Select(&warns, "SELECT * FROM warn WHERE guild_id=$1 AND is_valid ORDER BY given_at DESC", guildID)
	return warns, err
}

func (db *WarnDB) ClearActiveUserWarns(guildID, userID, clearedByID string) error {
	_, err := db.GetConn().Exec("UPDATE warn SET is_valid=false, cleared_by_id=$1, cleared_at=$2 WHERE guild_id=$3 AND user_id=$4 and is_valid",
		clearedByID, time.Now(), guildID, userID)
	return err
}

func (db *WarnDB) GetGuildWarns(guildID string) ([]*Warn, error) {
	var warns []*Warn
	err := db.GetConn().Select(&warns, "SELECT * FROM warn WHERE guild_id=$1 ORDER BY given_at DESC", guildID)
	return warns, err
}

func (db *WarnDB) GetMemberWarns(guildID, userID string) ([]*Warn, error) {
	var warns []*Warn
	err := db.GetConn().Select(&warns, "SELECT * FROM warn WHERE guild_id=$1 AND user_id=$2 ORDER BY given_at DESC", guildID, userID)
	return warns, err
}

func (db *WarnDB) GetMemberWarnsIfActive(guildID, userID string) ([]*Warn, error) {
	var warns []*Warn
	err := db.GetConn().Select(&warns, "SELECT * FROM warn WHERE guild_id=$1 AND user_id=$2 AND is_valid ORDER BY given_at DESC", guildID, userID)
	return warns, err
}

func (db *WarnDB) UpdateMemberWarn(warn *Warn) error {
	_, err := db.GetConn().Exec("UPDATE warn SET is_valid=false, cleared_by_id=$1, cleared_at=$2 WHERE uid = $3",
		warn.ClearedByID, warn.ClearedAt, warn.UID)
	return err
}
