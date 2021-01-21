package moderationmod

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meidov2"
	"strconv"
	"strings"
	"time"
)

func NewFilterWordCommand(m *ModerationMod) *meidov2.ModCommand {
	return &meidov2.ModCommand{
		Mod:           m,
		Name:          "filterword",
		Description:   "Adds or removes a word or phrase to the server filter.",
		Triggers:      []string{"m?fw", "m?filterword"},
		Usage:         "m?fw jeff",
		Cooldown:      2,
		RequiredPerms: discordgo.PermissionManageMessages,
		RequiresOwner: false,
		AllowedTypes:  meidov2.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           m.filterwordCommand,
	}
}
func (m *ModerationMod) filterwordCommand(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 2 {
		return
	}

	phrase := strings.Join(msg.Args()[1:], " ")

	fe := &FilterEntry{}

	err := m.db.Get(fe, "SELECT phrase FROM filters WHERE phrase = $1 AND guild_id = $2;", phrase, msg.Message.GuildID)
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

func NewFilterWordListCommand(m *ModerationMod) *meidov2.ModCommand {
	return &meidov2.ModCommand{
		Mod:           m,
		Name:          "filterwordlist",
		Description:   "Lists of all filtered phrases for this server",
		Triggers:      []string{"m?fwl", "m?filterwordlist"},
		Usage:         "m?fwl",
		Cooldown:      10,
		RequiredPerms: discordgo.PermissionManageMessages,
		RequiresOwner: false,
		AllowedTypes:  meidov2.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           m.filterwordlistCommand,
	}
}
func (m *ModerationMod) filterwordlistCommand(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 1 {
		return
	}

	var fel []*FilterEntry
	err := m.db.Select(&fel, "SELECT * FROM filters WHERE guild_id=$1;", msg.Message.GuildID)
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

	if len(filterListBuilder.String()) > 1000 {
		buf := &bytes.Buffer{}
		buf.WriteString(filterListBuilder.String())

		msg.Sess.ChannelFileSend(msg.Message.ChannelID, "filter.txt", buf)
	} else {
		msg.Reply(filterListBuilder.String())
	}
}

func NewClearFilterCommand(m *ModerationMod) *meidov2.ModCommand {
	return &meidov2.ModCommand{
		Mod:           m,
		Name:          "clearfilter",
		Description:   "Removes all phrases from the server filter",
		Triggers:      []string{"m?clearfilter"},
		Usage:         "m?clearfilter",
		Cooldown:      10,
		RequiredPerms: discordgo.PermissionAdministrator,
		RequiresOwner: false,
		AllowedTypes:  meidov2.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           m.clearfilterCommand,
	}
}

func (m *ModerationMod) clearfilterCommand(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 1 {
		return
	}

	_, err := m.db.Exec("DELETE FROM filters WHERE guild_id=$1", msg.Message.GuildID)
	if err != nil {
		msg.Reply("there was an error")
		return
	}

	msg.Reply("Filter was cleared")
}

func NewModerationSettingsCommand(m *ModerationMod) *meidov2.ModCommand {
	return &meidov2.ModCommand{
		Mod:           m,
		Name:          "moderationsettings",
		Description:   "Moderation settings:\n- Toggle warn system [enable/disable]\n- Set max warns [0-10]\n- Set warn duration [0(infinite)-365]",
		Triggers:      []string{"m?settings moderation"},
		Usage:         "m?settings moderation warns enable/disable\nm?settings moderation maxwarns [0-10]\nm?settings moderation warnduration [0-365]",
		Cooldown:      2,
		RequiredPerms: discordgo.PermissionAdministrator,
		RequiresOwner: false,
		AllowedTypes:  meidov2.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           m.moderationsettingsCommand,
	}
}
func (m *ModerationMod) moderationsettingsCommand(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 2 {
		return
	}

	switch msg.LenArgs() {
	case 2:
		dge := &DiscordGuild{}
		err := m.db.Get(dge, "SELECT * FROM guilds WHERE guild_id=$1", msg.Message.GuildID)
		if err != nil {
			return
		}

		emb := &discordgo.MessageEmbed{
			Title: "Moderation settings",
			Color: 0xFEFEFE,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Warns enabled",
					Value:  fmt.Sprint(dge.UseWarns),
					Inline: true,
				},
				{
					Name:   "Max warns",
					Value:  fmt.Sprint(dge.MaxWarns),
					Inline: true,
				},
				{
					Name:   "Warn duration",
					Value:  fmt.Sprintf("%v days", dge.WarnDuration),
					Inline: true,
				},
			},
		}
		msg.ReplyEmbed(emb)

	case 4:
		switch msg.Args()[2] {
		case "warns":
			if msg.Args()[3] == "enable" {
				m.db.Exec("UPDATE guilds SET use_warns=true WHERE guild_id=$1 AND NOT use_warns", msg.Message.GuildID)
				msg.Reply("Strike system is now ENABLED")

			} else if msg.Args()[3] == "disable" {
				m.db.Exec("UPDATE guilds SET use_warns=false WHERE guild_id=$1 AND use_warns", msg.Message.GuildID)
				msg.Reply("Strike system is now DISABLED")
			}
		case "maxwarns":

			n, err := strconv.Atoi(msg.Args()[3])
			if err != nil {
				return
			}

			n = meidov2.Clamp(0, 10, n)

			_, err = m.db.Exec("UPDATE guilds SET max_warns=$1 WHERE guild_id=$2", n, msg.Message.GuildID)
			if err != nil {
				msg.Reply("error setting max warns")
				return
			}
			msg.Reply(fmt.Sprintf("set max warns to %v", n))
		case "warnduration":

			n, err := strconv.Atoi(msg.Args()[3])
			if err != nil {
				return
			}

			n = meidov2.Clamp(0, 365, n)

			_, err = m.db.Exec("UPDATE guilds SET warn_duration=$1 WHERE guild_id=$2", n, msg.Message.GuildID)
			if err != nil {
				msg.Reply("error setting warn duration")
				return
			}
			msg.Reply(fmt.Sprintf("set warn duration to %v days", n))
		}
	}
}

