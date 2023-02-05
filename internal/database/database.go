package database

import (
	"github.com/intrntsrfr/meido/internal/structs"
	"github.com/jmoiron/sqlx"
)

type DB interface {
	GetConn() *sqlx.DB
	Close() error

	CommandLogDB
	GuildDB
	FilterDB
	WarnDB
	CustomRoleDB
	AquariumDB
}

type CommandLogDB interface {
	CreateCommandLogEntry(e *structs.CommandLogEntry) error
	GetCommandCount() (int, error)
}

type GuildDB interface {
	CreateGuild(guildID string) error
	UpdateGuild(g *structs.Guild) error
	GetGuild(guildID string) (*structs.Guild, error)
}

type FilterDB interface {
	CreateGuildFilter(guildID, phrase string) error
	GetGuildFilterByPhrase(guildID, phrase string) (*structs.Filter, error)
	GetGuildFilters(guildID string) ([]*structs.Filter, error)
	DeleteGuildFilter(filterID int) error
	DeleteGuildFilters(guildID string) error
}

type WarnDB interface {
	CreateMemberWarn(guildID, userID, reason, authorID string) error
	GetGuildWarns(guildID string) ([]*structs.Warn, error)
	GetGuildWarnsIfActive(guildID string) ([]*structs.Warn, error)
	GetMemberWarns(guildID, userID string) ([]*structs.Warn, error)
	GetMemberWarnsIfActive(guildID, userID string) ([]*structs.Warn, error)
	UpdateMemberWarn(warn *structs.Warn) error
}

type CustomRoleDB interface {
	CreateCustomRole(guildID, userID, roleID string) error
	GetCustomRole(guildID, userID string) (*structs.CustomRole, error)
	GetCustomRolesByGuild(guildID string) ([]*structs.CustomRole, error)
	UpdateCustomRole(role *structs.CustomRole) error
	DeleteCustomRole(uid int) error
}

type AquariumDB interface {
	CreateAquarium(userID string) error
	GetAquarium(userID string) (*structs.Aquarium, error)
	UpdateAquarium(aquarium *structs.Aquarium) error
}
