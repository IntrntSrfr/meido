package database

import (
	"github.com/intrntsrfr/meido/internal/structs"
	"github.com/jmoiron/sqlx"

	_ "github.com/lib/pq"
)

type PsqlDB struct {
	pool    *sqlx.DB
	connStr string
	IGuildDB
	ICommandLogDB
}

func NewPSQLDatabase(connStr string) (*PsqlDB, error) {
	pool, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return nil, err
	}

	db := &PsqlDB{
		pool:    pool,
		connStr: connStr,
	}
	db.IGuildDB = &GuildDB{db}
	db.ICommandLogDB = &CommandLogDB{db}
	return db, nil
}

func (p *PsqlDB) GetConn() *sqlx.DB {
	return p.pool
}

func (p *PsqlDB) Close() error {
	return p.pool.Close()
}

type CommandLogDB struct {
	DB
}

func (db *CommandLogDB) CreateCommandLogEntry(e *structs.CommandLogEntry) error {
	_, err := db.GetConn().Exec("INSERT INTO command_log VALUES(DEFAULT, $1, $2, $3, $4, $5, $6, $7);",
		e.Command, e.Args, e.UserID, e.GuildID, e.ChannelID, e.MessageID, e.SentAt)
	return err
}

func (db *CommandLogDB) GetCommandCount() (int, error) {
	var count int
	err := db.GetConn().Get(&count, "SELECT COUNT(*) FROM command_log;")
	return count, err
}

type GuildDB struct {
	DB
}

func (db *GuildDB) CreateGuild(guildID string) error {
	_, err := db.GetConn().Exec("INSERT INTO guild VALUES($1)", guildID)
	return err
}

func (db *GuildDB) GetGuild(guildID string) (*structs.Guild, error) {
	var guild structs.Guild
	err := db.GetConn().Get(&guild, "SELECT * FROM guild WHERE guild_id=$1", guildID)
	return &guild, err
}

func (db *GuildDB) UpdateGuild(g *structs.Guild) error {
	_, err := db.GetConn().Exec("UPDATE guild SET use_warns=$1, max_warns=$2, warn_duration=$3, automod_log_channel_id=$4, fishing_channel_id=$5 WHERE guild_id=$6",
		g.UseWarns, g.MaxWarns, g.WarnDuration, g.AutomodLogChannelID, g.FishingChannelID, g.GuildID)
	return err
}
