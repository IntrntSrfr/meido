package moderation

import (
	"bytes"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/internal/helpers"
	"github.com/intrntsrfr/meido/pkg/mio"
	"github.com/intrntsrfr/meido/pkg/utils"
	"go.uber.org/zap"
)

func newFilterWordCommand(m *Module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:           m,
		Name:          "filterword",
		Description:   "Adds or removes a word or phrase to the server filter.",
		Triggers:      []string{"m?fw", "m?filterword"},
		Usage:         "m?fw jeff",
		Cooldown:      2,
		RequiredPerms: discordgo.PermissionManageMessages,
		RequiresOwner: false,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           m.filterwordCommand,
	}
}
func (m *Module) filterwordCommand(msg *mio.DiscordMessage) {
	if msg.LenArgs() < 2 {
		return
	}

	phrase := strings.Join(msg.Args()[1:], " ")
	phrase = strings.ToLower(phrase)

	f, err := m.db.GetGuildFilterByPhrase(msg.GuildID(), phrase)
	switch err {
	case nil:
		if err := m.db.DeleteGuildFilter(f.UID); err != nil {
			_, _ = msg.Reply("There was an issue, please try again!")
			return
		}
		_, _ = msg.Reply(fmt.Sprintf("Removed `%v` from the filter.", phrase))
	case sql.ErrNoRows:
		if err := m.db.CreateGuildFilter(msg.GuildID(), phrase); err != nil {
			_, _ = msg.Reply("There was an issue, please try again!")
			return
		}
		_, _ = msg.Reply(fmt.Sprintf("Added `%v` to the filter.", phrase))
	default:
		_, _ = msg.Reply("There was an issue, please try again!")
	}
}

func newFilterWordListCommand(m *Module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:           m,
		Name:          "filterwordlist",
		Description:   "Lists of all filtered phrases for this server",
		Triggers:      []string{"m?fwl", "m?filterwordlist"},
		Usage:         "m?fwl",
		Cooldown:      10,
		RequiredPerms: discordgo.PermissionManageMessages,
		RequiresOwner: false,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           m.filterwordlistCommand,
	}
}
func (m *Module) filterwordlistCommand(msg *mio.DiscordMessage) {
	if msg.LenArgs() < 1 {
		return
	}

	filterEntries, err := m.db.GetGuildFilters(msg.GuildID())
	if err != nil {
		_, _ = msg.Reply("There was an issue, please try again!")
		return
	}
	if len(filterEntries) == 0 {
		_, _ = msg.Reply("The filter is empty")
		return
	}

	builder := strings.Builder{}
	builder.WriteString("Filtered phrases:\n\n")
	for i, fe := range filterEntries {
		if (i+1)%10 == 0 {
			builder.WriteRune('\n')
		}
		builder.WriteString(fmt.Sprintf("`%s`, ", fe.Phrase))
	}

	if len(builder.String()) > 1000 {
		_, _ = msg.Sess.ChannelFileSend(msg.Message.ChannelID, "filter.txt", bytes.NewBufferString(builder.String()))
		return
	}
	_, _ = msg.Reply(builder.String())
}

func newClearFilterCommand(m *Module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:           m,
		Name:          "clearfilter",
		Description:   "Removes all phrases from the server filter",
		Triggers:      []string{"m?clearfilter"},
		Usage:         "m?clearfilter",
		Cooldown:      10,
		RequiredPerms: discordgo.PermissionAdministrator,
		RequiresOwner: false,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           m.clearfilterCommand,
	}
}

func (m *Module) clearfilterCommand(msg *mio.DiscordMessage) {
	if msg.LenArgs() < 1 {
		return
	}
	rpl, err := msg.Reply("Are you sure you want to REMOVE ALL FILTERS? Please reply `YES`, in all caps, if you are")
	if err != nil {
		_, _ = msg.Reply("There was an issue, please try again!")
		return
	}
	ch, err := m.Bot.Callbacks.Make(fmt.Sprintf("%v:%v", msg.ChannelID(), msg.AuthorID()))
	if err != nil {
		_, _ = msg.Reply("There was an issue, please try again!")
		return
	}
	defer m.Bot.Callbacks.Delete(fmt.Sprintf("%v:%v", msg.ChannelID(), msg.AuthorID()))

	var reply *mio.DiscordMessage
	t := time.NewTimer(time.Second * 15)
	for {
		select {
		case reply = <-ch:
		case <-t.C:
			_ = msg.Sess.ChannelMessageDelete(rpl.ChannelID, rpl.ID)
			_ = msg.Sess.ChannelMessageDelete(msg.ChannelID(), msg.Message.ID)
			return
		}
		if strings.ToLower(reply.RawContent()) == "YES" {
			_ = msg.Sess.ChannelMessageDelete(reply.ChannelID(), reply.Message.ID)
			_ = msg.Sess.ChannelMessageDelete(msg.ChannelID(), msg.Message.ID)
			break
		}
	}
	if err = m.db.DeleteGuildFilters(msg.GuildID()); err != nil {
		_, _ = msg.Reply("There was an issue, please try again!")
		m.Log.Error("unable to delete guild filters", zap.Any("message", msg))
		return
	}
	_, _ = msg.Reply("All filters successfully deleted")
}

func newModerationSettingsCommand(m *Module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:           m,
		Name:          "moderationsettings",
		Description:   "Moderation settings:\n- Toggle warn system [enable / disable]\n- Set max warns [0 - 10]\n- Set warn duration [0 (forever) - 365]",
		Triggers:      []string{"m?settings moderation"},
		Usage:         "m?settings moderation warns [enable / disable]\nm?settings moderation maxwarns [0 - 10]\nm?settings moderation warnduration [0 - 365]",
		Cooldown:      2,
		RequiredPerms: discordgo.PermissionAdministrator,
		RequiresOwner: false,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           m.moderationsettingsCommand,
	}
}

