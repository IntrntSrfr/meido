package userrolemod

import (
	"database/sql"
	"fmt"
	"github.com/intrntsrfr/meido/internal/database"
	"github.com/intrntsrfr/meido/pkg/mio"
	"github.com/intrntsrfr/meido/pkg/utils"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/owo"
)

type UserRoleMod struct {
	*mio.ModuleBase
	db  database.DB
	owo *owo.Client
}

func New(b *mio.Bot, db *database.PsqlDB, owo *owo.Client, log *zap.Logger) mio.Module {
	return &UserRoleMod{
		ModuleBase: mio.NewModule(b, "UserRoles", log),
		db:         db,
		owo:        owo,
	}
}

func (m *UserRoleMod) Hook() error {
	m.Bot.Discord.Sess.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		refreshTicker := time.NewTicker(time.Hour)

		go func() {
			for range refreshTicker.C {
				for _, g := range m.Bot.Discord.Guilds() {
					if g.Unavailable {
						continue
					}

					roles, err := m.db.GetMemberRolesByGuild(g.ID)
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

						if !hasRole {
							if err := m.db.DeleteMemberRole(ur.UID); err != nil {
								m.Log.Error("could not delete member role",
									zap.Int("member role ID", ur.UID),
									zap.String("guild id", ur.GuildID),
									zap.String("role id", ur.RoleID),
									zap.String("user id", ur.UserID))
							}
						}
					}
				}
			}
		}()
	})

	return m.RegisterCommands([]*mio.ModuleCommand{
		NewSetUserRoleCommand(m),
		NewMyRoleCommand(m),
	})
}

func NewSetUserRoleCommand(m *UserRoleMod) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:           m,
		Name:          "setuserrole",
		Description:   "Binds, unbinds or changes a userrole bind to a user",
		Triggers:      []string{"m?setuserrole"},
		Usage:         "m?setuserrole 1231231231231 cool role",
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
				_, _ = msg.Reply("could not find that user")
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
				_, _ = msg.Reply("Could not find that role!")
				return
			}

			if ur, err := m.db.GetMemberRole(g.ID, targetMember.User.ID); err == nil {
				if selectedRole.ID == ur.RoleID {
					return
				}
				ur.RoleID = selectedRole.ID
				if err := m.db.UpdateMemberRole(ur); err != nil {
					_, _ = msg.Reply("Could not set role, please try again!")
					return
				}
				_, _ = msg.Reply(fmt.Sprintf("Updated member role for **%v** to **%v**", targetMember.User.String(), selectedRole.Name))
				return
			}

			if err := m.db.CreateMemberRole(g.ID, targetMember.User.ID, selectedRole.ID); err != nil {
				_, _ = msg.Reply("Could not set role, please try again!")
				return
			}
			_, _ = msg.Reply(fmt.Sprintf("Bound role **%v** to user **%v**", selectedRole.Name, targetMember.User.String()))

			/*
				//m.db.Exec("INSERT INTO user_role(guild_id, user_id, role_id) VALUES($1, $2, $3);", g.ID, targetMember.User.ID, selectedRole.ID)
				userRole := &database.UserRole{}
				err = m.db.Get(userRole, "SELECT * FROM user_role WHERE guild_id=$1 AND user_id=$2", g.ID, targetMember.User.ID)
				switch err {
				case nil:
					if selectedRole.ID == userRole.RoleID {
						m.db.Exec("DELETE FROM user_role WHERE guild_id=$1 AND user_id=$2 AND role_id=$3;", g.ID, targetMember.User.ID, selectedRole.ID)
						msg.Reply(fmt.Sprintf("Unbound role **%v** from user **%v**", selectedRole.Name, targetMember.User.String()))
					} else {
						m.db.Exec("UPDATE user_role SET role_id=$1 WHERE guild_id=$2 AND user_id=$3", selectedRole.ID, g.ID, targetMember.User.ID)
						msg.Reply(fmt.Sprintf("Updated userrole for **%v** to **%v**", targetMember.User.String(), selectedRole.Name))
					}
				case sql.ErrNoRows:
					m.db.Exec("INSERT INTO user_role(guild_id, user_id, role_id) VALUES($1, $2, $3);", g.ID, targetMember.User.ID, selectedRole.ID)
					msg.Reply(fmt.Sprintf("Bound role **%v** to user **%v**", selectedRole.Name, targetMember.User.String()))
				default:
					fmt.Println(err)
					msg.Reply("there was an error, please try again")
				}
			*/
		},
	}
}

func NewMyRoleCommand(m *UserRoleMod) *mio.ModuleCommand {
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

func (m *UserRoleMod) myroleCommand(msg *mio.DiscordMessage) {
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

		ur, err := m.db.GetMemberRole(msg.GuildID(), msg.AuthorID())
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

	ur, err := m.db.GetMemberRole(msg.GuildID(), target.User.ID)
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

	embed := &discordgo.MessageEmbed{
		Color: customRole.Color,
		Title: fmt.Sprintf("Custom role for %v", target.User.String()),
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Name",
				Value:  customRole.Name,
				Inline: true,
			},
			{
				Name:   "Color",
				Value:  fmt.Sprintf("#" + fmt.Sprintf("%06X", customRole.Color)),
				Inline: true,
			},
		},
	}
	_, _ = msg.ReplyEmbed(embed)
}

/*
func NewListUserRolesCommand(m *UserRoleMod) *base.ModuleCommand {
	return &base.ModuleCommand{
		Module:           m,
		Name:          "listuserroles",
		Description:   "Returns a list of the user roles that are in the server, displays if some users still are in the server or not",
		Triggers:      []string{"m?listuserroles"},
		Usage:         "m?listuserroles",
		Cooldown:      30,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  base.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           m.listuserrolesCommand,
	}
}

func (m *UserRoleMod) listuserrolesCommand(msg *base.DiscordMessage) {
	if msg.LenArgs() != 1 {
		return
	}

	var userRoles []*database.UserRole

	err := m.db.Select(&userRoles, "SELECT role_id, user_id FROM userroles WHERE guild_id=$1;", msg.Message.GuildID)
	if err != nil {
		msg.Reply("there was an error, please try again")
		return
	}

	g, err := msg.Discord.Guild(msg.Message.GuildID)
	if err != nil {
		msg.Reply("some error occurred, please try again")
		return
	}

	text := fmt.Sprintf("Userroles in %v\n\n", g.Name)
	count := 0
	for _, ur := range userRoles {
		role, err := msg.Discord.Role(g.UID, ur.RoleID)
		if err != nil {
			continue
		}

		mem, err := msg.Discord.Member(g.UID, ur.UserID)
		if err != nil {
			text += fmt.Sprintf("Role #%v: %v (%v) | Bound user: %v - User no longer in guild.\n", count, role.Name, role.UID, ur.UserID)
		} else {
			text += fmt.Sprintf("Role #%v: %v (%v) | Bound user: %v (%v)\n", count, role.Name, role.UID, mem.User.String(), mem.User.UID)
		}
		count++
	}

	link, err := m.owo.Upload(text)
	if err != nil {
		msg.Reply("Error getting user roles.")
		return
	}
	msg.Reply(fmt.Sprintf("User roles in %v\n%v", g.Name, link))
}
*/
