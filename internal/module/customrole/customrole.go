package customrole

import (
	"bytes"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/g4s8/hexcolor"
	"github.com/intrntsrfr/meido/internal/database"
	"github.com/intrntsrfr/meido/pkg/mio/bot"
	"github.com/intrntsrfr/meido/pkg/mio/discord"
	"github.com/intrntsrfr/meido/pkg/utils"
	"github.com/intrntsrfr/meido/pkg/utils/builders"

	"go.uber.org/zap"

	"github.com/bwmarrin/discordgo"
)

type module struct {
	*bot.ModuleBase
	db ICustomRoleDB
}

func New(b *bot.Bot, db database.DB, logger *zap.Logger) bot.Module {
	logger = logger.Named("CustomRole")
	return &module{
		ModuleBase: bot.NewModule(b, "CustomRole", logger),
		db:         &CustomRoleDB{db},
	}
}

func (m *module) Hook() error {
	m.Bot.Discord.Sess.AddHandlerOnce(func(s *discordgo.Session, r *discordgo.Ready) {
		go clearDeletedRoles(m)
	})

	return m.RegisterCommands(
		newSetCustomRoleCommand(m),
		newRemoveCustomRoleCommand(m),
		newMyRoleCommand(m),
		newListCustomRolesCommand(m),
	)
}

func clearDeletedRoles(m *module) {
	refreshTicker := time.NewTicker(time.Hour)
	for range refreshTicker.C {
		m.Logger.Info("Checking for deleted roles")
		for _, g := range m.Bot.Discord.Guilds() {
			if g.Unavailable {
				continue
			}
			roles, err := m.db.GetCustomRolesByGuild(g.ID)
			if err != nil {
				continue
			}
			for _, ur := range roles {
				hasRole := false
				for _, gr := range g.Roles {
					if gr.ID == ur.RoleID {
						hasRole = true
						break
					}
				}
				if hasRole {
					continue
				}
				// delete role from guild if it no longer exists
				if err := m.db.DeleteCustomRole(ur.UID); err != nil {
					m.Logger.Error("Delete custom role failed",
						zap.Int("member roleID", ur.UID),
						zap.String("guildID", ur.GuildID),
						zap.String("roleID", ur.RoleID),
						zap.String("userID", ur.UserID))
				}
			}
		}
	}
}

func newSetCustomRoleCommand(m *module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "setcustomrole",
		Description:      "Sets or changes a custom role for a user",
		Triggers:         []string{"m?setuserrole", "m?setcustomrole"},
		Usage:            "m?setcustomrole [userID] [role]",
		Cooldown:         time.Second * 3,
		CooldownScope:    bot.CooldownScopeChannel,
		RequiredPerms:    discordgo.PermissionManageRoles,
		CheckBotPerms:    false,
		RequiresUserType: bot.UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         false,
		Enabled:          true,
		Execute: func(msg *discord.DiscordMessage) {
			if len(msg.Args()) < 3 {
				return
			}
			targetMember, err := msg.GetMemberAtArg(1)
			if err != nil {
				_, _ = msg.Reply("Could not find that user")
				return
			}

			if targetMember.User.Bot {
				_, _ = msg.Reply("Bots dont get to join the fun")
				return
			}

			g, err := msg.Discord.Guild(msg.Message.GuildID)
			if err != nil {
				_, _ = msg.Reply(err.Error())
				return
			}

			var selectedRole *discordgo.Role
			for _, role := range g.Roles {
				if role.ID == msg.Args()[2] {
					selectedRole = role
				} else if strings.EqualFold(role.Name, strings.Join(msg.Args()[2:], " ")) {
					selectedRole = role
				}
			}

			if selectedRole == nil {
				_, _ = msg.Reply("Could not find that role")
				return
			}

			if ur, err := m.db.GetCustomRole(g.ID, targetMember.User.ID); err == nil {
				if selectedRole.ID == ur.RoleID {
					return
				}
				ur.RoleID = selectedRole.ID
				if err := m.db.UpdateCustomRole(ur); err != nil {
					_, _ = msg.Reply("Could not set role, please try again")
					return
				}
				_, _ = msg.Reply(fmt.Sprintf("Updated member role for **%v** to **%v**", targetMember.User.String(), selectedRole.Name))
				return
			}

			if err := m.db.CreateCustomRole(g.ID, targetMember.User.ID, selectedRole.ID); err != nil {
				_, _ = msg.Reply("Could not set role, please try again")
				return
			}
			_, _ = msg.Reply(fmt.Sprintf("Bound role **%v** to user **%v**", selectedRole.Name, targetMember.User.String()))
		},
	}
}

