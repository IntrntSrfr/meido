package userrolemod

import (
	"database/sql"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido"
	"github.com/intrntsrfr/owo"
	"github.com/jmoiron/sqlx"
	"strconv"
	"strings"
	"sync"
	"time"
)

type UserRoleMod struct {
	sync.Mutex
	name         string
	commands     map[string]*meido.ModCommand
	db           *sqlx.DB
	owo          *owo.Client
	allowedTypes meido.MessageType
	allowDMs     bool
}

func New(name string) meido.Mod {
	return &UserRoleMod{
		name:         name,
		commands:     make(map[string]*meido.ModCommand),
		allowedTypes: meido.MessageTypeCreate,
		allowDMs:     false,
	}
}
func (m *UserRoleMod) Name() string {
	return m.name
}
func (m *UserRoleMod) Save() error {
	return nil
}
func (m *UserRoleMod) Load() error {
	return nil
}
func (m *UserRoleMod) Passives() []*meido.ModPassive {
	return []*meido.ModPassive{}
}
func (m *UserRoleMod) Commands() map[string]*meido.ModCommand {
	return m.commands
}
func (m *UserRoleMod) AllowedTypes() meido.MessageType {
	return m.allowedTypes
}
func (m *UserRoleMod) AllowDMs() bool {
	return m.allowDMs
}
func (m *UserRoleMod) Hook(b *meido.Bot) error {
	m.db = b.DB
	m.owo = b.Owo

	b.Discord.Sess.AddHandler(func(s *discordgo.Session, r *discordgo.GuildRoleDelete) {
		m.db.Exec("DELETE FROM userroles WHERE guild_id=$1 AND role_id=$2", r.GuildID, r.RoleID)
	})

	b.Discord.Sess.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		refreshTicker := time.NewTicker(time.Hour)

		go func() {
			for range refreshTicker.C {
				for _, g := range b.Discord.Guilds() {
					if g.Unavailable {
						continue
					}

					var userRoles []*Userrole

					err := b.DB.Get(&userRoles, "SELECT * FROM userroles WHERE guild_id=$1", g.ID)
					if err != nil {
						continue
					}

					for _, ur := range userRoles {
						hasRole := false

						for _, r := range g.Roles {
							if r.ID == ur.RoleID {
								hasRole = true
							}
						}

						if !hasRole {
							b.DB.Exec("DELETE FROM userroles WHERE uid=$1", ur.UID)
						}
					}
				}
			}
		}()
	})

	m.RegisterCommand(NewSetUserRoleCommand(m))
	m.RegisterCommand(NewMyRoleCommand(m))
	m.RegisterCommand(NewListUserRolesCommand(m))

	return nil
}
func (m *UserRoleMod) RegisterCommand(cmd *meido.ModCommand) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.commands[cmd.Name]; ok {
		panic(fmt.Sprintf("command '%v' already exists in %v", cmd.Name, m.Name()))
	}
	m.commands[cmd.Name] = cmd
}

func NewSetUserRoleCommand(m *UserRoleMod) *meido.ModCommand {
	return &meido.ModCommand{
		Mod:           m,
		Name:          "setuserrole",
		Description:   "Binds, unbinds or changes a userrole bind to a user",
		Triggers:      []string{"m?setuserrole"},
		Usage:         "m?setuserrole 1231231231231 cool role",
		Cooldown:      3,
		RequiredPerms: discordgo.PermissionManageRoles,
		RequiresOwner: false,
		AllowedTypes:  meido.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           m.setuserroleCommand,
	}
}

