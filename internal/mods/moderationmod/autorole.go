package moderationmod

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/base"
	"strings"
)

func NewAutoRoleCommand(m *ModerationMod) *base.ModCommand {
	return &base.ModCommand{
		Mod:           m,
		Name:          "autorolesettings",
		Description:   "Sets the autorole to a supplied role name. If no role is supplied, it will be reset.",
		Triggers:      []string{"m?setautorole"},
		Usage:         "m?setautorole | m?setautorole [rolename]",
		Cooldown:      2,
		RequiredPerms: discordgo.PermissionAdministrator,
		RequiresOwner: false,
		AllowedTypes:  base.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           m.autoroleCommand,
	}
}
func (m *ModerationMod) autoroleCommand(msg *base.DiscordMessage) {
	if msg.LenArgs() == 1 {
		_, err := m.db.Exec("UPDATE auto_role SET role_id=$1 WHERE guild_id=$2", "", msg.Message.GuildID)
		if err != nil {
			return
		}
		msg.Reply("Cleared autorole")
		return
	} else {
		query := strings.Join(msg.Args()[1:], " ")

		role, err := msg.Discord.GuildRoleByName(msg.GuildID(), query)
		if err != nil {
			msg.Reply("Couldn't find that role!")
			return
		}

		_, err = m.db.Exec("UPDATE auto_role SET role_id=$1 WHERE guild_id=$2", role.ID, msg.Message.GuildID)
		if err != nil {
			return
		}
		msg.Reply(fmt.Sprintf("Autorole set to role `%v` (%v)", role.Name, role.ID))
	}
}
