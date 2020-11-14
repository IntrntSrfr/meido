package moderationmod

import (
	"database/sql"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/intrntsrfr/meidov2"
	"strconv"
	"strings"
	"time"
)

func (m *ModerationMod) Warn(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 2 || (msg.Args()[0] != ".warn" && msg.Args()[0] != "m?warn") {
		return
	}
	if msg.Type != meidov2.MessageTypeCreate {
		return
	}

	uPerms, err := msg.Discord.UserChannelPermissions(msg.Author, msg.Message.ChannelID)
	if err != nil {
		fmt.Println(err)
		return
	}
	if uPerms&discordgo.PermissionBanMembers == 0 && uPerms&discordgo.PermissionAdministrator == 0 {
		return
	}

	botPerms, err := msg.Discord.Sess.State.UserChannelPermissions(msg.Discord.Sess.State.User.ID, msg.Message.ChannelID)
	if err != nil {
		return
	}
	if botPerms&discordgo.PermissionBanMembers == 0 && botPerms&discordgo.PermissionAdministrator == 0 {
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
		targetUser *discordgo.Member
		reason     = "no reason"
	)

	if len(msg.Message.Mentions) >= 1 {
		targetUser, err = msg.Discord.Sess.GuildMember(msg.Message.GuildID, msg.Message.Mentions[0].ID)
		if err != nil {
			msg.Reply("that person isnt even here wtf :(")
			return
		}
	} else {
		_, err := strconv.Atoi(msg.Args()[1])
		if err != nil {
			return
		}
		targetUser, err = msg.Discord.Sess.GuildMember(msg.Message.GuildID, msg.Args()[1])
		if err != nil {
			msg.Reply("that person isnt even here wtf :(")
			return
		}
	}

	if targetUser.User.ID == msg.Sess.State.User.ID || targetUser.User.Bot || targetUser.User.ID == msg.Message.Author.ID {
		msg.Reply("no")
		return
	}

	topUserRole := msg.HighestRole(msg.Message.GuildID, msg.Message.Author.ID)
	topTargetRole := msg.HighestRole(msg.Message.GuildID, targetUser.User.ID)
	topBotRole := msg.HighestRole(msg.Message.GuildID, msg.Sess.State.User.ID)

	if topUserRole <= topTargetRole || topBotRole <= topTargetRole {
		msg.Reply("no")
		return
	}

	if msg.LenArgs() > 3 {
		reason = strings.Join(msg.Args()[3:], " ")
	}

	warnCount := 0

	err = m.db.Get(&warnCount, "SELECT COUNT(*) FROM warns WHERE user_id=$1 AND guild_id=$2 AND is_valid",
		targetUser.User.ID, msg.Message.GuildID)
	if err != nil {
		msg.Reply("something wrong happened")
		return
	}

	g, err := msg.Discord.Sess.State.Guild(msg.Message.GuildID)
	if err != nil {
		msg.Reply("error occurred")
		return
	}

	_, err = m.db.Exec("INSERT INTO warns VALUES(DEFAULT, $1, $2, $3, $4, $5, $6)",
		msg.Message.GuildID, targetUser.User.ID, reason, msg.Message.Author.ID, time.Now(), true)
	if err != nil {
		msg.Reply("error giving strike, try again?")
		return
	}

	userChannel, userChError := msg.Discord.Sess.UserChannelCreate(targetUser.User.ID)

	// 3 / 3 strikes
	if warnCount+1 >= dge.MaxStrikes {

		if userChError == nil {
			msg.Discord.Sess.ChannelMessageSend(userChannel.ID, fmt.Sprintf("You have been banned from %v for acquiring %v warns.\nLast warning was: %v",
				g.Name, dge.MaxStrikes, reason))
		}
		err = msg.Discord.Sess.GuildBanCreateWithReason(msg.Message.GuildID, targetUser.User.ID, reason, 0)
		if err != nil {
			msg.Reply(err.Error())
			return
		}
		_, err = m.db.Exec("UPDATE warns SET is_valid=false, cleared_by_id=$1, cleared_at=$2 WHERE guild_id=$3 AND user_id=$4 and is_valid",
			msg.Sess.State.User.ID, time.Now(), g.ID, msg.Message.Author.ID)

		msg.Reply(fmt.Sprintf("%v has been banned after acquiring too many warns. miss them.", targetUser.Mention()))

	} else {
		if userChError == nil {
			msg.Discord.Sess.ChannelMessageSend(userChannel.ID, fmt.Sprintf("You have been warned in %v.\nWarned for: %v\nYou are currently at warn %v/%v",
				g.Name, reason, warnCount+1, dge.MaxStrikes))
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

	uPerms, err := msg.Discord.UserChannelPermissions(msg.Author, msg.Message.ChannelID)
	if err != nil {
		fmt.Println(err)
		return
	}
	if uPerms&discordgo.PermissionBanMembers == 0 && uPerms&discordgo.PermissionAdministrator == 0 {
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

	var targetUser *discordgo.User

	if len(msg.Message.Mentions) >= 1 {
		targetUser = msg.Message.Mentions[0]
	} else {
		_, err := strconv.Atoi(msg.Args()[1])
		if err != nil {
			return
		}
		targetUser, err = msg.Discord.Sess.User(msg.Args()[1])
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

	embed := &discordgo.MessageEmbed{}
	embed.Title = fmt.Sprintf("Warns issued to %v", targetUser.String())
	embed.Footer = &discordgo.MessageEmbedFooter{
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
			field := &discordgo.MessageEmbedField{}
			field.Value = warn.Reason

			gb, err := msg.Discord.Sess.User(warn.GivenByID)
			if err != nil {
				msg.Reply("something terrible has happened")
				return
			}

			if warn.IsValid {
				field.Name = fmt.Sprintf("ID: %v | Issued by %v (%v) %v", warn.UID, gb.String(), gb.ID, humanize.Time(warn.GivenAt))
			} else {
				if warn.ClearedByID == nil {
					return
				}
				cb, err := msg.Discord.Sess.User(*warn.ClearedByID)
				if err != nil {
					return
				}
				field.Name = fmt.Sprintf("ID: %v | !NOT VALID! | Cleared by %v (%v) %v", warn.UID, cb.String(), cb.ID, humanize.Time(*warn.ClearedAt))
			}

			embed.Fields = append(embed.Fields, field)
		}
	}
	msg.ReplyEmbed(embed)
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

	uPerms, err := msg.Discord.UserChannelPermissions(msg.Author, msg.Message.ChannelID)
	if err != nil {
		fmt.Println(err)
		return
	}
	if uPerms&discordgo.PermissionBanMembers == 0 && uPerms&discordgo.PermissionAdministrator == 0 {
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
		msg.Reply("Warn does not exist")
		return
	}

	if msg.Message.GuildID != we.GuildID {
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

	uPerms, err := msg.Discord.UserChannelPermissions(msg.Author, msg.Message.ChannelID)
	if err != nil {
		fmt.Println(err)
		return
	}
	if uPerms&discordgo.PermissionBanMembers == 0 && uPerms&discordgo.PermissionAdministrator == 0 {
		return
	}

	m.cl <- msg

	var targetUser *discordgo.User

	if len(msg.Message.Mentions) >= 1 {
		targetUser = msg.Message.Mentions[0]
	} else {
		_, err := strconv.Atoi(msg.Args()[1])
		if err != nil {
			return
		}
		targetUser, err = msg.Discord.Sess.User(msg.Args()[1])
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