func (m *UserRoleMod) setuserroleCommand(msg *meido.DiscordMessage) {
	if msg.LenArgs() < 3 {
		return
	}

	var (
		targetUser   *discordgo.Member
		selectedRole *discordgo.Role
		err          error
	)

	if len(msg.Message.Mentions) >= 1 {
		targetUser, err = msg.Discord.Member(msg.Message.GuildID, msg.Message.Mentions[0].ID)
		if err != nil {
			//s.ChannelMessageSend(ch.ID, err.Error())
			return
		}
	} else {
		targetUser, err = msg.Discord.Member(msg.Message.GuildID, msg.Args()[1])
		if err != nil {
			//s.ChannelMessageSend(ch.ID, err.Error())
			return
		}
	}
	if targetUser.User.Bot {
		msg.Reply("Bots dont get to join the fun")
		return
	}

	g, err := msg.Discord.Guild(msg.Message.GuildID)
	if err != nil {
		msg.Reply(err.Error())
		return
	}

	for i := range g.Roles {
		role := g.Roles[i]

		if role.ID == msg.Args()[2] {
			selectedRole = role
		} else if strings.ToLower(role.Name) == strings.ToLower(strings.Join(msg.Args()[2:], " ")) {
			selectedRole = role
		}
	}

	if selectedRole == nil {
		msg.Reply("role not found")
		return
	}

	userRole := &Userrole{}

	err = m.db.Get(userRole, "SELECT * FROM userroles WHERE guild_id=$1 AND user_id=$2", g.ID, targetUser.User.ID)
	switch err {
	case nil:
		if selectedRole.ID == userRole.RoleID {
			m.db.Exec("DELETE FROM userroles WHERE guild_id=$1 AND user_id=$2 AND role_id=$3;", g.ID, targetUser.User.ID, selectedRole.ID)
			msg.Reply(fmt.Sprintf("Unbound role **%v** from user **%v**", selectedRole.Name, targetUser.User.String()))
		} else {
			m.db.Exec("UPDATE userroles SET role_id=$1 WHERE guild_id=$2 AND user_id=$3", selectedRole.ID, g.ID, targetUser.User.ID)
			msg.Reply(fmt.Sprintf("Updated userrole for **%v** to **%v**", targetUser.User.String(), selectedRole.Name))
		}
	case sql.ErrNoRows:
		m.db.Exec("INSERT INTO userroles(guild_id, user_id, role_id) VALUES($1, $2, $3);", g.ID, targetUser.User.ID, selectedRole.ID)
		msg.Reply(fmt.Sprintf("Bound role **%v** to user **%v**", selectedRole.Name, targetUser.User.String()))
	default:
		fmt.Println(err)
		msg.Reply("there was an error, please try again")
	}
}

func NewMyRoleCommand(m *UserRoleMod) *meido.ModCommand {
	return &meido.ModCommand{
		Mod:           m,
		Name:          "myrole",
		Description:   "Displays a users bound role, or lets the user change the name or color of their bound role",
		Triggers:      []string{"m?myrole"},
		Usage:         "m?myrole\nm?myrole 123123123123\nm?myrole color c0ffee\nm?myrole name jeff",
		Cooldown:      3,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  meido.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           m.myroleCommand,
	}
}

