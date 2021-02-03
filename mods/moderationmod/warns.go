package moderationmod

import (
	"database/sql"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/intrntsrfr/meido"
	"strconv"
	"strings"
	"time"
)

func NewWarnCommand(m *ModerationMod) *meido.ModCommand {
	return &meido.ModCommand{
		Mod:           m,
		Name:          "warn",
		Description:   "Warns a user. Does not work if warn system is disabled.",
		Triggers:      []string{"m?warn", ".warn"},
		Usage:         "m?warn 163454407999094786 stupid",
		Cooldown:      2,
		RequiredPerms: discordgo.PermissionBanMembers,
		RequiresOwner: false,
		CheckBotPerms: true,
		AllowedTypes:  meido.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           m.warnCommand,
	}
}

func (m *ModerationMod) warnCommand(msg *meido.DiscordMessage) {
	if msg.LenArgs() < 2 {
		return
	}

	dge := &DiscordGuild{}
	err := m.db.Get(dge, "SELECT use_warns, max_warns FROM guilds WHERE guild_id = $1;", msg.Message.GuildID)
	if err != nil {
		msg.Reply("there was an error, please try again")
		return
	}

	if !dge.UseWarns {
		msg.Reply("Strike system not enabled")
		return
	}

	var (
		targetUser *discordgo.Member
		reason     = "no reason"
	)

	if len(msg.Message.Mentions) >= 1 {
		targetUser, err = msg.Discord.Member(msg.Message.GuildID, msg.Message.Mentions[0].ID)
		if err != nil {
			msg.Reply("that person isnt even here wtf :(")
			return
		}
	} else {
		_, err := strconv.Atoi(msg.Args()[1])
		if err != nil {
			return
		}
		targetUser, err = msg.Discord.Member(msg.Message.GuildID, msg.Args()[1])
		if err != nil {
			msg.Reply("that person isnt even here wtf :(")
			return
		}
	}

	if targetUser.User.ID == msg.Sess.State.User.ID || targetUser.User.Bot || targetUser.User.ID == msg.Message.Author.ID {
		msg.Reply("no")
		return
	}

	topUserRole := msg.Discord.HighestRolePosition(msg.Message.GuildID, msg.Message.Author.ID)
	topTargetRole := msg.Discord.HighestRolePosition(msg.Message.GuildID, targetUser.User.ID)
	topBotRole := msg.Discord.HighestRolePosition(msg.Message.GuildID, msg.Sess.State.User.ID)

	if topUserRole <= topTargetRole || topBotRole <= topTargetRole {
		msg.Reply("no")
		return
	}

	if msg.LenArgs() > 2 {
		reason = strings.Join(msg.Args()[2:], " ")
	}

	warnCount := 0

	err = m.db.Get(&warnCount, "SELECT COUNT(*) FROM warns WHERE user_id=$1 AND guild_id=$2 AND is_valid",
		targetUser.User.ID, msg.Message.GuildID)
	if err != nil {
		msg.Reply("something wrong happened")
		return
	}

	g, err := msg.Discord.Guild(msg.Message.GuildID)
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
	if warnCount+1 >= dge.MaxWarns {

		if userChError == nil {
			msg.Discord.Sess.ChannelMessageSend(userChannel.ID, fmt.Sprintf("You have been banned from %v for acquiring %v warns.\nLast warning was: %v",
				g.Name, dge.MaxWarns, reason))
		}
		err = msg.Discord.Sess.GuildBanCreateWithReason(msg.Message.GuildID, targetUser.User.ID, fmt.Sprintf("Acquired %v strikes.", dge.MaxWarns), 0)
		if err != nil {
			msg.Reply(err.Error())
			return
		}
		_, err = m.db.Exec("UPDATE warns SET is_valid=false, cleared_by_id=$1, cleared_at=$2 WHERE guild_id=$3 AND user_id=$4 and is_valid",
			msg.Sess.State.User.ID, time.Now(), g.ID, targetUser.User.ID)

		msg.Reply(fmt.Sprintf("%v has been banned after acquiring too many warns. miss them.", targetUser.Mention()))

	} else {
		if userChError == nil {
			msg.Discord.Sess.ChannelMessageSend(userChannel.ID, fmt.Sprintf("You have been warned in %v.\nWarned for: %v\nYou are currently at warn %v/%v",
				g.Name, reason, warnCount+1, dge.MaxWarns))
		}
		msg.Reply(fmt.Sprintf("%v has been warned\nThey are currently at warn %v/%v", targetUser.Mention(), warnCount+1, dge.MaxWarns))
	}
}

