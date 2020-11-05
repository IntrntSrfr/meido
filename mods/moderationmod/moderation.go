package moderationmod

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/andersfylling/disgord"
	"github.com/intrntsrfr/meidov2"
	"github.com/jmoiron/sqlx"
	"strconv"
	"strings"
)

type ModerationMod struct {
	cl       chan *meidov2.DiscordMessage
	commands []func(msg *meidov2.DiscordMessage)
	db       *sqlx.DB
}

func New() meidov2.Mod {
	return &ModerationMod{
		//cl: make(chan *meidov2.DiscordMessage),
	}
}

func (m *ModerationMod) Save() error {
	return nil
}

func (m *ModerationMod) Load() error {
	return nil
}

func (m *ModerationMod) Settings(msg *meidov2.DiscordMessage) {

}

func (m *ModerationMod) Help(msg *meidov2.DiscordMessage) {

}
func (m *ModerationMod) Commands() []meidov2.ModCommand {
	return nil
}

func (m *ModerationMod) Hook(b *meidov2.Bot, db *sqlx.DB, cl chan *meidov2.DiscordMessage) error {
	m.cl = cl
	m.db = db

	b.Discord.Client.On(disgord.EvtGuildCreate, func(s disgord.Session, g *disgord.GuildCreate) {
		dbg := &DiscordGuild{}
		err := db.Get(dbg, "SELECT guild_id FROM guilds WHERE guild_id = $1;", g.Guild.ID)
		if err != nil && err != sql.ErrNoRows {
			fmt.Println(err)
		} else if err == sql.ErrNoRows {
			db.Exec("INSERT INTO guilds(guild_id, use_strikes, max_strikes) VALUES($1, $2, $3)", g.Guild.ID, false, 3)
			fmt.Println(fmt.Sprintf("Inserted new guild: %v [%v]", g.Guild.Name, g.Guild.ID))
		}
	})

	m.commands = append(m.commands, m.Ban, m.Unban, m.Hackban)
	m.commands = append(m.commands, m.Warn, m.WarnLog, m.RemoveWarn, m.ClearWarns)
	m.commands = append(m.commands, m.CheckFilter, m.FilterWord, m.ClearFilter, m.FilterWordsList)
	m.commands = append(m.commands, m.SetMaxStrikes, m.ToggleStrikes)

	return nil
}

func (m *ModerationMod) Message(msg *meidov2.DiscordMessage) {
	// moderation only is for servers, so dms are ignored
	if msg.Message.IsDirectMessage() {
		return
	}
	if msg.Type == meidov2.MessageTypeDelete {
		return
	}
	for _, c := range m.commands {
		go c(msg)
	}
}

func (m *ModerationMod) Ban(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 2 || (msg.Args()[0] != ".ban" && msg.Args()[0] != ".b" && msg.Args()[0] != "m?ban" && msg.Args()[0] != "m?b") {
		return
	}
	if msg.Type != meidov2.MessageTypeCreate {
		return
	}

	cu, err := msg.Discord.Client.CurrentUser().Get()
	if err != nil {
		return
	}
	botPerms, err := msg.Discord.Client.Guild(msg.Message.GuildID).GetMemberPermissions(cu.ID)
	if err != nil {
		return
	}
	if botPerms&disgord.PermissionBanMembers == 0 && botPerms&disgord.PermissionAdministrator == 0 {
		return
	}

	uPerms, err := msg.Discord.Client.Guild(msg.Message.GuildID).GetMemberPermissions(msg.Message.Author.ID)
	if err != nil {
		return
	}
	if uPerms&disgord.PermissionBanMembers == 0 && uPerms&disgord.PermissionAdministrator == 0 {
		return
	}

	m.cl <- msg

	var (
		targetUser *disgord.User
		reason     string
		pruneDays  int
	)

	switch la := msg.LenArgs(); {
	case la == 2:
		pruneDays = 0
		reason = ""
	case la >= 3:
		pruneDays, err = strconv.Atoi(msg.Args()[2])
		if err != nil {
			pruneDays = 0
			reason = strings.Join(msg.Args()[2:], " ")
		} else {
			reason = strings.Join(msg.Args()[3:], " ")
		}

		//pruneDays = int(math.Max(float64(0), float64(pruneDays)))
		if pruneDays > 7 {
			pruneDays = 7
		} else if pruneDays < 0 {
			pruneDays = 0
		}
	}
	/*
		if msg.LenArgs() == 2 {
			pruneDays = 0
			reason = ""
		} else if msg.LenArgs() >= 3 {
			pruneDays, err = strconv.Atoi(msg.Args()[2])
			if err != nil {
				pruneDays = 0
				reason = strings.Join(msg.Args()[2:], " ")
			} else {
				reason = strings.Join(msg.Args()[3:], " ")
			}

			//pruneDays = int(math.Max(float64(0), float64(pruneDays)))
			if pruneDays > 7 {
				pruneDays = 7
			} else if pruneDays < 0 {
				pruneDays = 0
			}
		}
	*/
	if len(msg.Message.Mentions) > 0 {
		targetUser = msg.Message.Mentions[0]
	} else {
		sn, err := strconv.ParseUint(msg.Args()[1], 10, 64)
		if err != nil {
			return
		}
		targetUser, err = msg.Discord.Client.User(disgord.Snowflake(sn)).Get()
		if err != nil {
			return
		}
	}

	if targetUser.ID == cu.ID {
		msg.Reply("no")
		return
	}
	if targetUser.ID == msg.Message.Author.ID {
		msg.Reply("no")
		return
	}

	topUserRole := msg.HighestRole(msg.Message.GuildID, msg.Message.Author.ID)
	topTargetRole := msg.HighestRole(msg.Message.GuildID, targetUser.ID)
	topBotRole := msg.HighestRole(msg.Message.GuildID, cu.ID)

	if topUserRole <= topTargetRole || topBotRole <= topTargetRole {
		msg.Reply("no")
		return
	}

	if topTargetRole > 0 {

		userChannel, userChErr := msg.Discord.Client.User(targetUser.ID).CreateDM()

		if userChErr == nil {
			g, err := msg.Discord.Client.Guild(msg.Message.GuildID).Get()
			if err != nil {
				return
			}

			if reason == "" {
				userChannel.SendMsgString(context.Background(), msg.Discord.Client, fmt.Sprintf("You have been banned from %v", g.Name))

			} else {
				userChannel.SendMsgString(context.Background(), msg.Discord.Client, fmt.Sprintf("You have been banned from %v for the following reason:\n%v", g.Name, reason))
			}
		}
	}

	err = msg.Discord.Client.Guild(msg.Message.GuildID).Member(targetUser.ID).Ban(&disgord.BanMemberParams{
		DeleteMessageDays: pruneDays,
		Reason:            fmt.Sprintf("%v: %v", msg.Message.Author.Tag(), reason),
	})
	if err != nil {
		return
	}

	embed := &disgord.Embed{
		Title: "User banned",
		Color: 0xC80000,
		Fields: []*disgord.EmbedField{
			{
				Name:   "Username",
				Value:  fmt.Sprintf("%v", targetUser.Mention()),
				Inline: true,
			},
			{
				Name:   "ID",
				Value:  fmt.Sprintf("%v", targetUser.ID),
				Inline: true,
			},
		},
	}

	msg.Reply(embed)
}

