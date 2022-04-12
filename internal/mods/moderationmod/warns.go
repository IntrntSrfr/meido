package moderationmod

import (
	"database/sql"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/intrntsrfr/meido/base"
	"github.com/intrntsrfr/meido/database"
	"github.com/intrntsrfr/meido/utils"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"time"
)

func NewWarnCommand(m *ModerationMod) *base.ModCommand {
	return &base.ModCommand{
		Mod:           m,
		Name:          "warn",
		Description:   "Warns a user. Does not work if warn system is disabled.",
		Triggers:      []string{"m?warn", ".warn"},
		Usage:         "m?warn 163454407999094786 stupid",
		Cooldown:      2,
		RequiredPerms: discordgo.PermissionBanMembers,
		RequiresOwner: false,
		CheckBotPerms: true,
		AllowedTypes:  base.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           m.warnCommand,
	}
}

func (m *ModerationMod) warnCommand(msg *base.DiscordMessage) {
	if msg.LenArgs() < 2 {
		return
	}

	dge := &database.Guild{}
	err := m.db.Get(dge, "SELECT use_warns, max_warns FROM guild WHERE guild_id = $1;", msg.Message.GuildID)
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
			msg.Reply("i could not find that member")
			return
		}
	} else {
		_, err := strconv.Atoi(msg.Args()[1])
		if err != nil {
			return
		}
		targetUser, err = msg.Discord.Member(msg.Message.GuildID, msg.Args()[1])
		if err != nil {
			msg.Reply("i could not find that member")
			return
		}
	}

	if targetUser.User.ID == msg.Sess.State.User.ID {
		msg.Reply("no (i will not warn myself)")
		return
	}
	if targetUser.User.ID == msg.Message.Author.ID {
		msg.Reply("no (you can not warn yourself)")
		return
	}

	topUserRole := msg.Discord.HighestRolePosition(msg.Message.GuildID, msg.Message.Author.ID)
	topTargetRole := msg.Discord.HighestRolePosition(msg.Message.GuildID, targetUser.User.ID)
	topBotRole := msg.Discord.HighestRolePosition(msg.Message.GuildID, msg.Sess.State.User.ID)

	if topUserRole <= topTargetRole || topBotRole <= topTargetRole {
		msg.Reply("no (you can only kick users who are below you and me in the role hierarchy)")
		return
	}

	if msg.LenArgs() > 2 {
		reason = strings.Join(msg.RawArgs()[2:], " ")
	}

	warnCount, err := m.db.GetValidUserWarnCount(msg.GuildID(), targetUser.User.ID)
	if err != nil {
		msg.Reply("something went wrong when trying to warn user, please try again")
		return
	}

	g, err := msg.Discord.Guild(msg.Message.GuildID)
	if err != nil {
		msg.Reply("something went wrong when trying to warn user, please try again")
		return
	}

	err = m.db.InsertWarn(msg.GuildID(), targetUser.User.ID, reason, msg.Author().ID)
	if err != nil {
		msg.Reply("something went wrong when trying to warn user, please try again")
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

		m.db.ClearActiveUserWarns(msg.GuildID(), targetUser.User.ID, msg.Sess.State.User.ID)
		msg.Reply(fmt.Sprintf("%v has been banned after acquiring too many warns", targetUser.Mention()))

	} else {
		if userChError == nil {
			msg.Discord.Sess.ChannelMessageSend(userChannel.ID, fmt.Sprintf("You have been warned in %v.\nWarned for: %v\nYou are currently at warn %v/%v",
				g.Name, reason, warnCount+1, dge.MaxWarns))
		}
		msg.Reply(fmt.Sprintf("%v has been warned\nThey are currently at warn %v/%v", targetUser.Mention(), warnCount+1, dge.MaxWarns))
	}
}

func NewWarnLogCommand(m *ModerationMod) *base.ModCommand {
	return &base.ModCommand{
		Mod:           m,
		Name:          "warnlog",
		Description:   "Displays a users warns",
		Triggers:      []string{"m?warnlog"},
		Usage:         "m?warnlog 163454407999094786",
		Cooldown:      5,
		RequiredPerms: discordgo.PermissionManageMessages,
		RequiresOwner: false,
		AllowedTypes:  base.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           m.warnlogCommand,
	}
}

