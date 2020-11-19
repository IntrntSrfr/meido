package moderationmod

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meidov2"
)

type LockdownChannelCommand struct {
	m       *ModerationMod
	Enabled bool
}

func NewLockdownChannelCommand(m *ModerationMod) meidov2.ModCommand {
	return &LockdownChannelCommand{
		m:       m,
		Enabled: true,
	}
}

func (c *LockdownChannelCommand) Name() string {
	return "Lockdown"
}
func (c *LockdownChannelCommand) Description() string {
	return "Locks down the current channel, denying the everyonerole send message perms."
}
func (c *LockdownChannelCommand) Triggers() []string {
	return []string{"m?lockdown"}
}
func (c *LockdownChannelCommand) Usage() string {
	return "m?lockdown"
}
func (c *LockdownChannelCommand) Cooldown() int {
	return 30
}
func (c *LockdownChannelCommand) RequiredPerms() int {
	return discordgo.PermissionManageRoles
}
func (c *LockdownChannelCommand) RequiresOwner() bool {
	return false
}
func (c *LockdownChannelCommand) IsEnabled() bool {
	return c.Enabled
}
func (c *LockdownChannelCommand) Run(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 1 || msg.Args()[0] != "m?lockdown" {
		return
	}

	uPerms, err := msg.Discord.UserChannelPermissions(msg.Member, msg.Message.ChannelID)
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

	c.m.cl <- msg

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

type UnlockChannelCommand struct {
	m       *ModerationMod
	Enabled bool
}

func NewUnlockChannelCommand(m *ModerationMod) meidov2.ModCommand {
	return &UnlockChannelCommand{
		m:       m,
		Enabled: true,
	}
}

func (c *UnlockChannelCommand) Name() string {
	return "unlock"
}
func (c *UnlockChannelCommand) Description() string {
	return "Unlocks a previously locked channel, setting the everyone roles send message permissions to default."
}
func (c *UnlockChannelCommand) Triggers() []string {
	return []string{"m?lockdown"}
}
func (c *UnlockChannelCommand) Usage() string {
	return "m?unlock"
}
func (c *UnlockChannelCommand) Cooldown() int {
	return 30
}
func (c *UnlockChannelCommand) RequiredPerms() int {
	return discordgo.PermissionManageRoles
}
func (c *UnlockChannelCommand) RequiresOwner() bool {
	return false
}
func (c *UnlockChannelCommand) IsEnabled() bool {
	return c.Enabled
}
func (c *UnlockChannelCommand) Run(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 1 || msg.Args()[0] != "m?unlock" {
		return
	}

	uPerms, err := msg.Discord.UserChannelPermissions(msg.Member, msg.Message.ChannelID)
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

	c.m.cl <- msg

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