func newRemoveCustomRoleCommand(m *module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "removecustomrole",
		Description:      "Removes a custom role that is bound to a user",
		Triggers:         []string{"m?removeuserrole", "m?removecustomrole"},
		Usage:            "m?removecustomrole [userID]",
		Cooldown:         time.Second * 3,
		CooldownScope:    bot.CooldownScopeChannel,
		RequiredPerms:    discordgo.PermissionManageRoles,
		CheckBotPerms:    false,
		RequiresUserType: bot.UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         false,
		Enabled:          true,
		Execute: func(msg *discord.DiscordMessage) {
			if len(msg.Args()) < 2 {
				return
			}
			targetUser, err := msg.GetMemberOrUserAtArg(1)
			if err != nil {
				_, _ = msg.Reply("Could not find user")
				return
			}

			if targetUser.Bot {
				_, _ = msg.Reply("Bots dont get to join the fun")
				return
			}

			if ur, err := m.db.GetCustomRole(msg.GuildID(), targetUser.ID); err == nil {
				if err := m.db.DeleteCustomRole(ur.UID); err != nil {
					_, _ = msg.Reply("Could not remove custom role, please try again")
					return
				}
				_, _ = msg.Reply(fmt.Sprintf("Removed custom role from %v", targetUser.Mention()))
				return
			}
		},
	}
}

func newMyRoleCommand(m *module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "myrole",
		Description:      "Displays a users bound role, or lets the user change the name or color of their bound role",
		Triggers:         []string{"m?myrole"},
		Usage:            "m?myrole | m?myrole 123123123123 | m?myrole color c0ffee | m?myrole name jeff",
		Cooldown:         time.Second * 3,
		CooldownScope:    bot.CooldownScopeChannel,
		RequiredPerms:    0,
		CheckBotPerms:    false,
		RequiresUserType: bot.UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         false,
		Enabled:          true,
		Execute:          m.myroleCommand,
	}
}