func NewWarnLogCommand(m *ModerationMod) *meido.ModCommand {
	return &meido.ModCommand{
		Mod:           m,
		Name:          "warnlog",
		Description:   "Displays a users warns",
		Triggers:      []string{"m?warnlog"},
		Usage:         "m?warnlog 163454407999094786",
		Cooldown:      5,
		RequiredPerms: discordgo.PermissionManageMessages,
		RequiresOwner: false,
		AllowedTypes:  meido.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           m.warnlogCommand,
	}
}

func (m *ModerationMod) warnlogCommand(msg *meido.DiscordMessage) {
	if msg.LenArgs() < 2 {
		return
	}

	page := 0

	if msg.LenArgs() > 2 {
		page, err := strconv.Atoi(msg.Args()[2])
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
	err := m.db.Select(&warns, "SELECT * FROM warns WHERE user_id=$1 AND guild_id=$2 ORDER BY given_at DESC;", targetUser.ID, msg.Message.GuildID)
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

		if page*10 > len(warns) {
			msg.Reply("Page does not exist.")
			return
		}

		warns = warns[page*10 : meido.Min(page*10+10, len(warns))]

		userCache := make(map[string]*discordgo.User)

		for _, warn := range warns {
			field := &discordgo.MessageEmbedField{}
			field.Value = warn.Reason

			var gb *discordgo.User
			gb, ok := userCache[warn.GivenByID]
			if !ok {
				gb, err = msg.Discord.Sess.User(warn.GivenByID)
				if err != nil {
					msg.Reply("something terrible has happened")
					return
				}
				userCache[warn.GivenByID] = gb
			}

			if warn.IsValid {
				field.Name = fmt.Sprintf("ID: %v | Issued by %v (%v) %v", warn.UID, gb.String(), gb.ID, humanize.Time(warn.GivenAt))
			} else {
				if warn.ClearedByID == nil {
					return
				}

				var cb *discordgo.User
				cb, ok := userCache[*warn.ClearedByID]
				if !ok {
					cb, err = msg.Discord.Sess.User(*warn.ClearedByID)
					if err != nil {
						msg.Reply("something terrible has happened")
						return
					}
					userCache[*warn.ClearedByID] = cb
				}

				field.Name = fmt.Sprintf("ID: %v | !CLEARED! | Cleared by %v (%v) %v", warn.UID, cb.String(), cb.ID, humanize.Time(*warn.ClearedAt))
			}

			embed.Fields = append(embed.Fields, field)
		}
	}
	msg.ReplyEmbed(embed)
}

func NewWarnCountCommand(m *ModerationMod) *meido.ModCommand {
	return &meido.ModCommand{
		Mod:           m,
		Name:          "warncount",
		Description:   "Displays how many warns a user has",
		Triggers:      []string{"m?warncount"},
		Usage:         "m?warncount | m?warncount @user",
		Cooldown:      2,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  meido.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           m.warncountCommand,
	}
}

func (m *ModerationMod) warncountCommand(msg *meido.DiscordMessage) {

	var (
		err        error
		targetUser *discordgo.User
	)

	dge := DiscordGuild{}
	err = m.db.Get(&dge, "SELECT use_warns, max_warns FROM guilds WHERE guild_id=$1", msg.Message.GuildID)
	if err != nil {
		return
	}

	if !dge.UseWarns {
		msg.Reply("warn system not enabled")
		return
	}

	if msg.LenArgs() > 1 {
		if len(msg.Message.Mentions) >= 1 {
			targetUser = msg.Message.Mentions[0]
		} else {
			_, err := strconv.Atoi(msg.Args()[1])
			if err != nil {
				return
			}
			targetUser, err = msg.Sess.User(msg.Args()[1])
			if err != nil {
				return
			}
		}
	} else {
		targetUser = msg.Message.Author
	}

	warnCount := 0

	err = m.db.Get(&warnCount, "SELECT COUNT(*) FROM warns WHERE guild_id=$1 AND user_id=$2 AND is_valid", msg.Message.GuildID, targetUser.ID)
	if err != nil {
		return
	}

	msg.Reply(fmt.Sprintf("%v is at %v/%v warns", targetUser.String(), warnCount, dge.MaxWarns))
}

