package userrolemod

import (
	"database/sql"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meidov2"
	"github.com/intrntsrfr/owo"
	"github.com/jmoiron/sqlx"
	"strconv"
	"strings"
	"sync"
)

type UserRoleMod struct {
	Name string
	sync.Mutex
	cl       chan *meidov2.DiscordMessage
	commands map[string]meidov2.ModCommand
	db       *sqlx.DB
	owo      *owo.Client
}

func New(name string) meidov2.Mod {
	return &UserRoleMod{
		Name:     name,
		commands: make(map[string]meidov2.ModCommand),
	}
}

func (m *UserRoleMod) Save() error {
	return nil
}

func (m *UserRoleMod) Load() error {
	return nil
}

func (m *UserRoleMod) Settings(msg *meidov2.DiscordMessage) {

}
func (m *UserRoleMod) Help(msg *meidov2.DiscordMessage) {

}

func (m *UserRoleMod) Commands() map[string]meidov2.ModCommand {
	return nil
}

func (m *UserRoleMod) Hook(b *meidov2.Bot) error {
	m.cl = b.CommandLog
	m.db = b.DB
	m.owo = b.Owo

	b.Discord.Sess.AddHandler(func(s *discordgo.Session, r *discordgo.GuildRoleDelete) {
		m.db.Exec("DELETE FROM userroles WHERE guild_id=$1 AND role_id=$2", r.GuildID, r.RoleID)
	})

	m.RegisterCommand(NewToggleUserRoleCommand(m))
	m.RegisterCommand(NewMyRoleCommand(m))
	m.RegisterCommand(NewListUserRolesCommand(m))

	return nil
}

func (m *UserRoleMod) RegisterCommand(cmd meidov2.ModCommand) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.commands[cmd.Name()]; ok {
		panic(fmt.Sprintf("command '%v' already exists in %v", cmd.Name(), m.Name))
	}
	m.commands[cmd.Name()] = cmd
}

func (m *UserRoleMod) Message(msg *meidov2.DiscordMessage) {
	if msg.IsDM() {
		return
	}
	if msg.Type != meidov2.MessageTypeCreate {
		return
	}
	for _, c := range m.commands {
		go c.Run(msg)
	}
}

type ToggleUserRoleCommand struct {
	m       *UserRoleMod
	Enabled bool
}

