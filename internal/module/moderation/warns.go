package moderation

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	iutils "github.com/intrntsrfr/meido/internal/utils"
	"github.com/intrntsrfr/meido/pkg/mio"
	"go.uber.org/zap"
)

func newWarnCommand(m *Module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:              m,
		Name:             "warn",
		Description:      "Warns a user. Requires warnings enabled.",
		Triggers:         []string{"m?warn", ".warn"},
		Usage:            "m?warn [user] <reason>",
		Cooldown:         2,
		CooldownScope:    mio.Channel,
		RequiredPerms:    discordgo.PermissionBanMembers,
		CheckBotPerms:    true,
		RequiresUserType: mio.UserTypeAny,
		AllowedTypes:     mio.MessageTypeCreate,
		AllowDMs:         false,
		Enabled:          true,
		Run:              m.warnCommand,
	}
}

func (m *Module) warnCommand(msg *mio.DiscordMessage) {
	if len(msg.Args()) < 2 {
		return
	}

	gc, err := m.db.GetGuild(msg.GuildID())
	if err != nil {
		_, _ = msg.Reply("There was an issue, please try again!")
		return
	}

	if !gc.UseWarns {
		_, _ = msg.Reply("Warnings are not enabled")
		return
	}

	reason := "No reason"
	targetMember, err := msg.GetMemberAtArg(1)
	if err != nil {
		_, _ = msg.Reply("Could not find that user!")
		return
	}

	if targetMember.User.ID == msg.Discord.BotUser().ID {
		_, _ = msg.Reply("no (I will not warn myself)")
		return
	}
	if targetMember.User.ID == msg.AuthorID() {
		_, _ = msg.Reply("no (you can not warn yourself)")
		return
	}

	topUserRole := msg.Discord.HighestRolePosition(msg.Message.GuildID, msg.Message.Author.ID)
	topTargetRole := msg.Discord.HighestRolePosition(msg.Message.GuildID, targetMember.User.ID)
	topBotRole := msg.Discord.HighestRolePosition(msg.Message.GuildID, msg.Sess.State().User.ID)
	if topUserRole <= topTargetRole || topBotRole <= topTargetRole {
		_, _ = msg.Reply("no (you can only kick users who are below you and me in the role hierarchy)")
		return
	}

	if len(msg.Args()) > 2 {
		reason = strings.Join(msg.RawArgs()[2:], " ")
	}

	warns, err := m.db.GetMemberWarnsIfActive(msg.GuildID(), targetMember.User.ID)
	if err != nil {
		return
	}
	warnCount := len(warns)

	g, err := msg.Discord.Guild(msg.GuildID())
	if err != nil {
		_, _ = msg.Reply("There was an issue, please try again!")
		return
	}

	if err := m.db.CreateMemberWarn(msg.GuildID(), targetMember.User.ID, reason, msg.AuthorID()); err != nil {
		_, _ = msg.Reply("There was an issue, please try again!")
		return
	}

	userChannel, userChError := msg.Discord.Sess.UserChannelCreate(targetMember.User.ID)
	if warnCount+1 < gc.MaxWarns {
		if userChError == nil {
			_, _ = msg.Discord.Sess.ChannelMessageSend(userChannel.ID, fmt.Sprintf("You have been warned in %v.\nYou were warned for: %v\nYou now have %v/%v warnings",
				g.Name, reason, warnCount+1, gc.MaxWarns))
		}
		_, _ = msg.Reply(fmt.Sprintf("%v has been warned\nThey now have %v/%v warnings", msg.Author().Mention(), warnCount+1, gc.MaxWarns))
		return
	}

	if userChError == nil {
		_, _ = msg.Discord.Sess.ChannelMessageSend(userChannel.ID, fmt.Sprintf("You have been banned from %v for acquiring %v warnings.\nLast warning was: %v",
			g.Name, gc.MaxWarns, reason))
	}
	if err := msg.Discord.Sess.GuildBanCreateWithReason(g.ID, targetMember.User.ID, fmt.Sprintf("Acquired %v warnings.", gc.MaxWarns), 0); err != nil {
		_, _ = msg.Reply("Failed to ban user!")
		return
	}

	t := time.Now()
	for _, warn := range warns {
		warn.IsValid = false
		warn.ClearedByID = &msg.Discord.BotUser().ID
		warn.ClearedAt = &t
		if err := m.db.UpdateMemberWarn(warn); err != nil {
			m.Logger.Error("could not update warn", zap.Error(err), zap.Int("warnID", warn.UID))
		}
	}
}

