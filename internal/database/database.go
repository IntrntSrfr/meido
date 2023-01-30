package database

import (
	"github.com/intrntsrfr/meido/pkg/structs"
	"github.com/jmoiron/sqlx"
)

type DB interface {
	GetConn() *sqlx.DB
	Close() error

	CreateCommandLogEntry(e *structs.CommandLogEntry) error
	GetCommandCount() (int, error)

	CreateGuild(gid string) error
	UpdateGuild(g *structs.Guild) error
	GetGuild(gid string) (*structs.Guild, error)

	CreateGuildFilter(guildID, phrase string) error
	GetGuildFilterByPhrase(guildID, phrase string) (*structs.Filter, error)
	GetGuildFilters(guildID string) ([]*structs.Filter, error)
	DeleteGuildFilters(guildID string) error

	CreateMemberWarn(guildID, userID, reason, authorID string) error
	GetMemberWarns(guildID, userID string) ([]*structs.Warn, error)
	GetMemberWarnsIfActive(guildID, userID string) ([]*structs.Warn, error)
	UpdateMemberWarn(warn *structs.Warn) error

	CreateMemberRole(guildID, userID string) error
	GetMemberRole(guildID, userID string) (*structs.UserRole, error)
	GetMemberRolesByGuild(guildID string) ([]*structs.UserRole, error)
	UpdateMemberRole(role *structs.UserRole) error
	DeleteMemberRole(uid string) error

	CreateAquarium(userID string) error
	GetAquarium(userID string) (*structs.Aquarium, error)
	UpdateAquarium(aquarium *structs.Aquarium) error
}