func NewClearWarnCommand(m *ModerationMod) *meido.ModCommand {
	return &meido.ModCommand{
		Mod:           m,
		Name:          "clearwarn",
		Description:   "Clears a warn from a user using a warnID. Use warnlog to get warnIDs",
		Triggers:      []string{"m?clearwarn"},
		Usage:         "m?clearwarn 123",
		Cooldown:      3,
		RequiredPerms: discordgo.PermissionBanMembers,
		RequiresOwner: false,
		AllowedTypes:  meido.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           m.clearwarnCommand,
	}
}

func (m *ModerationMod) clearwarnCommand(msg *meido.DiscordMessage) {
	if msg.LenArgs() < 2 {
		return
	}
	/*
		id := meido.UserIDRegex.FindStringSubmatch(msg.Args()[1])[0]
		if len(id) < 1 {
			return
		}
	*/

	_, err := strconv.Atoi(msg.Args()[1])
	if err != nil {
		msg.Reply("no")
		return
	}

	var entries []*WarnEntry
	err = m.db.Select(&entries, "SELECT * FROM warns WHERE user_id=$1 AND guild_id=$2 AND is_valid", msg.Args()[1], msg.Message.GuildID)
	if err != nil && err != sql.ErrNoRows {
		fmt.Println(err)
		msg.Reply("there was an error, please try again")
		return
	} else if err == sql.ErrNoRows {
		msg.Reply("User has no warns.")
		return
	}

	if len(entries) == 0 {
		msg.Reply("no warns")
		return
	}

	sb := strings.Builder{}
	sb.WriteString("Which warn would you like to remove? [1-10]\n")
	for i, entry := range entries {
		sb.WriteString(fmt.Sprintf("`%2v.` | '%v' given %v\n", i+1, entry.Reason, humanize.Time(entry.GivenAt)))
	}
	menu, err := msg.Reply(sb.String())
	if err != nil {
		return
	}

	cb, err := m.bot.MakeCallback(msg.Message.ChannelID, msg.Author.ID)
	if err != nil {
		return
	}

	// this needs a timeout
	var n int
	for {
		reply := <-cb

		n, err = strconv.Atoi(reply.RawContent())
		if err == nil && n-1 >= 0 && n-1 < len(entries) {
			break
		}
	}
	msg.Sess.ChannelMessageDelete(menu.ChannelID, menu.ID)

	m.bot.CloseCallback(msg.Message.ChannelID, msg.Author.ID)

	selectedEntry := entries[n-1]

	_, err = m.db.Exec("UPDATE warns SET is_valid=false, cleared_by_id=$1, cleared_at=$2 WHERE uid=$3 AND is_valid", msg.Message.Author.ID, time.Now(), selectedEntry.UID)
	if err != nil {
		msg.Reply("there was an error, please try again")
		return
	}

	msg.Reply(fmt.Sprintf("Invalidated warn with ID: %v", selectedEntry.UID))
}

func NewClearAllWarnsCommand(m *ModerationMod) *meido.ModCommand {
	return &meido.ModCommand{
		Mod:           m,
		Name:          "clearallwarns",
		Description:   "Clears all active warns for a member",
		Triggers:      []string{"m?clearallwarns"},
		Usage:         "m?clearallwarns 163454407999094786",
		Cooldown:      5,
		RequiredPerms: discordgo.PermissionBanMembers,
		RequiresOwner: false,
		AllowedTypes:  meido.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           m.clearallwarnsCommand,
	}
}
func (m *ModerationMod) clearallwarnsCommand(msg *meido.DiscordMessage) {
	if msg.LenArgs() < 2 {
		return
	}

	var (
		err        error
		targetUser *discordgo.User
	)

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

	msg.Reply(fmt.Sprintf("Cleared all active warns issued to %v", targetUser.Mention()))
}
