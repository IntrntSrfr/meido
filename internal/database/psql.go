package database

import (
	"github.com/intrntsrfr/meido/internal/structs"
	"github.com/jmoiron/sqlx"
	"time"
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

// "INSERT INTO command_log VALUES(DEFAULT, $1, $2, $3, $4, $5, $6, $7);",
//
//	cmd.Name, strings.Join(msg.Args(), " "), msg.AuthorID(), gid,
//	msg.ChannelID(), msg.Message.ID, time.Now()
func (p *PsqlDB) GetConn() *sqlx.DB {
	return p.pool
}

func (p *PsqlDB) Close() error {
	return p.pool.Close()
}

func (p *PsqlDB) GetGuild(guildID string) (*structs.Guild, error) {
	var guild structs.Guild
	err := p.pool.Get(&guild, "SELECT * FROM guild WHERE guild_id=$1", guildID)
	return &guild, err
}

func (p *PsqlDB) GetUserRole(guildID, userID string) (*structs.UserRole, error) {
	var userRole structs.UserRole
	err := p.pool.Get(&userRole, "SELECT * FROM user_role WHERE guild_id=$1 AND user_id=$2", guildID, userID)
	return &userRole, err
}

func (p *PsqlDB) GetValidUserWarnCount(guildID, userID string) (int, error) {
	var count int
	err := p.pool.Get(&count, "SELECT COUNT(*) FROM warn WHERE guild_id=$1 AND user_id=$2 AND is_valid", guildID, userID)
	return count, err
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

func (p *PsqlDB) CreateCommandLogEntry(e *structs.CommandLogEntry) error {
	//TODO implement me
	panic("implement me")
}

func (p *PsqlDB) CreateGuild(gid string) error {
	//TODO implement me
	panic("implement me")
}

func (p *PsqlDB) UpdateGuild(g *structs.Guild) error {
	//TODO implement me
	panic("implement me")
}

func (p *PsqlDB) CreateGuildFilter(guildID, phrase string) error {
	//TODO implement me
	panic("implement me")
}

func (p *PsqlDB) CreateMemberWarn(guildID, userID, reason, authorID string) error {
	//TODO implement me
	panic("implement me")
}

func (p *PsqlDB) GetMemberWarns(guildID, userID string) ([]*structs.Warn, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PsqlDB) GetMemberWarnsIfActive(guildID, userID string) ([]*structs.Warn, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PsqlDB) UpdateMemberWarn(warn *structs.Warn) error {
	//TODO implement me
	panic("implement me")
}

func (p *PsqlDB) CreateMemberRole(guildID, userID string) error {
	//TODO implement me
	panic("implement me")
}

func (p *PsqlDB) GetMemberRole(guildID, userID string) (*structs.UserRole, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PsqlDB) GetMemberRolesByGuild(guildID string) ([]*structs.UserRole, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PsqlDB) UpdateMemberRole(role *structs.UserRole) error {
	//TODO implement me
	panic("implement me")
}

func (p *PsqlDB) DeleteMemberRole(uid string) error {
	//TODO implement me
	panic("implement me")
}

func (p *PsqlDB) CreateAquarium(userID string) error {
	//TODO implement me
	panic("implement me")
}

func (p *PsqlDB) CreateAutoRole(guildID, roleID string) error {
	//TODO implement me
	panic("implement me")
}

func (p *PsqlDB) GetAutoRole(guildID string) (*structs.AutoRole, error) {
	//TODO implement me
	panic("implement me")
}

func (p *PsqlDB) UpdateAutoRole(guildID, roleID string) error {
	//TODO implement me
	panic("implement me")
}

func (p *PsqlDB) DeleteAutoRole(guildID string) error {
	//TODO implement me
	panic("implement me")
}