var (
	warnsEnabledText     = map[bool]string{true: "Enabled", false: "Disabled"}
	warnsEnabledSettings = map[string]bool{"enable": true, "disable": false}
)

func (m *Module) moderationsettingsCommand(msg *mio.DiscordMessage) {
	if msg.LenArgs() < 2 {
		return
	}
	gc, err := m.db.GetGuild(msg.GuildID())
	if err != nil {
		_, _ = msg.Reply("There was an issue, please try again!")
		return
	}

	switch msg.LenArgs() {
	case 2:
		embed := helpers.NewEmbed().
			WithTitle("Moderation settings").
			WithOkColor().
			AddField("Warnings", warnsEnabledText[gc.UseWarns], true).
			AddField("Max warnings", fmt.Sprint(gc.MaxWarns), true).
			AddField("Warning duration", fmt.Sprintf("%v days", gc.WarnDuration), true)
		_, _ = msg.ReplyEmbed(embed.Build())
	case 4:
		switch msg.Args()[2] {
		case "warns":
			if v, ok := warnsEnabledSettings[msg.Args()[3]]; ok && gc.UseWarns != v {
				gc.UseWarns = v
				if err := m.db.UpdateGuild(gc); err != nil {
					_, _ = msg.Reply("There was an issue, please try again!")
					return
				}
				_, _ = msg.Reply(fmt.Sprintf("Warnings: %v -> %v", !gc.UseWarns, gc.UseWarns))
			}
		case "maxwarns":
			n, err := strconv.Atoi(msg.Args()[3])
			if err != nil {
				_, _ = msg.Reply("Please provide a number between 0 and 10")
				return
			}

			before := gc.MaxWarns
			n = utils.Clamp(0, 10, n)
			gc.MaxWarns = n
			if err := m.db.UpdateGuild(gc); err != nil {
				_, _ = msg.Reply("There was an issue, please try again!")
				return
			}
			_, _ = msg.Reply(fmt.Sprintf("Max warnings: %v -> %v", before, n))
		case "warnduration":
			n, err := strconv.Atoi(msg.Args()[3])
			if err != nil {
				_, _ = msg.Reply("Please provide a number between 0 and 365")
				return
			}

			before := gc.MaxWarns
			n = utils.Clamp(0, 365, n)
			gc.WarnDuration = n
			if err := m.db.UpdateGuild(gc); err != nil {
				_, _ = msg.Reply("There was an issue, please try again!")
				return
			}
			_, _ = msg.Reply(fmt.Sprintf("Warn duration: %v days -> %v days", before, n))
		}
	}
}

func newCheckFilterPassive(m *Module) *mio.ModulePassive {
	return &mio.ModulePassive{
		Mod:          m,
		Name:         "checkfilter",
		Description:  "checks if messages contain phrases found in the server filter",
		Enabled:      true,
		AllowedTypes: mio.MessageTypeCreate | mio.MessageTypeUpdate,
		Run: func(msg *mio.DiscordMessage) {
			if msg.LenArgs() < 1 {
				return
			}

			isIllegal, trigger := false, ""
			if perms, err := msg.HasPermissions(discordgo.PermissionManageMessages); err != nil || perms {
				return
			}

			entries, err := m.db.GetGuildFilters(msg.GuildID())
			if err != nil {
				return
			}

			for _, entry := range entries {
				if strings.Contains(strings.ToLower(msg.Message.Content), strings.ToLower(entry.Phrase)) {
					isIllegal, trigger = true, entry.Phrase
					break
				}
			}
			if !isIllegal {
				return
			}

			_ = msg.Sess.ChannelMessageDelete(msg.Message.ChannelID, msg.Message.ID)

			gc, err := m.db.GetGuild(msg.GuildID())
			if err != nil {
				return
			}

			if !gc.UseWarns {
				_, _ = msg.Reply(fmt.Sprintf("%v, you are not allowed to use a banned word/phrase", msg.Message.Author.Mention()))
				return
			}

			reason := "Triggering filter: " + trigger
			warns, err := m.db.GetMemberWarnsIfActive(msg.GuildID(), msg.AuthorID())
			if err != nil {
				return
			}
			warnCount := len(warns)

			g, err := msg.Discord.Guild(msg.GuildID())
			if err != nil {
				return
			}
			cu := msg.Discord.BotUser()

			if err := m.db.CreateMemberWarn(msg.GuildID(), msg.AuthorID(), reason, cu.ID); err != nil {
				return
			}

			userChannel, userChError := msg.Discord.Sess.UserChannelCreate(msg.AuthorID())
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
			if err := msg.Discord.Sess.GuildBanCreateWithReason(g.ID, msg.AuthorID(), fmt.Sprintf("Acquired %v warnings.", gc.MaxWarns), 0); err != nil {
				_, _ = msg.Reply("Failed to ban user!")
				return
			}

			t := time.Now()
			for _, warn := range warns {
				warn.IsValid = false
				warn.ClearedByID = &cu.ID
				warn.ClearedAt = &t
				if err := m.db.UpdateMemberWarn(warn); err != nil {
					m.Log.Error("could not update warn", zap.Error(err), zap.Int("warn ID", warn.UID))
				}
			}
			_, _ = msg.Reply(fmt.Sprintf("%v has been banned after acquiring too many warns. miss them.", msg.Message.Author.Mention()))
		},
	}
}
