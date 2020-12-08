package moderationmod

import (
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meidov2"
)

func NewLockdownChannelCommand(m *ModerationMod) *meidov2.ModCommand {
	return &meidov2.ModCommand{
		Mod:           m,
		Name:          "lockdown",
		Description:   "Locks the current channel.",
		Triggers:      []string{"m?lockdown"},
		Usage:         "m?lockdown",
		Cooldown:      10,
		RequiredPerms: discordgo.PermissionManageRoles,
		RequiresOwner: false,
		CheckBotPerms: true,
		AllowedTypes:  meidov2.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           m.lockdownCommand,
	}
}

func (m *ModerationMod) lockdownCommand(msg *meidov2.DiscordMessage) {
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
			"role",
			ePerms.Allow,
			ePerms.Deny+discordgo.PermissionSendMessages,
		)
		if err != nil {
			msg.Reply("Could not lock channel.")
			return
		}
		msg.Reply("Channel locked.")
	} else if ePerms.Allow&discordgo.PermissionSendMessages != 0 && ePerms.Deny&discordgo.PermissionSendMessages == 0 {
		// IS ALLOWED
		err := msg.Sess.ChannelPermissionSet(
			ch.ID,
			eRole.ID,
			"role",
			ePerms.Allow-discordgo.PermissionSendMessages,
			ePerms.Deny+discordgo.PermissionSendMessages,
		)
		if err != nil {
			msg.Reply("Could not lock channel.")
			return
		}
		msg.Reply("Channel locked")
	} else if ePerms.Allow&discordgo.PermissionSendMessages == 0 && ePerms.Deny&discordgo.PermissionSendMessages != 0 {
		// IS DENIED
		msg.Reply("Channel already locked")
	}
}

func NewUnlockChannelCommand(m *ModerationMod) *meidov2.ModCommand {
	return &meidov2.ModCommand{
		Mod:           m,
		Name:          "unlock",
		Description:   "Unlocks a previously locked channel.",
		Triggers:      []string{"m?unlock"},
		Usage:         "m?unlock",
		Cooldown:      10,
		RequiredPerms: discordgo.PermissionManageRoles,
		RequiresOwner: false,
		CheckBotPerms: true,
		AllowedTypes:  meidov2.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           m.unlockCommand,
	}
}

func (m *ModerationMod) unlockCommand(msg *meidov2.DiscordMessage) {
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
		msg.Reply("Channel is already unlocked.")
	} else if ePerms.Allow&discordgo.PermissionSendMessages != 0 && ePerms.Deny&discordgo.PermissionSendMessages == 0 {
		// IS ALLOWED
		msg.Reply("Channel is already unlocked.")
	} else if ePerms.Allow&discordgo.PermissionSendMessages == 0 && ePerms.Deny&discordgo.PermissionSendMessages != 0 {
		// IS DENIED
		err := msg.Sess.ChannelPermissionSet(
			ch.ID,
			eRole.ID,
			"role",
			ePerms.Allow,
			ePerms.Deny-discordgo.PermissionSendMessages,
		)
		if err != nil {
			msg.Reply("Could not unlock channel")
			return
		}
		msg.Reply("Channel unlocked")
	}
}
