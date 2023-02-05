package customrole

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/intrntsrfr/meido/internal/database"
	"github.com/intrntsrfr/meido/internal/helpers"
	"github.com/intrntsrfr/meido/pkg/mio"
	"github.com/intrntsrfr/meido/pkg/utils"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/bwmarrin/discordgo"
)

type Module struct {
	*mio.ModuleBase
	db database.DB
}

func New(bot *mio.Bot, db *database.PsqlDB, logger *zap.Logger) mio.Module {
	return &Module{
		ModuleBase: mio.NewModule(bot, "CustomRole", logger.Named("customrole")),
		db:         db,
	}
}

func (m *Module) Hook() error {
	m.Bot.Discord.Sess.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		go clearDeletedRoles(m)
	})

	return m.RegisterCommands([]*mio.ModuleCommand{
		newSetCustomRoleCommand(m),
		newRemoveCustomRoleCommand(m),
		newMyRoleCommand(m),
		NewListCustomRolesCommand(m),
	})
}

func clearDeletedRoles(m *Module) {
	refreshTicker := time.NewTicker(time.Hour)
	for range refreshTicker.C {
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
					m.Log.Error("could not delete custom role",
						zap.Int("member role ID", ur.UID),
						zap.String("guild id", ur.GuildID),
						zap.String("role id", ur.RoleID),
						zap.String("user id", ur.UserID))
				}
			}
		}
	}
}

