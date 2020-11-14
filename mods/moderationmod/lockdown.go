package moderationmod

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meidov2"
)

func (m *ModerationMod) LockdownChannel(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 1 || msg.Args()[0] != "m?lockdown" {
		return
	}

	uPerms, err := msg.Discord.UserChannelPermissions(msg.Author, msg.Message.ChannelID)
	if err != nil {
		fmt.Println(err)
		return
	}
	if uPerms&discordgo.PermissionManageRoles == 0 && uPerms&discordgo.PermissionAdministrator == 0 {
		return
	}

	botPerms, err := msg.Discord.Sess.State.UserChannelPermissions(msg.Sess.State.User.ID, msg.Message.ChannelID)
	if err != nil {
		return
	}
	if botPerms&discordgo.PermissionManageRoles == 0 && botPerms&discordgo.PermissionAdministrator == 0 {
		return
	}

	m.cl <- msg

	g, err := msg.Sess.State.Guild(msg.Message.GuildID)
	if err != nil {
		return
	}

	var erole *discordgo.Role

	for _, val := range g.Roles {
		if val.ID == g.ID {
			erole = val
		}
	}

	if erole == nil {
		return
	}

	ch, err := msg.Sess.State.Channel(msg.Message.ChannelID)
	if err != nil {
		return
	}

	var eperms *discordgo.PermissionOverwrite

	for _, val := range ch.PermissionOverwrites {
		if val.ID == erole.ID {
			eperms = val
		}
	}

	if eperms == nil {
		return
	}

	if eperms.Allow&discordgo.PermissionSendMessages == 0 && eperms.Deny&discordgo.PermissionSendMessages == 0 {
		// DEFAULT
		err := msg.Sess.ChannelPermissionSet(
			ch.ID,
			erole.ID,
			"role",
			eperms.Allow,
			eperms.Deny+discordgo.PermissionSendMessages,
		)
		if err != nil {
			msg.Reply("Could not lock channel.")
			return
		}
		msg.Reply("Channel locked.")
	} else if eperms.Allow&discordgo.PermissionSendMessages != 0 && eperms.Deny&discordgo.PermissionSendMessages == 0 {
		// IS ALLOWED
		err := msg.Sess.ChannelPermissionSet(
			ch.ID,
			erole.ID,
			"role",
			eperms.Allow-discordgo.PermissionSendMessages,
			eperms.Deny+discordgo.PermissionSendMessages,
		)
		if err != nil {
			msg.Reply("Could not lock channel.")
			return
		}
		msg.Reply("Channel locked")
	} else if eperms.Allow&discordgo.PermissionSendMessages == 0 && eperms.Deny&discordgo.PermissionSendMessages != 0 {
		// IS DENIED
		msg.Reply("Channel already locked")
	}
}

func (m *ModerationMod) UnlockChannel(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 1 || msg.Args()[0] != "m?unlock" {
		return
	}

	uPerms, err := msg.Discord.UserChannelPermissions(msg.Author, msg.Message.ChannelID)
	if err != nil {
		fmt.Println(err)
		return
	}
	if uPerms&discordgo.PermissionManageRoles == 0 && uPerms&discordgo.PermissionAdministrator == 0 {
		return
	}

	botPerms, err := msg.Discord.Sess.State.UserChannelPermissions(msg.Sess.State.User.ID, msg.Message.ChannelID)
	if err != nil {
		return
	}
	if botPerms&discordgo.PermissionManageRoles == 0 && botPerms&discordgo.PermissionAdministrator == 0 {
		return
	}

	m.cl <- msg

	g, err := msg.Sess.State.Guild(msg.Message.GuildID)
	if err != nil {
		return
	}

	var erole *discordgo.Role

	for _, val := range g.Roles {
		if val.ID == g.ID {
			erole = val
		}
	}

	if erole == nil {
		return
	}

	ch, err := msg.Sess.State.Channel(msg.Message.ChannelID)
	if err != nil {
		return
	}

	var eperms *discordgo.PermissionOverwrite

	for _, val := range ch.PermissionOverwrites {
		if val.ID == erole.ID {
			eperms = val
		}
	}

	if eperms == nil {
		return
	}

	if eperms.Allow&discordgo.PermissionSendMessages == 0 && eperms.Deny&discordgo.PermissionSendMessages == 0 {
		// DEFAULT
		msg.Reply("Channel is already unlocked.")
	} else if eperms.Allow&discordgo.PermissionSendMessages != 0 && eperms.Deny&discordgo.PermissionSendMessages == 0 {
		// IS ALLOWED
		msg.Reply("Channel is already unlocked.")
	} else if eperms.Allow&discordgo.PermissionSendMessages == 0 && eperms.Deny&discordgo.PermissionSendMessages != 0 {
		// IS DENIED
		err := msg.Sess.ChannelPermissionSet(
			ch.ID,
			erole.ID,
			"role",
			eperms.Allow,
			eperms.Deny-discordgo.PermissionSendMessages,
		)
		if err != nil {
			msg.Reply("Could not unlock channel")
			return
		}
		msg.Reply("Channel unlocked")
	}
}
