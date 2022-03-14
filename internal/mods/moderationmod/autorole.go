package moderationmod

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	base2 "github.com/intrntsrfr/meido/base"
	"strings"
)

func NewAutoRoleCommand(m *ModerationMod) *base2.ModCommand {
	return &base2.ModCommand{
		Mod:           m,
		Name:          "autorolesettings",
		Description:   "Sets the autorole to a supplied role name. If no role is supplied, it will be reset.",
		Triggers:      []string{"m?setautorole"},
		Usage:         "m?setautorole | m?setautorole [rolename]",
		Cooldown:      2,
		RequiredPerms: discordgo.PermissionAdministrator,
		RequiresOwner: false,
		AllowedTypes:  base2.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           m.autoroleCommand,
	}
}
func (m *ModerationMod) autoroleCommand(msg *base2.DiscordMessage) {
	if msg.LenArgs() == 1 {
		_, err := m.db.Exec("UPDATE auto_role SET role_id=$1 WHERE guild_id=$2", "", msg.Message.GuildID)
		if err != nil {
			return
		}
		msg.Reply("Cleared autorole")
		return
	} else {
		g, err := msg.Discord.Guild(msg.Message.GuildID)
		if err != nil {
			msg.Reply("an error occurred, try again")
			return
		}

		query := strings.Join(msg.Args()[1:], " ")

		var role *discordgo.Role
		for _, gRole := range g.Roles {
			if gRole.Name == query {
				role = gRole
			}
		}

		if role == nil {
			msg.Reply("Could not find that role")
			return
		}

		_, err = m.db.Exec("UPDATE auto_role SET role_id=$1 WHERE guild_id=$2", role.ID, msg.Message.GuildID)
		if err != nil {
			return
		}
		msg.Reply(fmt.Sprintf("Autorole set to role `%v` (%v)", role.Name, role.ID))
	}
}
