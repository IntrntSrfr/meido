package customrole

import "github.com/intrntsrfr/meido/internal/database"

type ICustomRoleDB interface {
	database.DB

	CreateCustomRole(guildID, userID, roleID string) error
	GetCustomRole(guildID, userID string) (*CustomRole, error)
	GetCustomRolesByGuild(guildID string) ([]*CustomRole, error)
	UpdateCustomRole(role *CustomRole) error
	DeleteCustomRole(uid int) error
}

type CustomRoleDB struct {
	database.DB
}

func (db *CustomRoleDB) CreateCustomRole(guildID, userID, roleID string) error {
	_, err := db.Conn().Exec("INSERT INTO custom_role(guild_id, user_id, role_id) VALUES($1, $2, $3);", guildID, userID, roleID)
	return err
}

func (db *CustomRoleDB) GetCustomRole(guildID, userID string) (*CustomRole, error) {
	var role *CustomRole
	err := db.Conn().Get(&role, "SELECT * FROM custom_role WHERE guild_id=$1 AND user_id=$2", guildID, userID)
	return role, err
}

func (db *CustomRoleDB) GetCustomRolesByGuild(guildID string) ([]*CustomRole, error) {
	var roles []*CustomRole
	err := db.Conn().Select(&roles, "SELECT * FROM custom_role WHERE guild_id=$1", guildID)
	return roles, err
}

func (db *CustomRoleDB) UpdateCustomRole(role *CustomRole) error {
	_, err := db.Conn().Exec("UPDATE custom_role SET role_id=$1 WHERE guild_id=$2 AND user_id=$3", role.RoleID, role.GuildID, role.UserID)
	return err
}

func (db *CustomRoleDB) DeleteCustomRole(uid int) error {
	_, err := db.Conn().Exec("DELETE FROM custom_role WHERE uid=$1", uid)
	return err
}
