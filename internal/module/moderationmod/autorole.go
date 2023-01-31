package moderationmod

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/pkg/mio"
	"strings"
)

func NewAutoRoleCommand(m *ModerationMod) *mio.ModCommand {
	return &mio.ModCommand{
		Mod:           m,
		Name:          "autorolesettings",
		Description:   "Sets the autorole to a supplied role name. If no role is supplied, it will be reset.",
		Triggers:      []string{"m?setautorole"},
		Usage:         "m?setautorole | m?setautorole [rolename]",
		Cooldown:      2,
		RequiredPerms: discordgo.PermissionAdministrator,
		RequiresOwner: false,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           m.autoroleCommand,
	}
}
func (m *ModerationMod) autoroleCommand(msg *mio.DiscordMessage) {
	if msg.LenArgs() == 1 {
		err := m.db.DeleteAutoRole(msg.GuildID())
		if err != nil {
			_, _ = msg.Reply("Failed to remove autorole")
			return
		}
		_, _ = msg.Reply("Cleared autorole")
		return
	}

	query := strings.Join(msg.Args()[1:], " ")
	role, err := msg.Discord.GuildRoleByName(msg.GuildID(), query)
	if err != nil {
		_, _ = msg.Reply("I could not find that role")
		return
	}

	// the autorole already exists, update it
	if _, err = m.db.GetAutoRole(msg.GuildID()); err == nil {
		err = m.db.UpdateAutoRole(msg.GuildID(), role.ID)
		if err != nil {
			_, _ = msg.Reply("Failed to set autorole")
			return
		}
		_, _ = msg.Reply(fmt.Sprintf("Autorole set to role `%v` (%v)", role.Name, role.ID))
		return
	}

	err = m.db.CreateAutoRole(msg.GuildID(), role.ID)
	if err != nil {
		_, _ = msg.Reply("Failed to set autorole")
		return
	}
	_, _ = msg.Reply(fmt.Sprintf("Autorole set to role `%v` (%v)", role.Name, role.ID))
}
