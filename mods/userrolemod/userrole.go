package userrolemod

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/andersfylling/disgord"
	"github.com/intrntsrfr/meidov2"
	"github.com/jmoiron/sqlx"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

type UserRoleMod struct {
	cl       chan *meidov2.DiscordMessage
	commands []func(msg *meidov2.DiscordMessage)
	db       *sqlx.DB
}

func New() meidov2.Mod {
	return &UserRoleMod{
		//cl: make(chan *meidov2.DiscordMessage),
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

func (m *UserRoleMod) Hook(b *meidov2.Bot, cl chan *meidov2.DiscordMessage) error {
	m.cl = cl

	psql, err := sqlx.Connect("postgres", b.Config.ConnectionString)
	if err != nil {
		panic(err)
	}
	m.db = psql
	fmt.Println("psql connection for userroles established")

	m.commands = append(m.commands, m.ToggleUserRole, m.MyRole)
	//m.commands = append(m.commands, m.check)

	return nil
}

func (m *UserRoleMod) Message(msg *meidov2.DiscordMessage) {
	if msg.Message.IsDirectMessage() {
		return
	}
	for _, c := range m.commands {
		go c(msg)
	}
}

func (m *UserRoleMod) ToggleUserRole(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 3 || msg.Args()[0] != "m?setuserrole" {
		return
	}

	m.cl <- msg

	var (
		err          error
		targetUser   *disgord.Member
		selectedRole *disgord.Role
	)

	cu, err := msg.Discord.Client.GetCurrentUser(context.Background())
	if err != nil {
		return
	}
	botPerms, err := msg.Discord.Client.GetMemberPermissions(context.Background(), msg.Message.GuildID, cu.ID)
	if err != nil {
		fmt.Println(err)
		return
	}
	if botPerms&disgord.PermissionManageRoles == 0 && botPerms&disgord.PermissionAdministrator == 0 {
		return
	}

	uPerms, err := msg.Discord.Client.GetMemberPermissions(context.Background(), msg.Message.GuildID, msg.Message.Author.ID)
	if err != nil {
		fmt.Println(err)
		return
	}
	if uPerms&disgord.PermissionManageRoles == 0 && uPerms&disgord.PermissionAdministrator == 0 {
		return
	}

	if len(msg.Message.Mentions) >= 1 {
		targetUser, err = msg.Discord.Client.GetMember(context.Background(), msg.Message.GuildID, msg.Message.Mentions[0].ID)
		if err != nil {
			//s.ChannelMessageSend(ch.ID, err.Error())
			return
		}
	} else {
		id, err := strconv.Atoi(msg.Args()[1])
		if err != nil {
			return
		}
		targetUser, err = msg.Discord.Client.GetMember(context.Background(), msg.Message.GuildID, disgord.Snowflake(id))
		if err != nil {
			//s.ChannelMessageSend(ch.ID, err.Error())
			return
		}
	}
	if targetUser.User.Bot {
		msg.Reply("Bots dont get to join the fun")
		return
	}

	g, err := msg.Discord.Client.GetGuild(context.Background(), msg.Message.GuildID)
	if err != nil {
		msg.Reply(err.Error())
		return
	}

	for i := range g.Roles {
		role := g.Roles[i]

		roleID, err := strconv.Atoi(msg.Args()[2])
		if err != nil {
			if strings.ToLower(role.Name) == strings.ToLower(strings.Join(msg.Args()[2:], " ")) {
				selectedRole = role
			}
		} else {
			if role.ID == disgord.Snowflake(roleID) {
				selectedRole = role
			}
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
		if selectedRole.ID == disgord.Snowflake(userRole.RoleID) {
			m.db.Exec("DELETE FROM userroles WHERE guild_id=$1 AND user_id=$2 AND role_id=$3;", g.ID, targetUser.User.ID, selectedRole.ID)
			msg.Reply(fmt.Sprintf("Unbound role **%v** from user **%v**", selectedRole.Name, targetUser.User.Tag()))
		} else {
			m.db.Exec("UPDATE userroles SET role_id=$1 WHERE guild_id=$2 AND user_id=$3", selectedRole.ID, g.ID, targetUser.User.ID)
			msg.Reply(fmt.Sprintf("Updated userrole for **%v** to **%v**", targetUser.User.String(), selectedRole.Name))
		}
	case sql.ErrNoRows:
		m.db.Exec("INSERT INTO userroles(guild_id, user_id, role_id) VALUES($1, $2, $3);", g.ID, targetUser.User.ID, selectedRole.ID)
		msg.Reply(fmt.Sprintf("Bound role **%v** to user **%v**", selectedRole.Name, targetUser.User.Tag()))
	default:
		fmt.Println(err)
		msg.Reply("there was an error, please try again")
	}
}

func (m *UserRoleMod) MyRole(msg *meidov2.DiscordMessage) {

	if msg.LenArgs() < 1 || msg.Args()[0] != "m?myrole" {
		return
	}

	m.cl <- msg

	var (
		err     error
		oldRole *disgord.Role
		target  *disgord.Member
	)

	cu, err := msg.Discord.Client.GetCurrentUser(context.Background())
	if err != nil {
		return
	}
	botPerms, err := msg.Discord.Client.GetMemberPermissions(context.Background(), msg.Message.GuildID, cu.ID)
	if err != nil {
		fmt.Println(err)
		return
	}
	if botPerms&disgord.PermissionManageRoles == 0 && botPerms&disgord.PermissionAdministrator == 0 {
		return
	}

	switch la := msg.LenArgs(); {
	case la > 2:

		ur := &Userrole{}
		err = m.db.Get(ur, "SELECT * FROM userroles WHERE guild_id=$1 AND user_id=$2", msg.Message.GuildID, msg.Message.Author.ID)
		if err != nil && err != sql.ErrNoRows {
			fmt.Println(err)
			msg.Reply("there was an error, please try again")
			return
		} else if err == sql.ErrNoRows {
			msg.Reply("No custom role set.")
			return
		}

		g, err := msg.Discord.Client.GetGuild(context.Background(), msg.Message.GuildID)
		if err != nil {
			msg.Reply("some error occurred")
			return
		}

		for _, role := range g.Roles {
			if role.ID == disgord.Snowflake(ur.RoleID) {
				oldRole = role
			}
		}
		if oldRole == nil {
			msg.Reply("couldnt find role")
			return
		}

		switch msg.Args()[1] {
		case "name":

			newName := strings.Join(msg.Args()[2:], " ")

			_, err = msg.Discord.Client.UpdateGuildRole(context.Background(), msg.Message.GuildID, oldRole.ID).SetName(newName).Execute()
			if err != nil {
				/*
					if strings.Contains(err.Error(), strconv.Itoa(discordgo.ErrCodeMissingPermissions)) {
						ctx.SendEmbed(&discordgo.MessageEmbed{Description: "Missing permissions.", Color: dColorRed})
						return
					}
				*/
				msg.Reply(&disgord.Embed{Description: "Some error occured: `" + err.Error() + "`.", Color: 0xC80000})
				return
			}

			embed := &disgord.Embed{
				Color:       int(oldRole.Color),
				Description: fmt.Sprintf("Role name changed from %v to %v", oldRole.Name, newName),
			}
			msg.Reply(embed)

		case "color":

			if strings.HasPrefix(msg.Args()[2], "#") {
				msg.Args()[2] = msg.Args()[2][1:]
			}

			color, err := strconv.ParseInt("0x"+msg.Args()[2], 0, 64)
			if err != nil {
				msg.Reply(&disgord.Embed{Description: "Invalid color code.", Color: 0xC80000})
				return
			}
			if color < 0 || color > 0xFFFFFF {
				msg.Reply(&disgord.Embed{Description: "Invalid color code.", Color: 0xC80000})
				return
			}

			_, err = msg.Discord.Client.UpdateGuildRole(context.Background(), msg.Message.GuildID, oldRole.ID).SetColor(uint(color)).Execute()
			if err != nil {
				msg.Reply(&disgord.Embed{Description: "Some error occured: `" + err.Error() + "`.", Color: 0xC80000})
				return
			}

			embed := &disgord.Embed{
				Color:       int(color),
				Description: fmt.Sprintf("Color changed from #%v to #%v", fmt.Sprintf("%06X", oldRole.Color), fmt.Sprintf("%06X", color)),
			}
			msg.Reply(embed)
		default:
		}

		return
	case la == 1:
		target, err = msg.Discord.Client.GetMember(context.Background(), msg.Message.GuildID, msg.Message.Author.ID)
		if err != nil {
			//s.ChannelMessageSend(ch.ID, err.Error())
			return
		}
	case la == 2:

		if len(msg.Message.Mentions) >= 1 {
			target, err = msg.Discord.Client.GetMember(context.Background(), msg.Message.GuildID, msg.Message.Mentions[0].ID)
			if err != nil {
				//s.ChannelMessageSend(ch.ID, err.Error())
				fmt.Println(err)
				return
			}
		} else {
			id, err := strconv.Atoi(msg.Args()[1])
			if err != nil {
				return
			}
			target, err = msg.Discord.Client.GetMember(context.Background(), msg.Message.GuildID, disgord.Snowflake(id))
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
	err = m.db.Get(ur, "SELECT * FROM userroles WHERE guild_id=$1 AND user_id=$2", msg.Message.GuildID, target.User.ID)
	if err != nil && err != sql.ErrNoRows {
		msg.Reply("there was an error, please try again")
		fmt.Println(err)
		return
	} else if err == sql.ErrNoRows {
		msg.Reply("No custom role set.")
		return
	}

	var customRole *disgord.Role

	g, err := msg.Discord.Client.GetGuild(context.Background(), msg.Message.GuildID)
	if err != nil {
		msg.Reply("error occurred")
		return
	}
	for i := range g.Roles {
		role := g.Roles[i]

		if role.ID == disgord.Snowflake(ur.RoleID) {
			customRole = role
		}
	}

	if customRole == nil {
		msg.Reply("the custom role is broken, wait for someone to fix it or try setting a new userrole")
		return
	}

	embed := &disgord.Embed{
		Color: int(customRole.Color),
		Title: fmt.Sprintf("Custom role for %v", target.User.Tag()),
		Fields: []*disgord.EmbedField{
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
	msg.Reply(embed)
}