func newSetCustomRoleCommand(m *Module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:           m,
		Name:          "setcustomrole",
		Description:   "Sets or changes a custom role for a user",
		Triggers:      []string{"m?setuserrole", "m?setcustomrole"},
		Usage:         "m?setcustomrole 1231231231231 cool role",
		Cooldown:      3,
		RequiredPerms: discordgo.PermissionManageRoles,
		RequiresOwner: false,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run: func(msg *mio.DiscordMessage) {
			if msg.LenArgs() < 3 {
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
				} else if strings.ToLower(role.Name) == strings.ToLower(strings.Join(msg.Args()[2:], " ")) {
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

func newRemoveCustomRoleCommand(m *Module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:           m,
		Name:          "removecustomrole",
		Description:   "Removes a custom role that is bound to a user",
		Triggers:      []string{"m?removeuserrole", "m?removecustomrole"},
		Usage:         "m?removecustomrole 1231231231231",
		Cooldown:      3,
		RequiredPerms: discordgo.PermissionManageRoles,
		RequiresOwner: false,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run: func(msg *mio.DiscordMessage) {
			if msg.LenArgs() < 2 {
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

func newMyRoleCommand(m *Module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:           m,
		Name:          "myrole",
		Description:   "Displays a users bound role, or lets the user change the name or color of their bound role",
		Triggers:      []string{"m?myrole"},
		Usage:         "m?myrole | m?myrole 123123123123 | m?myrole color c0ffee | m?myrole name jeff",
		Cooldown:      3,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           m.myroleCommand,
	}
}

func (m *Module) myroleCommand(msg *mio.DiscordMessage) {
	if msg.LenArgs() < 1 {
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

	switch la := msg.LenArgs(); {
	case la > 2:
		if msg.Args()[1] != "name" && msg.Args()[1] != "color" {
			return
		}

		if allow, err := msg.Discord.HasPermissions(msg.Message.ChannelID, discordgo.PermissionManageRoles); err != nil || !allow {
			_, _ = msg.Reply("I am missing 'manage roles' permissions!")
			return
		}

		ur, err := m.db.GetCustomRole(msg.GuildID(), msg.AuthorID())
		if err != nil && err != sql.ErrNoRows {
			m.Log.Error("error fetching user role", zap.Error(err))
			_, _ = msg.Reply("there was an error, please try again")
			return
		} else if err == sql.ErrNoRows {
			_, _ = msg.Reply("No custom role set.")
			return
		}

		for _, role := range g.Roles {
			if role.ID == ur.RoleID {
				oldRole = role
			}
		}

		if oldRole == nil {
			_, _ = msg.Reply("couldnt find role")
			return
		}

		if msg.Args()[1] == "name" {
			newName := strings.Join(msg.RawArgs()[2:], " ")

			_, err = msg.Discord.Sess.GuildRoleEdit(g.ID, oldRole.ID, newName, oldRole.Color, oldRole.Hoist, oldRole.Permissions, oldRole.Mentionable)
			if err != nil {
				if strings.Contains(err.Error(), strconv.Itoa(discordgo.ErrCodeMissingPermissions)) {
					_, _ = msg.ReplyEmbed(&discordgo.MessageEmbed{Description: "Missing permissions.", Color: utils.ColorCritical})
					return
				}
				_, _ = msg.ReplyEmbed(&discordgo.MessageEmbed{Description: "Some error occurred: `" + err.Error() + "`.", Color: utils.ColorCritical})
				return
			}

			embed := &discordgo.MessageEmbed{
				Color:       oldRole.Color,
				Description: fmt.Sprintf("Role name changed from %v to %v", oldRole.Name, newName),
			}
			_, _ = msg.ReplyEmbed(embed)

		} else if msg.Args()[1] == "color" {
			clr := msg.Args()[2]
			if strings.HasPrefix(clr, "#") {
				clr = clr[1:]
			}

			color, err := strconv.ParseInt(clr, 16, 64)
			if err != nil || color < 0 || color > 0xFFFFFF {
				_, _ = msg.ReplyEmbed(&discordgo.MessageEmbed{Description: "Invalid color code.", Color: utils.ColorCritical})
				return
			}

			_, err = msg.Discord.Sess.GuildRoleEdit(g.ID, oldRole.ID, oldRole.Name, int(color), oldRole.Hoist, oldRole.Permissions, oldRole.Mentionable)
			if err != nil {
				_, _ = msg.ReplyEmbed(&discordgo.MessageEmbed{Description: "Some error occurred: `" + err.Error(), Color: utils.ColorCritical})
				return
			}

			embed := &discordgo.MessageEmbed{
				Color: int(color),
				//Description: fmt.Sprintf("Color changed from #%v to #%v", fmt.Sprintf("%06X", oldRole.Color), fmt.Sprintf("%06X", color)),
				Description: fmt.Sprintf("Color changed from #%v to #%v", strconv.FormatInt(int64(oldRole.Color), 16), strconv.FormatInt(color, 16)), // fmt.Sprintf("%06X", color)),
			}
			_, _ = msg.ReplyEmbed(embed)
		}
		return
	case la == 1:
		target = msg.Member()
	case la == 2:
		target, err = msg.GetMemberAtArg(1)
		if err != nil {
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
		m.Log.Error("error fetching user role", zap.Error(err))
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
		_, _ = msg.Reply("the custom role is broken, wait for someone to fix it or try setting a new userrole")
		return
	}

	embed := helpers.NewEmbed().
		WithTitle(fmt.Sprintf("Custom role for %v", target.User.String())).
		WithColor(customRole.Color).
		AddField("Name", customRole.Name, true).
		AddField("Color", fmt.Sprintf("#"+fmt.Sprintf("%06X", customRole.Color)), true)
	_, _ = msg.ReplyEmbed(embed.Build())
}

func NewListCustomRolesCommand(m *Module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:           m,
		Name:          "listcustomroles",
		Description:   "Returns a list of custom roles for the server. It also shows whether users with custom roles are in the server or not",
		Triggers:      []string{"m?listuserroles", "m?listcustomroles"},
		Usage:         "m?listcustomroles",
		Cooldown:      30,
		CooldownUser:  false,
		RequiredPerms: discordgo.PermissionManageRoles,
		RequiresOwner: false,
		CheckBotPerms: false,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run: func(msg *mio.DiscordMessage) {
			roles, err := m.db.GetCustomRolesByGuild(msg.GuildID())
			if err != nil {
				_, _ = msg.Reply("There was an issue, please try again")
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