func NewToggleUserRoleCommand(m *UserRoleMod) meidov2.ModCommand {
	return &ToggleUserRoleCommand{
		m:       m,
		Enabled: true,
	}
}
func (c *ToggleUserRoleCommand) Name() string {
	return "Userrole"
}
func (c *ToggleUserRoleCommand) Description() string {
	return "Binds, unbinds or changes a userrole bind to a user"
}
func (c *ToggleUserRoleCommand) Triggers() []string {
	return []string{"m?setuserrole"}
}
func (c *ToggleUserRoleCommand) Usage() string {
	return "m?setuserrole 1231231231231 cool role"
}
func (c *ToggleUserRoleCommand) Cooldown() int {
	return 30
}
func (c *ToggleUserRoleCommand) RequiredPerms() int {
	return discordgo.PermissionManageRoles
}
func (c *ToggleUserRoleCommand) RequiresOwner() bool {
	return false
}
func (c *ToggleUserRoleCommand) IsEnabled() bool {
	return c.Enabled
}
func (c *ToggleUserRoleCommand) Run(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 3 || msg.Args()[0] != "m?setuserrole" {
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

	botPerms, err := msg.Discord.Sess.State.UserChannelPermissions(msg.Discord.Sess.State.User.ID, msg.Message.ChannelID)
	if err != nil {
		fmt.Println(err)
		return
	}
	if botPerms&discordgo.PermissionManageRoles == 0 && botPerms&discordgo.PermissionAdministrator == 0 {
		return
	}

	c.m.cl <- msg

	var (
		targetUser   *discordgo.Member
		selectedRole *discordgo.Role
	)

	if len(msg.Message.Mentions) >= 1 {
		targetUser, err = msg.Discord.Sess.GuildMember(msg.Message.GuildID, msg.Message.Mentions[0].ID)
		if err != nil {
			//s.ChannelMessageSend(ch.ID, err.Error())
			return
		}
	} else {
		targetUser, err = msg.Discord.Sess.GuildMember(msg.Message.GuildID, msg.Args()[1])
		if err != nil {
			//s.ChannelMessageSend(ch.ID, err.Error())
			return
		}
	}
	if targetUser.User.Bot {
		msg.Reply("Bots dont get to join the fun")
		return
	}

	g, err := msg.Discord.Sess.State.Guild(msg.Message.GuildID)
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

	err = c.m.db.Get(userRole, "SELECT * FROM userroles WHERE guild_id=$1 AND user_id=$2", g.ID, targetUser.User.ID)
	switch err {
	case nil:
		if selectedRole.ID == userRole.RoleID {
			c.m.db.Exec("DELETE FROM userroles WHERE guild_id=$1 AND user_id=$2 AND role_id=$3;", g.ID, targetUser.User.ID, selectedRole.ID)
			msg.Reply(fmt.Sprintf("Unbound role **%v** from user **%v**", selectedRole.Name, targetUser.User.String()))
		} else {
			c.m.db.Exec("UPDATE userroles SET role_id=$1 WHERE guild_id=$2 AND user_id=$3", selectedRole.ID, g.ID, targetUser.User.ID)
			msg.Reply(fmt.Sprintf("Updated userrole for **%v** to **%v**", targetUser.User.String(), selectedRole.Name))
		}
	case sql.ErrNoRows:
		c.m.db.Exec("INSERT INTO userroles(guild_id, user_id, role_id) VALUES($1, $2, $3);", g.ID, targetUser.User.ID, selectedRole.ID)
		msg.Reply(fmt.Sprintf("Bound role **%v** to user **%v**", selectedRole.Name, targetUser.User.String()))
	default:
		fmt.Println(err)
		msg.Reply("there was an error, please try again")
	}
}

type MyRoleCommand struct {
	m       *UserRoleMod
	Enabled bool
}

func NewMyRoleCommand(m *UserRoleMod) meidov2.ModCommand {
	return &MyRoleCommand{
		m:       m,
		Enabled: true,
	}
}
func (c *MyRoleCommand) Name() string {
	return "MyRole"
}
func (c *MyRoleCommand) Description() string {
	return "Displays a users bound role, or lets the user change the name or color of their bound role"
}
func (c *MyRoleCommand) Triggers() []string {
	return []string{"m?myrole"}
}
func (c *MyRoleCommand) Usage() string {
	return "m?myrole\nm?myrole 123123123123\nm?myrole color c0ffee\nm?myrole name jeff"
}
func (c *MyRoleCommand) Cooldown() int {
	return 5
}
func (c *MyRoleCommand) RequiredPerms() int {
	return 0
}
func (c *MyRoleCommand) RequiresOwner() bool {
	return false
}
func (c *MyRoleCommand) IsEnabled() bool {
	return c.Enabled
}
func (c *MyRoleCommand) Run(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 1 || msg.Args()[0] != "m?myrole" {
		return
	}

	c.m.cl <- msg

	var (
		err     error
		oldRole *discordgo.Role
		target  *discordgo.Member
	)

	g, err := msg.Discord.Sess.State.Guild(msg.Message.GuildID)
	if err != nil {
		msg.Reply("some error occurred")
		return
	}

	switch la := msg.LenArgs(); {
	case la > 2:
		if msg.Args()[1] != "name" && msg.Args()[1] != "color" {
			return
		}

		botPerms, err := msg.Discord.Sess.State.UserChannelPermissions(msg.Discord.Sess.State.User.ID, msg.Message.ChannelID)
		if err != nil {
			return
		}
		if botPerms&discordgo.PermissionManageRoles == 0 && botPerms&discordgo.PermissionAdministrator == 0 {
			return
		}

		ur := &Userrole{}
		err = c.m.db.Get(ur, "SELECT * FROM userroles WHERE guild_id=$1 AND user_id=$2", g.ID, msg.Message.Author.ID)
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
			newName := strings.Join(msg.Content()[2:], " ")

			_, err = msg.Discord.Sess.GuildRoleEdit(g.ID, oldRole.ID, newName, oldRole.Color, oldRole.Hoist, oldRole.Permissions, oldRole.Mentionable)
			if err != nil {
				if strings.Contains(err.Error(), strconv.Itoa(discordgo.ErrCodeMissingPermissions)) {
					msg.ReplyEmbed(&discordgo.MessageEmbed{Description: "Missing permissions.", Color: 0xC80000})
					return
				}
				msg.ReplyEmbed(&discordgo.MessageEmbed{Description: "Some error occured: `" + err.Error() + "`.", Color: 0xC80000})
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

			_, err = msg.Discord.Sess.GuildRoleEdit(g.ID, oldRole.ID, oldRole.Name, int(color), oldRole.Hoist, oldRole.Permissions, oldRole.Mentionable)
			if err != nil {
				msg.ReplyEmbed(&discordgo.MessageEmbed{Description: "Some error occured: `" + err.Error(), Color: 0xC80000})
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
			target, err = msg.Discord.Sess.GuildMember(g.ID, msg.Message.Mentions[0].ID)
			if err != nil {
				//s.ChannelMessageSend(ch.ID, err.Error())
				fmt.Println(err)
				return
			}
		} else {
			target, err = msg.Discord.Sess.GuildMember(g.ID, msg.Args()[1])
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
	err = c.m.db.Get(ur, "SELECT * FROM userroles WHERE guild_id=$1 AND user_id=$2", g.ID, target.User.ID)
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

type ListUserRolesCommand struct {
	m       *UserRoleMod
	Enabled bool
}

func NewListUserRolesCommand(m *UserRoleMod) meidov2.ModCommand {
	return &ListUserRolesCommand{
		m:       m,
		Enabled: true,
	}
}
func (c *ListUserRolesCommand) Name() string {
	return "ListUserRoles"
}
func (c *ListUserRolesCommand) Description() string {
	return "Returns a list of the user roles that are in the server, displays if some users still are in the server or not"
}
func (c *ListUserRolesCommand) Triggers() []string {
	return []string{"m?listuserroles"}
}
func (c *ListUserRolesCommand) Usage() string {
	return "m?listuserroles"
}
func (c *ListUserRolesCommand) Cooldown() int {
	return 30
}
func (c *ListUserRolesCommand) RequiredPerms() int {
	return 0
}
func (c *ListUserRolesCommand) RequiresOwner() bool {
	return false
}
func (c *ListUserRolesCommand) IsEnabled() bool {
	return c.Enabled
}
func (c *ListUserRolesCommand) Run(msg *meidov2.DiscordMessage) {

	if msg.LenArgs() != 1 || msg.Args()[0] != "m?listuserroles" {
		return
	}
	c.m.cl <- msg

	var userRoles []*Userrole

	err := c.m.db.Select(&userRoles, "SELECT role_id, user_id FROM userroles WHERE guild_id=$1;", msg.Message.GuildID)
	if err != nil {
		msg.Reply("there was an error, please try again")
		return
	}

	g, err := msg.Discord.Sess.State.Guild(msg.Message.GuildID)
	if err != nil {
		msg.Reply("some error occurred, please try again")
		return
	}

	text := fmt.Sprintf("Userroles in %v\n\n", g.Name)
	count := 0
	for _, ur := range userRoles {
		role, err := msg.Sess.State.Role(g.ID, ur.RoleID)
		if err != nil {
			continue
		}

		mem, err := msg.Sess.State.Member(g.ID, ur.UserID)
		if err != nil {
			text += fmt.Sprintf("Role #%v: %v (%v) | Bound user: %v - User no longer in guild.\n", count, role.Name, role.ID, ur.UserID)
		} else {
			text += fmt.Sprintf("Role #%v: %v (%v) | Bound user: %v (%v)\n", count, role.Name, role.ID, mem.User.String(), mem.User.ID)
		}
		count++
	}

	link, err := c.m.owo.Upload(text)
	if err != nil {
		msg.Reply("Error getting user roles.")
		return
	}
	msg.Reply(fmt.Sprintf("User roles in %v\n%v", g.Name, link))
}
