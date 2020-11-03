package moderationmod

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/andersfylling/disgord"
	"github.com/dustin/go-humanize"
	"github.com/intrntsrfr/meidov2"
	"strconv"
	"strings"
	"time"
)

func (m *ModerationMod) Warn(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 3 || (msg.Args()[0] != ".warn" && msg.Args()[0] != "m?warn") {
		return
	}
	if msg.Type != meidov2.MessageTypeCreate {
		return
	}

	cu, err := msg.Discord.Client.GetCurrentUser(context.Background())
	if err != nil {
		return
	}
	botPerms, err := msg.Discord.Client.GetMemberPermissions(context.Background(), msg.Message.GuildID, cu.ID)
	if err != nil {
		return
	}
	if botPerms&disgord.PermissionBanMembers == 0 && botPerms&disgord.PermissionAdministrator == 0 {
		return
	}

	uPerms, err := msg.Discord.Client.GetMemberPermissions(context.Background(), msg.Message.GuildID, msg.Message.Author.ID)
	if err != nil {
		return
	}
	if uPerms&disgord.PermissionBanMembers == 0 && uPerms&disgord.PermissionAdministrator == 0 {
		return
	}

	m.cl <- msg

	dge := &DiscordGuild{}
	err = m.db.Get(dge, "SELECT use_strikes, max_strikes FROM guilds WHERE guild_id = $1;", msg.Message.GuildID)
	if err != nil {
		msg.Reply("there was an error, please try again")
		return
	}

	if !dge.UseStrikes {
		msg.Reply("Strike system not enabled")
		return
	}

	var (
		targetUser *disgord.Member
		reason     = "no reason"
	)

	if len(msg.Message.Mentions) >= 1 {
		targetUser, err = msg.Discord.Client.GetMember(context.Background(), msg.Message.GuildID, msg.Message.Mentions[0].ID)
		if err != nil {
			msg.Reply("that person isnt even here wtf :(")
			return
		}
	} else {
		id, err := strconv.Atoi(msg.Args()[1])
		if err != nil {
			return
		}
		targetUser, err = msg.Discord.Client.GetMember(context.Background(), msg.Message.GuildID, disgord.Snowflake(id))
		if err != nil {
			msg.Reply("that person isnt even here wtf :(")
			return
		}
	}

	if targetUser.UserID == cu.ID {
		msg.Reply("no")
		return
	}

	if targetUser.UserID == msg.Message.Author.ID {
		msg.Reply("no")
		return
	}

	topUserRole := msg.HighestRole(msg.Message.GuildID, msg.Message.Author.ID)
	topTargetRole := msg.HighestRole(msg.Message.GuildID, targetUser.UserID)
	topBotRole := msg.HighestRole(msg.Message.GuildID, cu.ID)

	if topUserRole <= topTargetRole || topBotRole <= topTargetRole {
		msg.Reply("no")
		return
	}

	if msg.LenArgs() > 3 {
		reason = strings.Join(msg.Args()[3:], " ")
	}

	warnCount := 0

	err = m.db.Get(&warnCount, "SELECT COUNT(*) FROM warns WHERE user_id=$1 AND guild_id=$2 AND is_valid",
		targetUser.UserID, msg.Message.GuildID)
	if err != nil {
		msg.Reply("something wrong happened")
		return
	}

	g, err := msg.Discord.Client.GetGuild(context.Background(), msg.Message.GuildID)
	if err != nil {
		msg.Reply("error occured")
		return
	}

	_, err = m.db.Exec("INSERT INTO warns VALUES(DEFAULT, $1, $2, $3, $4, $5, $6)",
		msg.Message.GuildID, targetUser.UserID, reason, msg.Message.Author.ID, time.Now(), true)
	if err != nil {
		fmt.Println(err)
		msg.Reply("couldnt give strike, try again?")
		return
	}

	userChannel, userChError := msg.Discord.Client.CreateDM(context.Background(), targetUser.UserID)

	// 3 / 3 strikes
	if warnCount+1 >= dge.MaxStrikes {

		if userChError == nil {
			msg.Discord.Client.SendMsg(context.Background(), userChannel.ID, fmt.Sprintf("You have been banned from %v for acquiring %v warns.\nLast warning was: %v",
				g.Name, dge.MaxStrikes, reason))
		}
		err = msg.Discord.Client.BanMember(context.Background(), g.ID, targetUser.UserID, &disgord.BanMemberParams{
			Reason:            reason,
			DeleteMessageDays: 0,
		})
		if err != nil {
			msg.Reply(err.Error())
			return
		}
		msg.Reply(fmt.Sprintf("%v has been banned after acquiring too many warns. miss them.", targetUser.Mention()))

	} else {

		if userChError == nil {
			msg.Discord.Client.SendMsg(context.Background(), userChannel.ID, fmt.Sprintf("You have been banned from %v for acquiring %v warns.\nLast warning was: %v",
				g.Name, dge.MaxStrikes, reason))
		}

		msg.Reply(fmt.Sprintf("%v has been warned\nThey are currently at warn %v/%v", targetUser.Mention(), warnCount+1, dge.MaxStrikes))

	}
}