func NewCheckFilterPassive(m *ModerationMod) *meidov2.ModPassive {
	return &meidov2.ModPassive{
		Mod:          m,
		Name:         "checkfilter",
		Description:  "checks if messages contain phrases found in the server filter",
		Enabled:      true,
		AllowedTypes: meidov2.MessageTypeCreate | meidov2.MessageTypeUpdate,
		Run:          m.checkfilterPassive,
	}
}

func (m *ModerationMod) checkfilterPassive(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 1 {
		return
	}

	isIllegal := false
	trigger := ""

	uPerms, err := msg.Discord.UserChannelPermissionsDirect(msg.Member, msg.Message.ChannelID)
	if err != nil {
		return
	}
	if uPerms&discordgo.PermissionManageMessages != 0 || uPerms&discordgo.PermissionAdministrator != 0 {
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

	msg.Sess.ChannelMessageDelete(msg.Message.ChannelID, msg.Message.ID)

	dge := &DiscordGuild{}
	err = m.db.Get(dge, "SELECT use_warns, max_warns FROM guilds WHERE guild_id=$1", msg.Message.GuildID)
	if err != nil {
		return
	}
	if !dge.UseWarns {
		msg.Reply(fmt.Sprintf("%v, you are not allowed to use a banned word/phrase", msg.Message.Author.Mention()))
		return
	}

	reason := "Triggering filter: " + trigger
	warnCount := 0

	err = m.db.Get(&warnCount, "SELECT COUNT(*) FROM warns WHERE user_id=$1 AND guild_id=$2 AND is_valid",
		msg.Message.Author.ID, msg.Message.GuildID)
	if err != nil {
		return
	}

	g, err := msg.Discord.Guild(msg.Message.GuildID)
	if err != nil {
		return
	}
	cu := msg.Discord.Sess.State.User

	_, err = m.db.Exec("INSERT INTO warns VALUES(DEFAULT, $1, $2, $3, $4, $5, $6)",
		msg.Message.GuildID, msg.Message.Author.ID, reason, cu.ID, time.Now(), true)
	if err != nil {
		fmt.Println(err)
		return
	}

	userChannel, userChError := msg.Discord.Sess.UserChannelCreate(msg.Message.Author.ID)

	// 3 / 3 strikes
	if warnCount+1 >= dge.MaxWarns {

		if userChError == nil {
			msg.Discord.Sess.ChannelMessageSend(userChannel.ID, fmt.Sprintf("You have been banned from %v for acquiring %v warns.\nLast warning was: %v",
				g.Name, dge.MaxWarns, reason))
		}
		err = msg.Discord.Sess.GuildBanCreateWithReason(g.ID, msg.Message.Author.ID, fmt.Sprintf("Acquired %v strikes.", dge.MaxWarns), 0)
		if err != nil {
			return
		}

		_, _ = m.db.Exec("UPDATE warns SET is_valid=false, cleared_by_id=$1, cleared_at=$2 WHERE guild_id=$3 AND user_id=$4 and is_valid",
			cu.ID, time.Now(), g.ID, msg.Message.Author.ID)

		msg.Reply(fmt.Sprintf("%v has been banned after acquiring too many warns. miss them.", msg.Message.Author.Mention()))

	} else {
		if userChError == nil {
			msg.Discord.Sess.ChannelMessageSend(userChannel.ID, fmt.Sprintf("You have been warned in %v.\nWarned for: %v\nYou are currently at warn %v/%v",
				g.Name, reason, warnCount+1, dge.MaxWarns))
		}
		msg.Reply(fmt.Sprintf("%v has been warned\nThey are currently at warn %v/%v", msg.Message.Author.Mention(), warnCount+1, dge.MaxWarns))
	}
}
