package moderation

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/pkg/mio"
)

func addAutoRoleOnJoin(m *Module) func(s *discordgo.Session, g *discordgo.GuildMemberAdd) {
	return func(s *discordgo.Session, g *discordgo.GuildMemberAdd) {
		gc, err := m.db.GetGuild(g.GuildID)
		if err != nil || gc.AutoRoleID == "" {
			return
		}

		if role, err := m.Bot.Discord.GuildRoleByNameOrID(g.GuildID, "", gc.AutoRoleID); err == nil {
			_ = s.GuildMemberRoleAdd(g.GuildID, g.User.ID, role.ID)
		}
	}
}

func newSetAutoRoleCommand(m *Module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:              m,
		Name:             "setautorole",
		Description:      "Sets an autorole for the server to a provided role",
		Triggers:         []string{"m?setautorole"},
		Usage:            "m?setautorole [role name / role ID]",
		Cooldown:         2,
		CooldownScope:    mio.Channel,
		RequiredPerms:    discordgo.PermissionAdministrator,
		CheckBotPerms:    false,
		RequiresUserType: mio.UserTypeAny,
		AllowedTypes:     mio.MessageTypeCreate,
		AllowDMs:         false,
		IsEnabled:        true,
		Run: func(msg *mio.DiscordMessage) {
			if len(msg.Args()) < 2 {
				return
			}

			query := strings.Join(msg.Args()[1:], " ")
			role, err := msg.Discord.GuildRoleByNameOrID(msg.GuildID(), query, msg.Args()[1])
			if err != nil {
				_, _ = msg.Reply("I could not find that role")
				return
			}

			// the autorole already exists, update it
			if g, err := m.db.GetGuild(msg.GuildID()); err == nil {
				g.AutoRoleID = role.ID
				if err := m.db.UpdateGuild(g); err != nil {
					_, _ = msg.Reply("Failed to set autorole")
					return
				}
				_, _ = msg.Reply(fmt.Sprintf("Autorole set to role `%v` (%v)", role.Name, role.ID))
				return
			}
		},
	}
}

func newRemoveAutoRoleCommand(m *Module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:              m,
		Name:             "removeautorole",
		Description:      "Removes the autorole for the server",
		Triggers:         []string{"m?removeautorole"},
		Usage:            "m?removeautorole",
		Cooldown:         2,
		CooldownScope:    mio.Channel,
		RequiredPerms:    discordgo.PermissionAdministrator,
		CheckBotPerms:    false,
		RequiresUserType: mio.UserTypeAny,
		AllowedTypes:     mio.MessageTypeCreate,
		AllowDMs:         false,
		IsEnabled:        true,
		Run: func(msg *mio.DiscordMessage) {
			rpl, err := msg.Reply("Are you sure you want to REMOVE the autorole? Please answer `yes` if you are.")
			if err != nil {
				_, _ = msg.Reply("There was an issue, please try again!")
				return
			}

			ch, err := m.Bot.Callbacks.Make(fmt.Sprintf("%v:%v", msg.ChannelID(), msg.AuthorID()))
			if err != nil {
				_, _ = msg.Reply("There was an issue, please try again!")
				return
			}
			defer m.Bot.Callbacks.Delete(fmt.Sprintf("%v:%v", msg.ChannelID(), msg.AuthorID()))

			var reply *mio.DiscordMessage
			t := time.NewTimer(time.Second * 15)
			for {
				select {
				case reply = <-ch:
				case <-t.C:
					_ = msg.Sess.ChannelMessageDelete(rpl.ChannelID, rpl.ID)
					_ = msg.Sess.ChannelMessageDelete(msg.ChannelID(), msg.Message.ID)
					return
				}

				if strings.ToLower(reply.RawContent()) == "yes" {
					_ = msg.Sess.ChannelMessageDelete(reply.ChannelID(), reply.Message.ID)
					_ = msg.Sess.ChannelMessageDelete(msg.ChannelID(), msg.Message.ID)
					break
				}
			}

			// the autorole exists, remove it
			if g, err := m.db.GetGuild(msg.GuildID()); err == nil {
				if g.AutoRoleID == "" {
					return
				}
				g.AutoRoleID = ""
				if err := m.db.UpdateGuild(g); err != nil {
					_, _ = msg.Reply("Failed to remove autorole")
					return
				}
				_, _ = msg.Reply("Autorole was removed")
				return
			}
		},
	}
}