func newWarnLogCommand(m *Module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:              m,
		Name:             "warnlog",
		Description:      "Displays a users warns",
		Triggers:         []string{"m?warnlog"},
		Usage:            "m?warnlog [user] <page>",
		Cooldown:         5,
		CooldownScope:    mio.Channel,
		RequiredPerms:    discordgo.PermissionManageMessages,
		CheckBotPerms:    false,
		RequiresUserType: mio.UserTypeAny,
		AllowedTypes:     mio.MessageTypeCreate,
		AllowDMs:         false,
		Enabled:          true,
		Run:              m.warnlogCommand,
	}
}

func (m *Module) warnlogCommand(msg *mio.DiscordMessage) {
	if len(msg.Args()) < 2 {
		return
	}
	page := 0
	var err error
	if len(msg.Args()) > 2 {
		page, err = strconv.Atoi(msg.Args()[2])
		if err != nil || page < 1 {
			_, _ = msg.Reply("Invalid page")
			return
		}
		page--
	}

	targetUser, err := msg.GetMemberOrUserAtArg(1)
	if err != nil {
		_, _ = msg.Reply("Could not find that user!")
		return
	}

	warns, err := m.db.GetMemberWarns(msg.GuildID(), targetUser.ID)
	if err != nil {
		_, _ = msg.Reply("There was an issue, please try again!")
		return
	}

	embed := iutils.NewEmbed().
		WithTitle(fmt.Sprintf("Warnings issued to %v", targetUser.String())).
		WithOkColor().
		WithFooter(fmt.Sprintf("Page %v", page+1), "")

	if len(warns) <= 0 {
		embed.WithDescription("No warns")
		_, _ = msg.ReplyEmbed(embed.Build())
		return
	}
	if page*10 > len(warns) {
		_, _ = msg.Reply("Page does not exist.")
		return
	}

	warns = warns[page*10 : min(page*10+10, len(warns))]
	userCache := make(map[string]*discordgo.User) // to cache users in case someone authored or cleared multiple
	for _, warn := range warns {
		field := &discordgo.MessageEmbedField{}
		field.Value = warn.Reason

		gb, ok := userCache[warn.GivenByID]
		if !ok {
			gb, err = msg.Discord.Sess.User(warn.GivenByID)
			if err != nil {
				_, _ = msg.Reply("There was an issue, please try again!")
				continue
			}
			userCache[warn.GivenByID] = gb
		}

		if warn.IsValid {
			field.Name = fmt.Sprintf("ID: %v | Issued by %v (%v) %v", warn.UID, gb.String(), gb.ID, humanize.Time(warn.GivenAt))
			embed.Fields = append(embed.Fields, field)
			continue
		}
		if warn.ClearedByID == nil {
			continue
		}

		cb, ok := userCache[*warn.ClearedByID]
		if !ok {
			cb, err = msg.Discord.Sess.User(*warn.ClearedByID)
			if err != nil {
				_, _ = msg.Reply("There was an issue, please try again!")
				continue
			}
			userCache[*warn.ClearedByID] = cb
		}
		field.Name = fmt.Sprintf("ID: %v | !CLEARED! | Cleared by %v (%v) %v", warn.UID, cb.String(), cb.ID, humanize.Time(*warn.ClearedAt))
		embed.Fields = append(embed.Fields, field)
	}
	_, _ = msg.ReplyEmbed(embed.Build())
}

func newWarnCountCommand(m *Module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:              m,
		Name:             "warncount",
		Description:      "Displays how many warns a user has. User can be specified. Message author will be used if no user is provided.",
		Triggers:         []string{"m?warncount"},
		Usage:            "m?warncount <user>",
		Cooldown:         2,
		CooldownScope:    mio.Channel,
		RequiredPerms:    0,
		CheckBotPerms:    false,
		RequiresUserType: mio.UserTypeAny,
		AllowedTypes:     mio.MessageTypeCreate,
		AllowDMs:         false,
		Enabled:          true,
		Run:              m.warncountCommand,
	}
}

func (m *Module) warncountCommand(msg *mio.DiscordMessage) {
	gc, err := m.db.GetGuild(msg.GuildID())
	if err != nil {
		_, _ = msg.Reply("There was an issue, please try again!")
		return
	}

	if !gc.UseWarns {
		_, _ = msg.Reply("Warnings are not enabled")
		return
	}

	targetUser := msg.Author()
	if len(msg.Args()) > 1 {
		targetUser, err = msg.GetMemberOrUserAtArg(1)
		if err != nil {
			_, _ = msg.Reply("Could not find that user!")
			return
		}
	}
	warns, err := m.db.GetMemberWarnsIfActive(msg.GuildID(), targetUser.ID)
	if err != nil {
		return
	}
	_, _ = msg.Reply(fmt.Sprintf("%v is at %v/%v warns", targetUser.String(), len(warns), gc.MaxWarns))
}