func (m *ModerationMod) WarnLog(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 2 || msg.Args()[0] != "m?warnlog" {
		return
	}
	if msg.Type != meidov2.MessageTypeCreate {
		return
	}

	uPerms, err := msg.Discord.Client.GetMemberPermissions(context.Background(), msg.Message.GuildID, msg.Message.Author.ID)
	if err != nil {
		msg.Reply("An error occurred: " + err.Error())
		return
	}
	if uPerms&disgord.PermissionBanMembers == 0 && uPerms&disgord.PermissionAdministrator == 0 {
		return
	}

	m.cl <- msg

	page := 0

	if msg.LenArgs() > 2 {
		page, err = strconv.Atoi(msg.Args()[2])
		if err != nil {
			msg.Reply("Invalid page")
			return
		}
		if page < 1 {
			msg.Reply("Invalid page")
			return
		}
		page--
	}

	var targetUser *disgord.User

	if len(msg.Message.Mentions) >= 1 {
		targetUser = msg.Message.Mentions[0]
	} else {
		id, err := strconv.Atoi(msg.Args()[1])
		if err != nil {
			return
		}
		targetUser, err = msg.Discord.Client.GetUser(context.Background(), disgord.Snowflake(id))
		if err != nil {
			msg.Reply("error occurred: " + err.Error())
			return
		}
	}
	if targetUser == nil {
		return
	}

	var warns []*WarnEntry
	err = m.db.Select(&warns, "SELECT * FROM warns WHERE user_id=$1 AND guild_id=$2 ORDER BY given_at DESC;", targetUser.ID, msg.Message.GuildID)
	if err != nil {
		msg.Reply("there was an error, please try again")
		return
	}

	embed := &disgord.Embed{}
	embed.Title = fmt.Sprintf("Warns issued to %v", targetUser.Tag())
	embed.Footer = &disgord.EmbedFooter{
		Text: fmt.Sprintf("Page %v", page+1),
	}
	embed.Color = 0xF08152

	if len(warns) < 1 {
		embed.Description = "No warns"
	} else {

		if page*10 > len(warns) || page < 0 {
			msg.Reply("Page does not exist.")
			return
		}

		warns = warns[page*10 : min(page*10+10, len(warns))]

		for _, warn := range warns {
			fmt.Println(warn)
			field := &disgord.EmbedField{}
			field.Value = warn.Reason

			gb, err := msg.Discord.Client.GetUser(context.Background(), disgord.Snowflake(warn.GivenByID))
			if err != nil {
				msg.Reply("something terrible has happened")
				return
			}

			if warn.IsValid {
				field.Name = fmt.Sprintf("ID: %v | Issued by %v (%v) %v", warn.UID, gb.Tag(), gb.ID, humanize.Time(warn.GivenAt))
			} else {
				if warn.ClearedByID == nil {
					return
				}
				cb, err := msg.Discord.Client.GetUser(context.Background(), disgord.Snowflake(*warn.ClearedByID))
				if err != nil {
					return
				}
				field.Name = fmt.Sprintf("ID: %v | !NOT VALID! | Cleared by %v (%v) %v", warn.UID, cb.Tag(), cb.ID, humanize.Time(*warn.ClearedAt))
			}

			embed.Fields = append(embed.Fields, field)
		}
	}
	msg.Reply(embed)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (m *ModerationMod) RemoveWarn(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 2 || (msg.Args()[0] != "m?removewarn" && msg.Args()[0] != "m?rmwarn") {
		return
	}
	if msg.Type != meidov2.MessageTypeCreate {
		return
	}

	uPerms, err := msg.Discord.Client.GetMemberPermissions(context.Background(), msg.Message.GuildID, msg.Message.Author.ID)
	if err != nil {
		msg.Reply("An error occurred: " + err.Error())
		return
	}
	if uPerms&disgord.PermissionBanMembers == 0 && uPerms&disgord.PermissionAdministrator == 0 {
		return
	}

	m.cl <- msg

	uid, err := strconv.Atoi(msg.Args()[1])
	if err != nil {
		msg.Reply("no")
		return
	}

	we := &WarnEntry{}
	err = m.db.Get(we, "SELECT guild_id FROM warns WHERE uid=$1;", uid)
	if err != nil && err != sql.ErrNoRows {
		msg.Reply("there was an error, please try again")
		return
	} else if err == sql.ErrNoRows {
		msg.Reply("Warn does not exist.")
		return
	}

	if msg.Message.GuildID != disgord.Snowflake(we.GuildID) {
		msg.Reply("Nice try")
		return
	}

	_, err = m.db.Exec("UPDATE warns SET is_valid=false, cleared_by_id=$1, cleared_at=$2 WHERE uid=$3 AND is_valid", msg.Message.Author.ID, time.Now(), uid)
	if err != nil {
		msg.Reply("there was an error, please try again")
		return
	}

	msg.Reply(fmt.Sprintf("Invalidated warn with ID: %v", uid))
}

func (m *ModerationMod) ClearWarns(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 2 || (msg.Args()[0] != "m?clearwarns" && msg.Args()[0] != "m?cw") {
		return
	}
	if msg.Type != meidov2.MessageTypeCreate {
		return
	}

	uPerms, err := msg.Discord.Client.GetMemberPermissions(context.Background(), msg.Message.GuildID, msg.Message.Author.ID)
	if err != nil {
		msg.Reply("An error occurred: " + err.Error())
		return
	}
	if uPerms&disgord.PermissionBanMembers == 0 && uPerms&disgord.PermissionAdministrator == 0 {
		return
	}

	m.cl <- msg

	var targetUser *disgord.User

	if len(msg.Message.Mentions) >= 1 {
		targetUser = msg.Message.Mentions[0]
	} else {
		id, err := strconv.Atoi(msg.Args()[1])
		if err != nil {
			return
		}
		targetUser, err = msg.Discord.Client.GetUser(context.Background(), disgord.Snowflake(id))
		if err != nil {
			return
		}
	}

	_, err = m.db.Exec("UPDATE warns SET is_valid=false, cleared_by_id=$1, cleared_at=$2 WHERE user_id=$3 AND guild_id=$4 AND is_valid",
		msg.Message.Author.ID, time.Now(), targetUser.ID, msg.Message.GuildID)
	if err != nil {
		msg.Reply("there was an error, please try again")
		return
	}

	msg.Reply(fmt.Sprintf("Invalidated warns issued to %v", targetUser.Mention()))
}