func (m *UserRoleMod) myroleCommand(msg *meido.DiscordMessage) {
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
		msg.Reply("some error occurred")
		return
	}

	switch la := msg.LenArgs(); {
	case la > 2:
		if msg.Args()[1] != "name" && msg.Args()[1] != "color" {
			return
		}

		if !msg.Discord.HasPermissions(msg.Message.ChannelID, discordgo.PermissionManageRoles) {
			return
		}

		/*
			botPerms, err := msg.Discord.UserChannelPermissions(msg.Discord.Sess.State.User.ID, msg.Message.ChannelID)
			if err != nil {
				return
			}
			if botPerms&discordgo.PermissionManageRoles == 0 && botPerms&discordgo.PermissionAdministrator == 0 {
				return
			}
		*/

		ur := &Userrole{}
		err = m.db.Get(ur, "SELECT * FROM userroles WHERE guild_id=$1 AND user_id=$2", g.ID, msg.Message.Author.ID)
		if err != nil && err != sql.ErrNoRows {
			fmt.Println(err)
			msg.Reply("there was an error, please try again")
			return
		} else if err == sql.ErrNoRows {
			msg.Reply("No custom role set.")
			return
		}

		for _, role := range g.Roles {
			if role.ID == ur.RoleID {
				oldRole = role
			}
		}
		if oldRole == nil {
			msg.Reply("couldnt find role")
			return
		}

		if msg.Args()[1] == "name" {
			newName := strings.Join(msg.RawArgs()[2:], " ")

			_, err = msg.Discord.Sess.GuildRoleEdit(g.ID, oldRole.ID, newName, oldRole.Color, oldRole.Hoist, int(oldRole.Permissions), oldRole.Mentionable)
			if err != nil {
				if strings.Contains(err.Error(), strconv.Itoa(discordgo.ErrCodeMissingPermissions)) {
					msg.ReplyEmbed(&discordgo.MessageEmbed{Description: "Missing permissions.", Color: 0xC80000})
					return
				}
				msg.ReplyEmbed(&discordgo.MessageEmbed{Description: "Some error occurred: `" + err.Error() + "`.", Color: 0xC80000})
				return
			}

			embed := &discordgo.MessageEmbed{
				Color:       oldRole.Color,
				Description: fmt.Sprintf("Role name changed from %v to %v", oldRole.Name, newName),
			}
			msg.ReplyEmbed(embed)

		} else if msg.Args()[1] == "color" {
			clr := msg.Args()[2]
			if strings.HasPrefix(clr, "#") {
				clr = clr[1:]
			}

			color, err := strconv.ParseInt(clr, 16, 64)
			if err != nil {
				msg.ReplyEmbed(&discordgo.MessageEmbed{Description: "Invalid color code.", Color: 0xC80000})
				return
			}
			if color < 0 || color > 0xFFFFFF {
				msg.ReplyEmbed(&discordgo.MessageEmbed{Description: "Invalid color code.", Color: 0xC80000})
				return
			}

			_, err = msg.Discord.Sess.GuildRoleEdit(g.ID, oldRole.ID, oldRole.Name, int(color), oldRole.Hoist, int(oldRole.Permissions), oldRole.Mentionable)
			if err != nil {
				msg.ReplyEmbed(&discordgo.MessageEmbed{Description: "Some error occurred: `" + err.Error(), Color: 0xC80000})
				return
			}

			embed := &discordgo.MessageEmbed{
				Color: int(color),
				//Description: fmt.Sprintf("Color changed from #%v to #%v", fmt.Sprintf("%06X", oldRole.Color), fmt.Sprintf("%06X", color)),
				Description: fmt.Sprintf("Color changed from #%v to #%v", strconv.FormatInt(int64(oldRole.Color), 16), strconv.FormatInt(color, 16)), // fmt.Sprintf("%06X", color)),
			}
			msg.ReplyEmbed(embed)
		}
		return
	case la == 1:
		target = msg.Member
	case la == 2:
		if len(msg.Message.Mentions) >= 1 {
			target, err = msg.Discord.Member(g.ID, msg.Message.Mentions[0].ID)
			if err != nil {
				//s.ChannelMessageSend(ch.ID, err.Error())
				fmt.Println(err)
				return
			}
		} else {
			target, err = msg.Discord.Member(g.ID, msg.Args()[1])
			if err != nil {
				//s.ChannelMessageSend(ch.ID, err.Error())
				fmt.Println(err)
				return
			}
		}
	default:
		return
	}

	if target == nil {
		return
	}

	ur := &Userrole{}
	err = m.db.Get(ur, "SELECT * FROM userroles WHERE guild_id=$1 AND user_id=$2", g.ID, target.User.ID)
	if err != nil && err != sql.ErrNoRows {
		msg.Reply("there was an error, please try again")
		fmt.Println(err)
		return
	} else if err == sql.ErrNoRows {
		msg.Reply("No custom role set.")
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
		msg.Reply("the custom role is broken, wait for someone to fix it or try setting a new userrole")
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
	msg.ReplyEmbed(embed)
}

func NewListUserRolesCommand(m *UserRoleMod) *meido.ModCommand {
	return &meido.ModCommand{
		Mod:           m,
		Name:          "listuserroles",
		Description:   "Returns a list of the user roles that are in the server, displays if some users still are in the server or not",
		Triggers:      []string{"m?listuserroles"},
		Usage:         "m?listuserroles",
		Cooldown:      30,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  meido.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           m.listuserrolesCommand,
	}
}

func (m *UserRoleMod) listuserrolesCommand(msg *meido.DiscordMessage) {
	if msg.LenArgs() != 1 {
		return
	}

	var userRoles []*Userrole

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
		role, err := msg.Discord.Role(g.ID, ur.RoleID)
		if err != nil {
			continue
		}

		mem, err := msg.Discord.Member(g.ID, ur.UserID)
		if err != nil {
			text += fmt.Sprintf("Role #%v: %v (%v) | Bound user: %v - User no longer in guild.\n", count, role.Name, role.ID, ur.UserID)
		} else {
			text += fmt.Sprintf("Role #%v: %v (%v) | Bound user: %v (%v)\n", count, role.Name, role.ID, mem.User.String(), mem.User.ID)
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
