package moderationmod

import (
	"database/sql"
	"fmt"
	"github.com/andersfylling/disgord"
	"github.com/intrntsrfr/meidov2"
	"strconv"
	"strings"
	"time"
)

func (m *ModerationMod) FilterWord(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 2 || (msg.Args()[0] != "m?filterword" && msg.Args()[0] != "m?fw") {
		return
	}
	if msg.Type != meidov2.MessageTypeCreate {
		return
	}

	uPerms, err := msg.Discord.Client.Guild(msg.Message.GuildID).GetMemberPermissions(msg.Message.Author.ID)
	if err != nil {
		fmt.Println(err)
		return
	}
	if uPerms&disgord.PermissionManageMessages == 0 && uPerms&disgord.PermissionAdministrator == 0 {
		return
	}

	m.cl <- msg

	phrase := strings.Join(msg.Args()[1:], " ")

	fe := &FilterEntry{}

	err = m.db.Get(fe, "SELECT phrase FROM filters WHERE phrase = $1 AND guild_id = $2;", phrase, msg.Message.GuildID)
	switch err {
	case nil:
		m.db.Exec("DELETE FROM filters WHERE guild_id=$1 AND phrase=$2;", msg.Message.GuildID, phrase)
		msg.Reply(fmt.Sprintf("Removed `%v` from the filter.", phrase))
	case sql.ErrNoRows:
		m.db.Exec("INSERT INTO filters (guild_id, phrase) VALUES ($1,$2);", msg.Message.GuildID, phrase)
		msg.Reply(fmt.Sprintf("Added `%v` to the filter.", phrase))
	default:
		msg.Reply("there was an error, please try again")
	}
}

func (m *ModerationMod) FilterWordsList(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 1 || (msg.Args()[0] != "m?filterwordlist" && msg.Args()[0] != "m?fwl") {
		return
	}
	if msg.Type != meidov2.MessageTypeCreate {
		return
	}

	uPerms, err := msg.Discord.Client.Guild(msg.Message.GuildID).GetMemberPermissions(msg.Message.Author.ID)
	if err != nil {
		fmt.Println(err)
		return
	}
	if uPerms&disgord.PermissionManageMessages == 0 && uPerms&disgord.PermissionAdministrator == 0 {
		return
	}

	m.cl <- msg

	var fel []*FilterEntry
	err = m.db.Select(&fel, "SELECT * FROM filters WHERE guild_id=$1;", msg.Message.GuildID)
	if err != nil {
		fmt.Println(err)
		return
	}
	if len(fel) < 1 {
		msg.Reply("filter is empty")
		return
	}
	filterListBuilder := strings.Builder{}
	filterListBuilder.WriteString("```\nList of currently filtered phrases\n")
	for _, fe := range fel {
		filterListBuilder.WriteString(fmt.Sprintf("- %s\n", fe.Phrase))
	}
	filterListBuilder.WriteString("```")
	msg.Reply(filterListBuilder.String())
}

func (m *ModerationMod) ClearFilter(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 1 || msg.Args()[0] != "m?clearfilter" {
		return
	}

	uPerms, err := msg.Discord.Client.Guild(msg.Message.GuildID).GetMemberPermissions(msg.Message.Author.ID)
	if err != nil {
		fmt.Println(err)
		return
	}
	if uPerms&disgord.PermissionAdministrator == 0 {
		return
	}

	m.cl <- msg

	_, err = m.db.Exec("DELETE FROM filters WHERE guild_id=$1", msg.Message.GuildID)
	if err != nil {
		msg.Reply("there was an error")
		return
	}

	msg.Reply("Filter was cleared")
}

