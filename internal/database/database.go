package database

import (
	"time"

	"github.com/intrntsrfr/meido/internal/structs"
	"github.com/jmoiron/sqlx"
)

type DB interface {
	Conn() *sqlx.DB
	Close() error

	ICommandLogDB
	IGuildDB
}

type ICommandLogDB interface {
	CreateCommandLogEntry(e *structs.CommandLogEntry) error
	GetCommandCount() (int, error)
}

type IGuildDB interface {
	CreateGuild(guildID string, joinedAt time.Time) error
	UpdateGuild(g *structs.Guild) error
	GetGuild(guildID string) (*structs.Guild, error)
}