func (m *module) myroleCommand(msg *discord.DiscordMessage) {
	if len(msg.Args()) < 1 {
		return
	}
	var (
		err     error
		oldRole *discordgo.Role
		target  *discordgo.Member
	)
	g, err := msg.Discord.Guild(msg.Message.GuildID)
	if err != nil {
		_, _ = msg.Reply("some error occurred")
		return
	}

	switch la := len(msg.Args()); {
	case la > 2:
		if msg.Args()[1] != "name" && msg.Args()[1] != "color" {
			return
		}
		if allow, err := msg.BotHasPermissions(discordgo.PermissionManageRoles); err != nil || !allow {
			_, _ = msg.Reply("I am missing `manage roles` permissions!")
			return
		}

		ur, err := m.db.GetCustomRole(msg.GuildID(), msg.AuthorID())
		if err != nil && err != sql.ErrNoRows {
			m.Logger.Error("error fetching user role", zap.Error(err))
			_, _ = msg.Reply("There was an issue, please try again!")
			return
		} else if err == sql.ErrNoRows {
			_, _ = msg.Reply("No custom role set")
			return
		}

		for _, role := range g.Roles {
			if role.ID == ur.RoleID {
				oldRole = role
			}
		}
		if oldRole == nil {
			_, _ = msg.Reply("Could not find custom role")
			return
		}
		topBotRole := msg.Discord.HighestRolePosition(msg.GuildID(), msg.Sess.State().User.ID)
		if oldRole.Position >= topBotRole {
			_, _ = msg.Reply("I cannot edit this role, it is above me in the role hierarchy!")
			return
		}

		if msg.Args()[1] == "name" {
			newName := strings.Join(msg.RawArgs()[2:], " ")
			if _, err = msg.Discord.Sess.GuildRoleEdit(g.ID, oldRole.ID, &discordgo.RoleParams{Name: newName}); err != nil {
				_, _ = msg.Reply("There was an issue, please try again!")
				m.Logger.Error("Editing custom role name failed", zap.Error(err))
				return
			}
			embed := builders.NewEmbedBuilder().
				WithColor(oldRole.Color).
				WithDescription(fmt.Sprintf("Role name changed from %v to %v", oldRole.Name, newName))
			_, _ = msg.ReplyEmbed(embed.Build())

		} else if msg.Args()[1] == "color" {
			clrStr := msg.Args()[2]
			if !strings.HasPrefix(clrStr, "#") {
				clrStr = "#" + strings.TrimSpace(clrStr)
			}
			clr, err := hexcolor.Parse(clrStr)
			if err != nil {
				_, _ = msg.ReplyEmbed(&discordgo.MessageEmbed{Description: "Invalid color code", Color: utils.ColorCritical})
				return
			}

			clrInt := int(clr.R)<<16 + int(clr.G)<<8 + int(clr.B)
			if _, err = msg.Discord.Sess.GuildRoleEdit(g.ID, oldRole.ID, &discordgo.RoleParams{Color: &clrInt}); err != nil {
				_, _ = msg.Reply("There was an issue, please try again!")
				m.Logger.Error("Editing custom role color failed", zap.Error(err))
				return
			}
			embed := builders.NewEmbedBuilder().
				WithColor(clrInt).
				WithDescription(fmt.Sprintf("Color changed from #%06X to #%06X", oldRole.Color, clrInt))
			_, _ = msg.ReplyEmbed(embed.Build())
		}
		return
	case la == 1:
		target = msg.Member()
	case la == 2:
		if target, err = msg.GetMemberAtArg(1); err != nil {
			_, _ = msg.Reply("Could not find that user")
			return
		}
	default:
		return
	}
	if target == nil {
		return
	}

	ur, err := m.db.GetCustomRole(msg.GuildID(), target.User.ID)
	if err != nil && err != sql.ErrNoRows {
		_, _ = msg.Reply("there was an error, please try again")
		m.Logger.Error("Fetching custom role failed", zap.Error(err))
		return
	} else if err == sql.ErrNoRows {
		_, _ = msg.Reply("No custom role set.")
		return
	}

	var customRole *discordgo.Role
	for i := range g.Roles {
		role := g.Roles[i]
		if role.ID == ur.RoleID {
			customRole = role
		}
	}
	if customRole == nil {
		_, _ = msg.Reply("The custom role is broken, wait for someone to fix it or try setting a new custom role")
		return
	}

	embed := builders.NewEmbedBuilder().
		WithTitle(fmt.Sprintf("Custom role for %v", target.User.String())).
		WithColor(customRole.Color).
		AddField("Name", customRole.Name, true).
		AddField("Color", fmt.Sprintf("#%06X", customRole.Color), true)
	_, _ = msg.ReplyEmbed(embed.Build())
}

func newListCustomRolesCommand(m *module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "listcustomroles",
		Description:      "Returns a list of custom roles for the server. It also shows whether users with custom roles are in the server or not",
		Triggers:         []string{"m?listuserroles", "m?listcustomroles"},
		Usage:            "m?listcustomroles",
		Cooldown:         time.Second * 30,
		CooldownScope:    bot.CooldownScopeChannel,
		RequiredPerms:    discordgo.PermissionManageRoles,
		CheckBotPerms:    false,
		RequiresUserType: bot.UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         false,
		Enabled:          true,
		Execute: func(msg *discord.DiscordMessage) {
			roles, err := m.db.GetCustomRolesByGuild(msg.GuildID())
			if err != nil {
				_, _ = msg.Reply("There was an issue, please try again!")
				return
			}

			g, err := msg.Discord.Guild(msg.Message.GuildID)
			if err != nil {
				_, _ = msg.Reply("some error occurred, please try again")
				return
			}

			builder := strings.Builder{}
			builder.WriteString(fmt.Sprintf("Custom roles in %v | Amount: %v\n\n", g.Name, len(roles)))
			for _, ur := range roles {
				role, err := msg.Discord.Role(g.ID, ur.RoleID)
				if err != nil {
					continue
				}

				mem, err := msg.Discord.Member(g.ID, ur.UserID)
				if err != nil {
					builder.WriteString(fmt.Sprintf("%v (%v) | Belongs to: %v - User no longer in guild.\n", role.Name, role.ID, ur.UserID))
				} else {
					builder.WriteString(fmt.Sprintf("%v (%v) | Belongs to: %v (%v)\n", role.Name, role.ID, mem.User.String(), mem.User.ID))
				}
			}

			data := &discordgo.MessageSend{Content: builder.String()}
			if builder.Len() > 1024 {
				data.File = &discordgo.File{
					Name:   "roles.txt",
					Reader: bytes.NewBufferString(builder.String()),
				}
				data.Content = ""
			}
			_, _ = msg.ReplyComplex(data)
		},
	}
}