func (m *ModerationMod) CheckFilter(msg *meidov2.DiscordMessage) {

	isIllegal := false
	trigger := ""

	uPerms, err := msg.Discord.Client.Guild(msg.Message.GuildID).GetMemberPermissions(msg.Message.Author.ID)
	if err != nil {
		return
	}
	if uPerms&disgord.PermissionManageMessages != 0 || uPerms&disgord.PermissionAdministrator != 0 {
		return
	}

	var filterEntries []*FilterEntry
	err = m.db.Select(&filterEntries, "SELECT phrase FROM filters WHERE guild_id=$1", msg.Message.GuildID)
	if err != nil {
		return
	}

	for _, entry := range filterEntries {
		if strings.Contains(strings.ToLower(msg.Message.Content), strings.ToLower(entry.Phrase)) {
			trigger = entry.Phrase
			isIllegal = true
			break
		}
	}

	if !isIllegal {
		return
	}

	dge := &DiscordGuild{}
	err = m.db.Get(dge, "SELECT use_strikes, max_strikes FROM guilds WHERE guild_id=$1", msg.Message.GuildID)
	if err != nil {
		return
	}

	if dge.UseStrikes {

		reason := "Triggering filter: " + trigger
		warnCount := 0

		err = m.db.Get(&warnCount, "SELECT COUNT(*) FROM warns WHERE user_id=$1 AND guild_id=$2 AND is_valid",
			msg.Message.Author.ID, msg.Message.GuildID)
		if err != nil {
			return
		}

		g, err := msg.Discord.Client.Guild(msg.Message.GuildID).Get()
		if err != nil {
			return
		}
		cu, err := msg.Discord.Client.CurrentUser().Get()
		if err != nil {
			return
		}

		_, err = m.db.Exec("INSERT INTO warns VALUES(DEFAULT, $1, $2, $3, $4, $5, $6)",
			msg.Message.GuildID, msg.Message.Author.ID, reason, cu.ID, time.Now(), true)
		if err != nil {
			fmt.Println(err)
			return
		}

		userChannel, userChError := msg.Discord.Client.User(msg.Message.Author.ID).CreateDM()

		// 3 / 3 strikes
		if warnCount+1 >= dge.MaxStrikes {

			if userChError == nil {
				msg.Discord.Client.SendMsg(userChannel.ID, fmt.Sprintf("You have been banned from %v for acquiring %v warns.\nLast warning was: %v",
					g.Name, dge.MaxStrikes, reason))
			}
			err = msg.Discord.Client.Guild(g.ID).Member(msg.Message.Author.ID).Ban(&disgord.BanMemberParams{
				Reason:            reason,
				DeleteMessageDays: 0,
			})
			if err != nil {
				return
			}
			/*
				_, err = m.db.Exec("INSERT INTO warns VALUES(DEFAULT, $1, $2, $3, $4, $5, $6)",
					msg.Message.GuildID, msg.Message.Author.ID, reason, cu.ID, time.Now(), false)
			*/
			_, err = m.db.Exec("UPDATE warns SET is_valid=false, cleared_by_id=$1, cleared_at=$2 WHERE guild_id=$3 AND user_id=$4 and is_valid",
				cu.ID, time.Now(), g.ID, msg.Message.Author.ID)

			msg.Reply(fmt.Sprintf("%v has been banned after acquiring too many warns. miss them.", msg.Message.Author.Mention()))

		} else {
			if userChError == nil {
				msg.Discord.Client.SendMsg(userChannel.ID, fmt.Sprintf("You have been warned in %v.\nWarned for: %v\nYou are currently at warn %v/%v",
					g.Name, reason, warnCount+1, dge.MaxStrikes))
			}
			/*
				_, err = m.db.Exec("INSERT INTO warns VALUES(DEFAULT, $1, $2, $3, $4, $5, $6)",
					msg.Message.GuildID, msg.Message.Author.ID, reason, cu.ID, time.Now(), true)
			*/
			msg.Reply(fmt.Sprintf("%v has been warned\nThey are currently at warn %v/%v", msg.Message.Author.Mention(), warnCount+1, dge.MaxStrikes))
		}
	} else {
		msg.Reply(fmt.Sprintf("%v, you are not allowed to use a banned word/phrase", msg.Message.Author.Mention()))
	}
}

func (m *ModerationMod) ToggleStrikes(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 1 || msg.Args()[0] != "m?togglestrikes" {
		return
	}
	if msg.Type != meidov2.MessageTypeCreate {
		return
	}

	uPerms, err := msg.Discord.Client.Guild(msg.Message.GuildID).GetMemberPermissions(msg.Message.Author.ID)
	if err != nil {
		fmt.Println(err)
		return
	}
	if uPerms&disgord.PermissionAdministrator == 0 {
		return
	}

	m.cl <- msg

	dge := &DiscordGuild{}

	err = m.db.Get(dge, "SELECT * FROM guilds WHERE guild_id = $1", msg.Message.GuildID)
	if err != nil {
		return
	}
	if dge.UseStrikes {
		m.db.Exec("UPDATE guilds SET use_strikes=false WHERE guild_id=$1 AND use_strikes=true", dge.GuildID)
		msg.Reply("Strike system is now DISABLED")
	} else {
		m.db.Exec("UPDATE guilds SET use_strikes=true WHERE guild_id=$1 AND use_strikes=false", dge.GuildID)
		msg.Reply("Strike system is now ENABLED")
	}
}
func (m *ModerationMod) SetMaxStrikes(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 2 || msg.Args()[0] != "m?maxstrikes" {
		return
	}
	if msg.Type != meidov2.MessageTypeCreate {
		return
	}

	uPerms, err := msg.Discord.Client.Guild(msg.Message.GuildID).GetMemberPermissions(msg.Message.Author.ID)
	if err != nil {
		fmt.Println(err)
		return
	}
	if uPerms&disgord.PermissionAdministrator == 0 {
		return
	}

	m.cl <- msg

	n, err := strconv.Atoi(msg.Args()[1])
	if err != nil {
		return
	}
	if n < 0 {
		n = 0
	} else if n > 10 {
		n = 10
	}

	m.db.Exec("UPDATE guilds SET max_strikes=$1 WHERE guild_id=$2", n, msg.Message.GuildID)
}