func newClearWarnCommand(m *Module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:              m,
		Name:             "pardon",
		Description:      "Pardons a user. Opens a menu to clear a warn belonging to them.",
		Triggers:         []string{"m?pardon"},
		Usage:            "m?pardon <user>",
		Cooldown:         3,
		CooldownScope:    mio.Channel,
		RequiredPerms:    discordgo.PermissionBanMembers,
		CheckBotPerms:    false,
		RequiresUserType: mio.UserTypeAny,
		AllowedTypes:     mio.MessageTypeCreate,
		AllowDMs:         false,
		Enabled:          true,
		Run:              m.clearwarnCommand,
	}
}

func (m *Module) clearwarnCommand(msg *mio.DiscordMessage) {
	if len(msg.Args()) < 2 {
		return
	}

	targetMember, err := msg.GetMemberAtArg(1)
	if err != nil {
		return
	}

	entries, err := m.db.GetMemberWarnsIfActive(msg.GuildID(), targetMember.User.ID)
	if err != nil && err != sql.ErrNoRows {
		_, _ = msg.Reply("There was an issue, please try again!")
		return
	} else if err == sql.ErrNoRows || len(entries) == 0 {
		_, _ = msg.Reply("User has active no warns")
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
		_, _ = msg.Reply("There was an issue, please try again!")
		return
	}

	key := fmt.Sprintf("%v:%v", msg.ChannelID(), msg.AuthorID())
	cb, err := m.Bot.Callbacks.Make(key)
	if err != nil {
		return
	}

	var n int
	var reply *mio.DiscordMessage
	// this needs a timeout
	for {
		select {
		case reply = <-cb:
		case <-time.After(time.Second * 30):
			m.Bot.Callbacks.Delete(key)
			_ = msg.Sess.ChannelMessageDelete(menu.ChannelID, menu.ID)
			return
		}

		if strings.ToLower(reply.RawContent()) == "cancel" {
			m.Bot.Callbacks.Delete(key)
			_ = msg.Sess.ChannelMessageDelete(menu.ChannelID, menu.ID)
			_ = msg.Sess.ChannelMessageDelete(reply.Message.ChannelID, reply.Message.ID)
			return
		}

		n, err = strconv.Atoi(reply.RawContent())
		if err == nil && n-1 >= 0 && n-1 < len(entries) {
			break
		}
	}

	m.Bot.Callbacks.Delete(key)
	_ = msg.Sess.ChannelMessageDelete(menu.ChannelID, menu.ID)
	_ = msg.Sess.ChannelMessageDelete(reply.Message.ChannelID, reply.Message.ID)

	selectedEntry := entries[n-1]
	t := time.Now()
	selectedEntry.IsValid = false
	selectedEntry.ClearedByID = &msg.Message.Author.ID
	selectedEntry.ClearedAt = &t
	selectedEntry.IsValid = false
	if err := m.db.UpdateMemberWarn(selectedEntry); err != nil {
		_, _ = msg.Reply("Failed to update warning, please try again")
		return
	}
	_, _ = msg.Reply("Updated warning")
}

func newClearAllWarnsCommand(m *Module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:              m,
		Name:             "pardonall",
		Description:      "Pardons all active warns for a member",
		Triggers:         []string{"m?pardonall"},
		Usage:            "m?pardonall <user>",
		Cooldown:         5,
		CooldownScope:    mio.Channel,
		RequiredPerms:    discordgo.PermissionBanMembers,
		CheckBotPerms:    false,
		RequiresUserType: mio.UserTypeAny,
		AllowedTypes:     mio.MessageTypeCreate,
		AllowDMs:         false,
		Enabled:          true,
		Run:              m.clearallwarnsCommand,
	}
}
func (m *Module) clearallwarnsCommand(msg *mio.DiscordMessage) {
	if len(msg.Args()) < 2 {
		return
	}

	targetMember, err := msg.GetMemberAtArg(1)
	if err != nil {
		_, _ = msg.Reply("Could not find that member")
		return
	}

	warns, err := m.db.GetMemberWarns(msg.GuildID(), targetMember.User.ID)
	if err != nil {
		_, _ = msg.Reply("There was an issue, please try again!")
		return
	}

	t := time.Now()
	for _, warn := range warns {
		warn.IsValid = false
		warn.ClearedByID = &msg.Message.Author.ID
		warn.ClearedAt = &t
		if err := m.db.UpdateMemberWarn(warn); err != nil {
			m.Logger.Error("could not update warn", zap.Error(err), zap.Int("warnID", warn.UID))
		}
	}
	_, _ = msg.Reply(fmt.Sprintf("Cleared active warns issued to %v", targetMember.Mention()))
}
