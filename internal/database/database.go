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
	AutoRoleDB
	FilterDB
	WarnDB
	MemberRoleDB
	AquariumDB
}

type CommandLogDB interface {
	CreateCommandLogEntry(e *structs.CommandLogEntry) error
	GetCommandCount() (int, error)
}

type GuildDB interface {
	CreateGuild(gid string) error
	UpdateGuild(g *structs.Guild) error
	GetGuild(gid string) (*structs.Guild, error)
}

type AutoRoleDB interface {
	CreateAutoRole(guildID, roleID string) error
	GetAutoRole(guildID string) (*structs.AutoRole, error)
	UpdateAutoRole(guildID, roleID string) error
	DeleteAutoRole(guildID string) error
}

type FilterDB interface {
	CreateGuildFilter(guildID, phrase string) error
	GetGuildFilterByPhrase(guildID, phrase string) (*structs.Filter, error)
	GetGuildFilters(guildID string) ([]*structs.Filter, error)
	DeleteGuildFilters(guildID string) error
}

type WarnDB interface {
	CreateMemberWarn(guildID, userID, reason, authorID string) error
	GetMemberWarns(guildID, userID string) ([]*structs.Warn, error)
	GetMemberWarnsIfActive(guildID, userID string) ([]*structs.Warn, error)
	UpdateMemberWarn(warn *structs.Warn) error
}

type MemberRoleDB interface {
	CreateMemberRole(guildID, userID string) error
	GetMemberRole(guildID, userID string) (*structs.UserRole, error)
	GetMemberRolesByGuild(guildID string) ([]*structs.UserRole, error)
	UpdateMemberRole(role *structs.UserRole) error
	DeleteMemberRole(uid string) error
}

type AquariumDB interface {
	CreateAquarium(userID string) error
	GetAquarium(userID string) (*structs.Aquarium, error)
	UpdateAquarium(aquarium *structs.Aquarium) error
}
