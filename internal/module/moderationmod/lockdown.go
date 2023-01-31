package moderationmod

import (
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/pkg/mio"
)

func NewLockdownChannelCommand(m *ModerationMod) *mio.ModCommand {
	return &mio.ModCommand{
		Mod:           m,
		Name:          "lockdown",
		Description:   "Locks the current channel.",
		Triggers:      []string{"m?lockdown"},
		Usage:         "m?lockdown",
		Cooldown:      10,
		RequiredPerms: discordgo.PermissionManageRoles,
		RequiresOwner: false,
		CheckBotPerms: true,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           m.lockdownCommand,
	}
}

func (m *ModerationMod) lockdownCommand(msg *mio.DiscordMessage) {
	if msg.LenArgs() < 1 {
		return
	}

	g, err := msg.Discord.Guild(msg.Message.GuildID)
	if err != nil {
		return
	}

	var eRole *discordgo.Role

	for _, val := range g.Roles {
		if val.ID == g.ID {
			eRole = val
		}
	}

	if eRole == nil {
		return
	}

	ch, err := msg.Discord.Channel(msg.Message.ChannelID)
	if err != nil {
		return
	}

	var ePerms *discordgo.PermissionOverwrite

	for _, val := range ch.PermissionOverwrites {
		if val.ID == eRole.ID {
			ePerms = val
		}
	}

	if ePerms == nil {
		return
	}

	if ePerms.Allow&discordgo.PermissionSendMessages == 0 && ePerms.Deny&discordgo.PermissionSendMessages == 0 {
		// DEFAULT
		err := msg.Sess.ChannelPermissionSet(
			ch.ID,
			eRole.ID,
			discordgo.PermissionOverwriteTypeRole,
			ePerms.Allow,
			ePerms.Deny+discordgo.PermissionSendMessages,
		)
		if err != nil {
			_, _ = msg.Reply("Could not lock channel.")
			return
		}
		_, _ = msg.Reply("Channel locked.")
	} else if ePerms.Allow&discordgo.PermissionSendMessages != 0 && ePerms.Deny&discordgo.PermissionSendMessages == 0 {
		// IS ALLOWED
		err := msg.Sess.ChannelPermissionSet(
			ch.ID,
			eRole.ID,
			discordgo.PermissionOverwriteTypeRole,
			ePerms.Allow-discordgo.PermissionSendMessages,
			ePerms.Deny+discordgo.PermissionSendMessages,
		)
		if err != nil {
			_, _ = msg.Reply("Could not lock channel.")
			return
		}
		_, _ = msg.Reply("Channel locked")
	} else if ePerms.Allow&discordgo.PermissionSendMessages == 0 && ePerms.Deny&discordgo.PermissionSendMessages != 0 {
		// IS DENIED
		_, _ = msg.Reply("Channel already locked")
	}
}

func NewUnlockChannelCommand(m *ModerationMod) *mio.ModCommand {
	return &mio.ModCommand{
		Mod:           m,
		Name:          "unlock",
		Description:   "Unlocks a previously locked channel.",
		Triggers:      []string{"m?unlock"},
		Usage:         "m?unlock",
		Cooldown:      10,
		RequiredPerms: discordgo.PermissionManageRoles,
		RequiresOwner: false,
		CheckBotPerms: true,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           m.unlockCommand,
	}
}

func (m *ModerationMod) unlockCommand(msg *mio.DiscordMessage) {
	if msg.LenArgs() < 1 {
		return
	}

	g, err := msg.Discord.Guild(msg.Message.GuildID)
	if err != nil {
		return
	}

	var eRole *discordgo.Role

	for _, val := range g.Roles {
		if val.ID == g.ID {
			eRole = val
		}
	}

	if eRole == nil {
		return
	}

	ch, err := msg.Discord.Channel(msg.Message.ChannelID)
	if err != nil {
		return
	}

	var ePerms *discordgo.PermissionOverwrite

	for _, val := range ch.PermissionOverwrites {
		if val.ID == eRole.ID {
			ePerms = val
		}
	}

	if ePerms == nil {
		return
	}

	if ePerms.Allow&discordgo.PermissionSendMessages == 0 && ePerms.Deny&discordgo.PermissionSendMessages == 0 {
		// DEFAULT
		_, _ = msg.Reply("Channel is already unlocked.")
	} else if ePerms.Allow&discordgo.PermissionSendMessages != 0 && ePerms.Deny&discordgo.PermissionSendMessages == 0 {
		// IS ALLOWED
		_, _ = msg.Reply("Channel is already unlocked.")
	} else if ePerms.Allow&discordgo.PermissionSendMessages == 0 && ePerms.Deny&discordgo.PermissionSendMessages != 0 {
		// IS DENIED
		err := msg.Sess.ChannelPermissionSet(
			ch.ID,
			eRole.ID,
			discordgo.PermissionOverwriteTypeRole,
			ePerms.Allow,
			ePerms.Deny-discordgo.PermissionSendMessages,
		)
		if err != nil {
			_, _ = msg.Reply("Could not unlock channel")
			return
		}
		_, _ = msg.Reply("Channel unlocked")
	}
}