func (m *ModerationMod) Unban(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 2 || (msg.Args()[0] != ".unban" && msg.Args()[0] != ".ub" && msg.Args()[0] != "m?unban" && msg.Args()[0] != "m?ub") {
		return
	}
	if msg.Type != meidov2.MessageTypeCreate {
		return
	}

	cu, err := msg.Discord.Client.CurrentUser().Get()
	if err != nil {
		return
	}
	botPerms, err := msg.Discord.Client.Guild(msg.Message.GuildID).GetMemberPermissions(cu.ID)
	if err != nil {
		fmt.Println(err)
		return
	}
	if botPerms&disgord.PermissionBanMembers == 0 && botPerms&disgord.PermissionAdministrator == 0 {
		return
	}

	uPerms, err := msg.Discord.Client.Guild(msg.Message.GuildID).GetMemberPermissions(msg.Message.Author.ID)
	if err != nil {
		fmt.Println(err)
		return
	}
	if uPerms&disgord.PermissionBanMembers == 0 && uPerms&disgord.PermissionAdministrator == 0 {
		return
	}

	m.cl <- msg

	userID, err := strconv.ParseUint(msg.Args()[1], 10, 64)
	if err != nil {
		return
	}

	err = msg.Discord.Client.Guild(msg.Message.GuildID).UnbanUser(disgord.Snowflake(userID), msg.Message.Author.Tag())
	if err != nil {
		return
	}

	targetUser, err := msg.Discord.Client.User(disgord.Snowflake(userID)).Get()
	if err != nil {
		return
	}

	embed := &disgord.Embed{
		Description: fmt.Sprintf("**Unbanned** %v - %v#%v (%v)", targetUser.Mention(), targetUser.Username, targetUser.Discriminator, targetUser.ID),
		Color:       0x00C800,
	}

	msg.Reply(embed)
}

func (m *ModerationMod) Hackban(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 2 || (msg.Args()[0] != "m?hackban" && msg.Args()[0] != "m?hb") {
		return
	}
	if msg.Type != meidov2.MessageTypeCreate {
		return
	}

	cu, err := msg.Discord.Client.CurrentUser().Get()
	if err != nil {
		return
	}
	botPerms, err := msg.Discord.Client.Guild(msg.Message.GuildID).GetMemberPermissions(cu.ID)
	if err != nil {
		fmt.Println(err)
		return
	}
	if botPerms&disgord.PermissionBanMembers == 0 && botPerms&disgord.PermissionAdministrator == 0 {
		return
	}

	uPerms, err := msg.Discord.Client.Guild(msg.Message.GuildID).GetMemberPermissions(msg.Message.Author.ID)
	if err != nil {
		fmt.Println(err)
		return
	}
	if uPerms&disgord.PermissionBanMembers == 0 && uPerms&disgord.PermissionAdministrator == 0 {
		return
	}

	m.cl <- msg

	var userList []string

	for _, mention := range msg.Message.Mentions {
		userList = append(userList, fmt.Sprint(mention.ID))
	}

	for _, userID := range msg.Args()[1:] {
		userList = append(userList, userID)
	}

	badBans := 0
	badIDs := 0

	for _, userIDString := range userList {
		userID, err := strconv.Atoi(userIDString)
		if err != nil {
			badIDs++
			continue
		}
		err = msg.Discord.Client.Guild(msg.Message.GuildID).Member(disgord.Snowflake(userID)).Ban(&disgord.BanMemberParams{
			DeleteMessageDays: 7,
			Reason:            fmt.Sprintf("[%v] - Hackban", msg.Message.Author.Tag()),
		})
		/*
			err = msg.Discord.Client.BanMember(context.Background(), msg.Message.GuildID, disgord.Snowflake(userID), &disgord.BanMemberParams{
				DeleteMessageDays: 7,
				Reason:            fmt.Sprintf("[%v] - Hackban", msg.Message.Author.Tag()),
			})
		*/
		if err != nil {
			fmt.Println(err)
			badBans++
			continue
		}
	}
	msg.Reply(fmt.Sprintf("Banned %v out of %v users provided.", len(userList)-badBans-badIDs, len(userList)-badIDs))
}
