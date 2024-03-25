package database

import (
	"time"

	"github.com/intrntsrfr/meido/internal/structs"
	"github.com/jmoiron/sqlx"

	_ "github.com/lib/pq"
)

type PsqlDB struct {
	pool    *sqlx.DB
	connStr string
	IGuildDB
	ICommandLogDB
	IProcessedEventsDB
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
	db.IProcessedEventsDB = &ProcessedEventsDB{db}
	return db, nil
}

func (p *PsqlDB) Conn() *sqlx.DB {
	return p.pool
}

func (p *PsqlDB) Close() error {
	return p.pool.Close()
}

type CommandLogDB struct {
	DB
}

func (db *CommandLogDB) CreateCommandLogEntry(e *structs.CommandLogEntry) error {
	_, err := db.Conn().Exec("INSERT INTO command_log VALUES(DEFAULT, $1, $2, $3, $4, $5, $6, $7);",
		e.Command, e.Args, e.UserID, e.GuildID, e.ChannelID, e.MessageID, e.SentAt)
	return err
}

func (db *CommandLogDB) GetCommandCount() (int, error) {
	var count int
	err := db.Conn().Get(&count, "SELECT COUNT(*) FROM command_log;")
	return count, err
}

type GuildDB struct {
	DB
}

func (db *GuildDB) CreateGuild(guildID string, joinedAt time.Time) error {
	_, err := db.Conn().Exec("INSERT INTO guild(guild_id, joined_at) VALUES($1, $2)", guildID, joinedAt)
	return err
}

func (db *GuildDB) GetGuild(guildID string) (*structs.Guild, error) {
	var guild structs.Guild
	err := db.Conn().Get(&guild, "SELECT * FROM guild WHERE guild_id=$1", guildID)
	return &guild, err
}

func (db *GuildDB) UpdateGuild(g *structs.Guild) error {
	_, err := db.Conn().Exec("UPDATE guild SET use_warns=$1, max_warns=$2, warn_duration=$3, automod_log_channel_id=$4, fishing_channel_id=$5, joined_at=$6 WHERE guild_id=$7",
		g.UseWarns, g.MaxWarns, g.WarnDuration, g.AutomodLogChannelID, g.FishingChannelID, g.JoinedAt, g.GuildID)
	return err
}

type ProcessedEventsDB struct {
	DB
}

func (db *ProcessedEventsDB) UpsertCount(eventType string, sentAt time.Time) error {
	query := `
    INSERT INTO processed_events (sent_at, event_type, count)
    VALUES (date_trunc('day', $2::timestamp), $1, 1)
    ON CONFLICT (sent_at, event_type)
    DO UPDATE SET count = processed_events.count + 1
    `

	_, err := db.Conn().Exec(query, eventType, sentAt)
	return err
}
