package database

import (
	"github.com/intrntsrfr/meido/internal/structs"
	"github.com/jmoiron/sqlx"
	"time"

	_ "github.com/lib/pq"
)

type PsqlDB struct {
	pool    *sqlx.DB
	connStr string
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
	return db, nil
}

func (p *PsqlDB) GetConn() *sqlx.DB {
	return p.pool
}

func (p *PsqlDB) Close() error {
	return p.pool.Close()
}

func (p *PsqlDB) GetGuildWarnsIfActive(guildID string) ([]*structs.Warn, error) {
	var warns []*structs.Warn
	err := p.pool.Select(&warns, "SELECT * FROM warn WHERE guild_id=$1 AND is_valid ORDER BY given_at DESC", guildID)
	return warns, err
}

func (p *PsqlDB) GetGuildWarns(guildID string) ([]*structs.Warn, error) {
	var warns []*structs.Warn
	err := p.pool.Select(&warns, "SELECT * FROM warn WHERE guild_id=$1 ORDER BY given_at DESC", guildID)
	return warns, err
}

func (p *PsqlDB) GetGuild(guildID string) (*structs.Guild, error) {
	var guild structs.Guild
	err := p.pool.Get(&guild, "SELECT * FROM guild WHERE guild_id=$1", guildID)
	return &guild, err
}

func (p *PsqlDB) GetUserRole(guildID, userID string) (*structs.CustomRole, error) {
	var userRole structs.CustomRole
	err := p.pool.Get(&userRole, "SELECT * FROM user_role WHERE guild_id=$1 AND user_id=$2", guildID, userID)
	return &userRole, err
}

func (p *PsqlDB) GetValidUserWarnCount(guildID, userID string) (int, error) {
	var count int
	err := p.pool.Get(&count, "SELECT COUNT(*) FROM warn WHERE guild_id=$1 AND user_id=$2 AND is_valid", guildID, userID)
	return count, err
}

func (p *PsqlDB) CreateCommandLogEntry(e *structs.CommandLogEntry) error {
	_, err := p.pool.Exec("INSERT INTO command_log VALUES(DEFAULT, $1, $2, $3, $4, $5, $6, $7);",
		e.Command, e.Args, e.UserID, e.GuildID, e.ChannelID, e.MessageID, e.SentAt)
	return err
}

func (p *PsqlDB) GetCommandCount() (int, error) {
	var count int
	err := p.pool.Get(&count, "SELECT COUNT(*) FROM command_log;")
	return count, err
}

func (p *PsqlDB) GetGuildFilterByPhrase(guildID, phrase string) (*structs.Filter, error) {
	var filter structs.Filter
	err := p.pool.Get(&filter, "SELECT * FROM filter WHERE guild_id = $1 AND phrase = $2", guildID, phrase)
	return &filter, err
}

func (p *PsqlDB) GetGuildFilters(guildID string) ([]*structs.Filter, error) {
	var filters []*structs.Filter
	err := p.pool.Select(&filters, "SELECT * FROM filter WHERE guild_id=$1", guildID)
	return filters, err
}

func (p *PsqlDB) DeleteGuildFilter(filterID int) error {
	_, err := p.pool.Exec("DELETE FROM filter WHERE uid=$1", filterID)
	return err
}

func (p *PsqlDB) DeleteGuildFilters(guildID string) error {
	_, err := p.pool.Exec("DELETE FROM filter WHERE guild_id=$1", guildID)
	return err
}

func (p *PsqlDB) InsertWarn(guildID, userID, reason, givenByID string) error {
	_, err := p.pool.Exec("INSERT INTO warn VALUES(DEFAULT, $1, $2, $3, $4, $5, $6)",
		guildID, userID, reason, givenByID, time.Now(), true)
	return err
}

func (p *PsqlDB) ClearActiveUserWarns(guildID, userID, clearedByID string) error {
	_, err := p.pool.Exec("UPDATE warn SET is_valid=false, cleared_by_id=$1, cleared_at=$2 WHERE guild_id=$3 AND user_id=$4 and is_valid",
		clearedByID, time.Now(), guildID, userID)
	return err
}

func (p *PsqlDB) InsertNewAquarium(userID string) error {
	_, err := p.pool.Exec("INSERT INTO aquarium VALUES ($1)", userID)
	return err
}

