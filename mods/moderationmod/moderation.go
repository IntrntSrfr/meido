package moderationmod

import (
	"database/sql"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meidov2"
	"github.com/jmoiron/sqlx"
	"strconv"
	"strings"
	"time"
)

type ModerationMod struct {
	cl       chan *meidov2.DiscordMessage
	commands []func(msg *meidov2.DiscordMessage)
	db       *sqlx.DB
}

func New() meidov2.Mod {
	return &ModerationMod{}
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
func (m *ModerationMod) Commands() map[string]meidov2.ModCommand {
	return nil
}

func (m *ModerationMod) Hook(b *meidov2.Bot) error {
	m.cl = b.CommandLog
	m.db = b.DB

	b.Discord.Sess.AddHandler(func(s *discordgo.Session, g *discordgo.GuildCreate) {
		dbg := &DiscordGuild{}
		err := m.db.Get(dbg, "SELECT guild_id FROM guilds WHERE guild_id = $1;", g.Guild.ID)
		if err != nil && err != sql.ErrNoRows {
			fmt.Println(err)
		} else if err == sql.ErrNoRows {
			m.db.Exec("INSERT INTO guilds(guild_id, use_strikes, max_strikes) VALUES($1, $2, $3)", g.Guild.ID, false, 3)
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
	if msg.IsDM() {
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

	botPerms, err := msg.Discord.Sess.State.UserChannelPermissions(msg.Sess.State.User.ID, msg.Message.ChannelID)
	if err != nil {
		return
	}
	if botPerms&discordgo.PermissionBanMembers == 0 && botPerms&discordgo.PermissionAdministrator == 0 {
		return
	}

	uPerms, err := msg.Discord.Sess.State.UserChannelPermissions(msg.Message.Author.ID, msg.Message.ChannelID)
	if err != nil {
		return
	}
	if uPerms&discordgo.PermissionBanMembers == 0 && uPerms&discordgo.PermissionAdministrator == 0 {
		return
	}

	m.cl <- msg

	var (
		targetUser *discordgo.User
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

	if len(msg.Message.Mentions) > 0 {
		targetUser = msg.Message.Mentions[0]
	} else {
		targetUser, err = msg.Discord.Sess.User(msg.Args()[1])
		if err != nil {
			return
		}
	}

	if targetUser.ID == msg.Sess.State.User.ID {
		msg.Reply("no")
		return
	}
	if targetUser.ID == msg.Message.Author.ID {
		msg.Reply("no")
		return
	}

	topUserRole := msg.HighestRole(msg.Message.GuildID, msg.Message.Author.ID)
	topTargetRole := msg.HighestRole(msg.Message.GuildID, targetUser.ID)
	topBotRole := msg.HighestRole(msg.Message.GuildID, msg.Sess.State.User.ID)

	if topUserRole <= topTargetRole || topBotRole <= topTargetRole {
		msg.Reply("no")
		return
	}

	if topTargetRole > 0 {
		userChannel, userChErr := msg.Discord.Sess.UserChannelCreate(targetUser.ID)
		if userChErr == nil {
			g, err := msg.Discord.Sess.State.Guild(msg.Message.GuildID)
			if err != nil {
				return
			}

			if reason == "" {
				msg.Sess.ChannelMessageSend(userChannel.ID, fmt.Sprintf("You have been banned from %v", g.Name))
			} else {
				msg.Sess.ChannelMessageSend(userChannel.ID, fmt.Sprintf("You have been banned from %v for the following reason:\n%v", g.Name, reason))
			}
		}
	}

	err = msg.Discord.Sess.GuildBanCreateWithReason(msg.Message.GuildID, targetUser.ID, fmt.Sprintf("%v - %v", msg.Message.Author.String(), reason), pruneDays)
	if err != nil {
		msg.Reply(err.Error())
		return
	}

	_, err = m.db.Exec("UPDATE warns SET is_valid=false, cleared_by_id=$1, cleared_at=$2 WHERE guild_id=$3 AND user_id=$4 and is_valid",
		msg.Sess.State.User.ID, time.Now(), msg.Message.GuildID, targetUser.ID)

	embed := &discordgo.MessageEmbed{
		Title: "User banned",
		Color: 0xC80000,
		Fields: []*discordgo.MessageEmbedField{
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

	msg.ReplyEmbed(embed)
}

func (m *ModerationMod) Unban(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 2 || (msg.Args()[0] != ".unban" && msg.Args()[0] != ".ub" && msg.Args()[0] != "m?unban" && msg.Args()[0] != "m?ub") {
		return
	}
	if msg.Type != meidov2.MessageTypeCreate {
		return
	}

	botPerms, err := msg.Discord.Sess.State.UserChannelPermissions(msg.Sess.State.User.ID, msg.Message.ChannelID)
	if err != nil {
		fmt.Println(err)
		return
	}
	if botPerms&discordgo.PermissionBanMembers == 0 && botPerms&discordgo.PermissionAdministrator == 0 {
		return
	}

	uPerms, err := msg.Discord.Sess.State.UserChannelPermissions(msg.Message.Author.ID, msg.Message.ChannelID)
	if err != nil {
		fmt.Println(err)
		return
	}
	if uPerms&discordgo.PermissionBanMembers == 0 && uPerms&discordgo.PermissionAdministrator == 0 {
		return
	}

	m.cl <- msg

	_, err = strconv.ParseUint(msg.Args()[1], 10, 64)
	if err != nil {
		return
	}

	err = msg.Discord.Sess.GuildBanDelete(msg.Message.GuildID, msg.Args()[1])
	if err != nil {
		return
	}

	targetUser, err := msg.Discord.Sess.User(msg.Args()[1])
	if err != nil {
		return
	}

	embed := &discordgo.MessageEmbed{
		Description: fmt.Sprintf("**Unbanned** %v - %v#%v (%v)", targetUser.Mention(), targetUser.Username, targetUser.Discriminator, targetUser.ID),
		Color:       0x00C800,
	}

	msg.ReplyEmbed(embed)
}

func (m *ModerationMod) Hackban(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 2 || (msg.Args()[0] != "m?hackban" && msg.Args()[0] != "m?hb") {
		return
	}
	if msg.Type != meidov2.MessageTypeCreate {
		return
	}

	botPerms, err := msg.Discord.Sess.State.UserChannelPermissions(msg.Sess.State.User.ID, msg.Message.ChannelID)
	if err != nil {
		fmt.Println(err)
		return
	}
	if botPerms&discordgo.PermissionBanMembers == 0 && botPerms&discordgo.PermissionAdministrator == 0 {
		return
	}

	uPerms, err := msg.Discord.Sess.State.UserChannelPermissions(msg.Message.Author.ID, msg.Message.ChannelID)
	if err != nil {
		fmt.Println(err)
		return
	}
	if uPerms&discordgo.PermissionBanMembers == 0 && uPerms&discordgo.PermissionAdministrator == 0 {
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
		_, err := strconv.Atoi(userIDString)
		if err != nil {
			badIDs++
			continue
		}
		err = msg.Discord.Sess.GuildBanCreateWithReason(msg.Message.GuildID, userIDString, fmt.Sprintf("[%v] - Hackban", msg.Message.Author.String()), 7)
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
