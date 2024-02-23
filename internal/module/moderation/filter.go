package moderation

import (
	"bytes"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/pkg/mio/bot"
	"github.com/intrntsrfr/meido/pkg/mio/discord"
	"github.com/intrntsrfr/meido/pkg/utils"
	"github.com/intrntsrfr/meido/pkg/utils/builders"
	"go.uber.org/zap"
)

func newFilterWordCommand(m *module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "filterword",
		Description:      "Adds or removes a word or phrase to the server filter.",
		Triggers:         []string{"m?fw", "m?filterword"},
		Usage:            "m?fw jeff",
		Cooldown:         2,
		CooldownScope:    bot.CooldownScopeChannel,
		RequiredPerms:    discordgo.PermissionManageMessages,
		CheckBotPerms:    false,
		RequiresUserType: bot.UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         false,
		Enabled:          true,
		Execute:          m.filterwordCommand,
	}
}
func (m *module) filterwordCommand(msg *discord.DiscordMessage) {
	if len(msg.Args()) < 2 {
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

func newFilterWordListCommand(m *module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "filterwordlist",
		Description:      "Lists of all filtered phrases for this server",
		Triggers:         []string{"m?fwl", "m?filterwordlist"},
		Usage:            "m?fwl",
		Cooldown:         10,
		CooldownScope:    bot.CooldownScopeChannel,
		RequiredPerms:    discordgo.PermissionManageMessages,
		CheckBotPerms:    false,
		RequiresUserType: bot.UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         false,
		Enabled:          true,
		Execute:          m.filterwordlistCommand,
	}
}
func (m *module) filterwordlistCommand(msg *discord.DiscordMessage) {
	if len(msg.Args()) < 1 {
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

func newClearFilterCommand(m *module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "clearfilter",
		Description:      "Removes all phrases from the server filter",
		Triggers:         []string{"m?clearfilter"},
		Usage:            "m?clearfilter",
		Cooldown:         10,
		CooldownScope:    bot.CooldownScopeChannel,
		RequiredPerms:    discordgo.PermissionAdministrator,
		CheckBotPerms:    false,
		RequiresUserType: bot.UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         false,
		Enabled:          true,
		Execute:          m.clearfilterCommand,
	}
}

func (m *module) clearfilterCommand(msg *discord.DiscordMessage) {
	if len(msg.Args()) < 1 {
		return
	}
	rpl, err := msg.Reply("Are you sure you want to REMOVE ALL FILTERS? Please reply `YES`, in all caps, if you are")
	if err != nil {
		_, _ = msg.Reply("There was an issue, please try again!")
		return
	}
	cb, err := m.Bot.Callbacks.Make(fmt.Sprintf("%v:%v", msg.ChannelID(), msg.AuthorID()))
	if err != nil {
		_, _ = msg.Reply("There was an issue, please try again!")
		return
	}
	defer m.Bot.Callbacks.Delete(fmt.Sprintf("%v:%v", msg.ChannelID(), msg.AuthorID()))

	var reply *discord.DiscordMessage
	t := time.NewTimer(time.Second * 15)
	for {
		select {
		case reply = <-cb:
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
		m.Logger.Error("Deleting guild filters failed", zap.Any("message", msg))
		return
	}
	_, _ = msg.Reply("All filters successfully deleted")
}

func newModerationSettingsCommand(m *module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "moderationsettings",
		Description:      "Moderation settings:\n- Toggle warn system [enable / disable]\n- Set max warns [0 - 10]\n- Set warn duration [0 (forever) - 365]",
		Triggers:         []string{"m?settings moderation"},
		Usage:            "m?settings moderation warns [enable / disable]\nm?settings moderation maxwarns [0 - 10]\nm?settings moderation warnduration [0 - 365]",
		Cooldown:         2,
		CooldownScope:    bot.CooldownScopeChannel,
		RequiredPerms:    discordgo.PermissionAdministrator,
		CheckBotPerms:    false,
		RequiresUserType: bot.UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         false,
		Enabled:          true,
		Execute:          m.moderationsettingsCommand,
	}
}

var (
	warnsEnabledText     = map[bool]string{true: "Enabled", false: "Disabled"}
	warnsEnabledSettings = map[string]bool{"enable": true, "disable": false}
)

func (m *module) moderationsettingsCommand(msg *discord.DiscordMessage) {
	if len(msg.Args()) < 2 {
		return
	}
	gc, err := m.db.GetGuild(msg.GuildID())
	if err != nil {
		_, _ = msg.Reply("There was an issue, please try again!")
		return
	}

	switch len(msg.Args()) {
	case 2:
		embed := builders.NewEmbedBuilder().
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

func newCheckFilterPassive(m *module) *bot.ModulePassive {
	return &bot.ModulePassive{
		Mod:          m,
		Name:         "checkfilter",
		Description:  "checks if messages contain phrases found in the server filter",
		Enabled:      true,
		AllowedTypes: discord.MessageTypeCreate | discord.MessageTypeUpdate,
		Execute: func(msg *discord.DiscordMessage) {
			if len(msg.Args()) < 1 {
				return
			}

			isIllegal, trigger := false, ""
			if perms, err := msg.AuthorHasPermissions(discordgo.PermissionManageMessages); err != nil || perms {
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
					m.Logger.Error("Updating warn failed", zap.Error(err), zap.Int("warnID", warn.UID))
				}
			}
			_, _ = msg.Reply(fmt.Sprintf("%v has been banned after acquiring too many warns. miss them.", msg.Message.Author.Mention()))
		},
	}
}