func (p *PsqlDB) GetAquarium(userID string) (*structs.Aquarium, error) {
	var aquarium structs.Aquarium
	err := p.pool.Get(&aquarium, "SELECT * FROM aquarium WHERE user_id=$1", userID)
	return &aquarium, err
}

func (p *PsqlDB) UpdateAquarium(aq *structs.Aquarium) error {
	_, err := p.pool.Exec("UPDATE aquarium SET common=$1, uncommon=$2, rare=$3, super_rare=$4, legendary=$5 WHERE user_id=$6",
		aq.Common, aq.Uncommon, aq.Rare, aq.SuperRare, aq.Legendary, aq.UserID)
	return err
}

func (p *PsqlDB) CreateGuild(guildID string) error {
	_, err := p.pool.Exec("INSERT INTO guild VALUES($1)", guildID)
	return err
}

func (p *PsqlDB) UpdateGuild(g *structs.Guild) error {
	_, err := p.pool.Exec("UPDATE guild SET use_warns=$1, max_warns=$2, warn_duration=$3, automod_log_channel_id=$4, fishing_channel_id=$5 WHERE guild_id=$6",
		g.UseWarns, g.MaxWarns, g.WarnDuration, g.AutomodLogChannelID, g.FishingChannelID, g.GuildID)
	return err
}

func (p *PsqlDB) CreateGuildFilter(guildID, phrase string) error {
	_, err := p.pool.Exec("INSERT INTO filter VALUES (DEFAULT, $1, $2)", guildID, phrase)
	return err
}

func (p *PsqlDB) CreateMemberWarn(guildID, userID, reason, authorID string) error {
	_, err := p.pool.Exec("INSERT INTO warn VALUES(DEFAULT, $1, $2, $3, $4, $5, $6)",
		guildID, userID, reason, authorID, time.Now(), true)
	return err
}

func (p *PsqlDB) GetMemberWarns(guildID, userID string) ([]*structs.Warn, error) {
	var warns []*structs.Warn
	err := p.pool.Select(&warns, "SELECT * FROM warn WHERE guild_id=$1 AND user_id=$2 ORDER BY given_at DESC", guildID, userID)
	return warns, err
}

func (p *PsqlDB) GetMemberWarnsIfActive(guildID, userID string) ([]*structs.Warn, error) {
	var warns []*structs.Warn
	err := p.pool.Select(&warns, "SELECT * FROM warn WHERE guild_id=$1 AND user_id=$2 AND is_valid ORDER BY given_at DESC", guildID, userID)
	return warns, err
}

func (p *PsqlDB) UpdateMemberWarn(warn *structs.Warn) error {
	_, err := p.pool.Exec("UPDATE warn SET is_valid=false, cleared_by_id=$1, cleared_at=$2 WHERE uid = $3",
		warn.ClearedByID, warn.ClearedAt, warn.UID)
	return err
}

func (p *PsqlDB) CreateCustomRole(guildID, userID, roleID string) error {
	_, err := p.pool.Exec("INSERT INTO user_role(guild_id, user_id, role_id) VALUES($1, $2, $3);", guildID, userID, roleID)
	return err
}

func (p *PsqlDB) GetCustomRole(guildID, userID string) (*structs.CustomRole, error) {
	var role *structs.CustomRole
	err := p.pool.Get(&role, "SELECT * FROM user_role WHERE guild_id=$1 AND user_id=$2", guildID, userID)
	return role, err
}

func (p *PsqlDB) GetCustomRolesByGuild(guildID string) ([]*structs.CustomRole, error) {
	var roles []*structs.CustomRole
	err := p.pool.Select(&roles, "SELECT * FROM user_role WHERE guild_id=$1", guildID)
	return roles, err
}

func (p *PsqlDB) UpdateCustomRole(role *structs.CustomRole) error {
	_, err := p.pool.Exec("UPDATE user_role SET role_id=$1 WHERE guild_id=$2 AND user_id=$3", role.RoleID, role.GuildID, role.UserID)
	return err
}

func (p *PsqlDB) DeleteCustomRole(uid int) error {
	_, err := p.pool.Exec("DELETE FROM user_role WHERE uid=$1", uid)
	return err
}

func (p *PsqlDB) CreateAquarium(userID string) error {
	_, err := p.pool.Exec("INSERT INTO aquarium VALUES($1)", userID)
	return err
}
