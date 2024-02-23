package moderation

import (
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/pkg/mio/bot"
	"github.com/intrntsrfr/meido/pkg/mio/discord"
)

func newLockdownChannelCommand(m *module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "lockdown",
		Description:      "Locks the current channel.",
		Triggers:         []string{"m?lockdown"},
		Usage:            "m?lockdown",
		Cooldown:         10,
		CooldownScope:    bot.CooldownScopeChannel,
		RequiredPerms:    discordgo.PermissionManageRoles,
		CheckBotPerms:    true,
		RequiresUserType: bot.UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         false,
		Enabled:          true,
		Execute:          m.lockdownCommand,
	}
}

func (m *module) lockdownCommand(msg *discord.DiscordMessage) {
	if len(msg.Args()) < 1 {
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

func newUnlockChannelCommand(m *module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "unlock",
		Description:      "Unlocks a previously locked channel.",
		Triggers:         []string{"m?unlock"},
		Usage:            "m?unlock",
		Cooldown:         10,
		CooldownScope:    bot.CooldownScopeChannel,
		RequiredPerms:    discordgo.PermissionManageRoles,
		CheckBotPerms:    true,
		RequiresUserType: bot.UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         false,
		Enabled:          true,
		Execute:          m.unlockCommand,
	}
}

func (m *module) unlockCommand(msg *discord.DiscordMessage) {
	if len(msg.Args()) < 1 {
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