func (m *ModerationMod) warnlogCommand(msg *base.DiscordMessage) {
	if msg.LenArgs() < 2 {
		return
	}

	page := 0
	var err error

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
		_, err = strconv.Atoi(msg.Args()[1])
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

	// todo: make this a method
	var warns []*database.Warn
	err = m.db.Select(&warns, "SELECT * FROM warn WHERE user_id=$1 AND guild_id=$2 ORDER BY given_at DESC;", targetUser.ID, msg.Message.GuildID)
	if err != nil {
		fmt.Println(err)
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

		warns = warns[page*10 : utils.Min(page*10+10, len(warns))]

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

func NewWarnCountCommand(m *ModerationMod) *base.ModCommand {
	return &base.ModCommand{
		Mod:           m,
		Name:          "warncount",
		Description:   "Displays how many warns a user has. User can be specified. Message author will be used if no user is provided.",
		Triggers:      []string{"m?warncount"},
		Usage:         "m?warncount <user>",
		Cooldown:      2,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  base.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           m.warncountCommand,
	}
}

func (m *ModerationMod) warncountCommand(msg *base.DiscordMessage) {

	dge := database.Guild{}
	err := m.db.Get(&dge, "SELECT use_warns, max_warns FROM guild WHERE guild_id=$1", msg.Message.GuildID)
	if err != nil {
		return
	}

	if !dge.UseWarns {
		msg.Reply("warn system not enabled")
		return
	}

	targetUser := msg.Author()
	if msg.LenArgs() > 1 {
		targetUser, err = msg.GetMemberOrUserAtArg(1)
		if err != nil {
			_, _ = msg.Reply("Could not find that user!")
			return
		}
	}

	var warnCount int
	err = m.db.Get(&warnCount, "SELECT COUNT(*) FROM warn WHERE guild_id=$1 AND user_id=$2 AND is_valid", msg.Message.GuildID, targetUser.ID)
	if err != nil {
		return
	}

	msg.Reply(fmt.Sprintf("%v is at %v/%v warns", targetUser.String(), warnCount, dge.MaxWarns))
}

func NewClearWarnCommand(m *ModerationMod) *base.ModCommand {
	return &base.ModCommand{
		Mod:           m,
		Name:          "pardon",
		Description:   "Pardons a user. Opens a menu to clear a warn belonging to them.",
		Triggers:      []string{"m?pardon"},
		Usage:         "m?pardon <user>",
		Cooldown:      3,
		RequiredPerms: discordgo.PermissionBanMembers,
		RequiresOwner: false,
		AllowedTypes:  base.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           m.clearwarnCommand,
	}
}

func (m *ModerationMod) clearwarnCommand(msg *base.DiscordMessage) {
	if msg.LenArgs() < 2 {
		return
	}

	targetMember, err := msg.GetMemberAtArg(1)
	if err != nil {
		return
	}

	var entries []*database.Warn
	err = m.db.Select(&entries, "SELECT * FROM warn WHERE user_id=$1 AND guild_id=$2 AND is_valid", targetMember.User.ID, msg.Message.GuildID)
	if err != nil && err != sql.ErrNoRows {
		m.log.Error("could not get valid warns", zap.Error(err))
		msg.Reply("there was an error, please try again")
		return
	} else if err == sql.ErrNoRows || len(entries) == 0 {
		msg.Reply("User has active no warns")
		return
	}

	sb := strings.Builder{}
	sb.WriteString("Which warn would you like to remove? [1-10]\n")
	for i, entry := range entries {
		sb.WriteString(fmt.Sprintf("`%2v.` | '%v' given %v\n", i+1, entry.Reason, humanize.Time(entry.GivenAt)))
	}
	sb.WriteString("\nType `cancel` to exit")
	menu, err := msg.Reply(sb.String())
	if err != nil {
		msg.Reply("Something went wrong, please try again!")
		return
	}

	key := fmt.Sprintf("%v:%v", msg.ChannelID(), msg.AuthorID())
	cb, err := m.bot.Callbacks.Make(key)
	if err != nil {
		return
	}

	var n int
	var reply *base.DiscordMessage
	// this needs a timeout
	for {
		select {
		case reply = <-cb:
		case <-time.After(time.Second * 30):
			//msg.Reply("You spent too much time")
			m.bot.Callbacks.Delete(key)
			msg.Sess.ChannelMessageDelete(menu.ChannelID, menu.ID)
			return
		}

		if strings.ToLower(reply.RawContent()) == "cancel" {
			m.bot.Callbacks.Delete(key)
			msg.Sess.ChannelMessageDelete(menu.ChannelID, menu.ID)
			msg.Sess.ChannelMessageDelete(reply.Message.ChannelID, reply.Message.ID)
			return
		}

		n, err = strconv.Atoi(reply.RawContent())
		if err == nil && n-1 >= 0 && n-1 < len(entries) {
			break
		}
	}

	m.bot.Callbacks.Delete(key)
	msg.Sess.ChannelMessageDelete(menu.ChannelID, menu.ID)
	msg.Sess.ChannelMessageDelete(reply.Message.ChannelID, reply.Message.ID)

	selectedEntry := entries[n-1]

	_, err = m.db.Exec("UPDATE warn SET is_valid=false, cleared_by_id=$1, cleared_at=$2 WHERE uid=$3 AND is_valid", msg.Message.Author.ID, time.Now(), selectedEntry.UID)
	if err != nil {
		msg.Reply("there was an error, please try again")
		return
	}

	msg.Reply(fmt.Sprintf("Invalidated warn with ID: %v", selectedEntry.UID))
}

func NewClearAllWarnsCommand(m *ModerationMod) *base.ModCommand {
	return &base.ModCommand{
		Mod:           m,
		Name:          "pardonall",
		Description:   "Pardons all active warns for a member",
		Triggers:      []string{"m?pardonall"},
		Usage:         "m?pardonall <user>",
		Cooldown:      5,
		RequiredPerms: discordgo.PermissionBanMembers,
		RequiresOwner: false,
		AllowedTypes:  base.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           m.clearallwarnsCommand,
	}
}
func (m *ModerationMod) clearallwarnsCommand(msg *base.DiscordMessage) {
	if msg.LenArgs() < 2 {
		return
	}

	targetMember, err := msg.GetMemberAtArg(1)
	if err != nil {
		msg.Reply("Could not find that member")
		return
	}

	// TODO: add confirmation menu

	_, err = m.db.Exec("UPDATE warn SET is_valid=false, cleared_by_id=$1, cleared_at=$2 WHERE user_id=$3 AND guild_id=$4 AND is_valid",
		msg.Message.Author.ID, time.Now(), targetMember.User.ID, msg.Message.GuildID)
	if err != nil {
		msg.Reply("there was an error, please try again")
		return
	}

	msg.Reply(fmt.Sprintf("Cleared all active warns issued to %v", targetMember.Mention()))
}
