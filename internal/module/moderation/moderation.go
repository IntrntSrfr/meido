package moderation

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/internal/database"
	"github.com/intrntsrfr/meido/internal/helpers"
	"github.com/intrntsrfr/meido/pkg/mio"
	"github.com/intrntsrfr/meido/pkg/utils"
	"go.uber.org/zap"
)

type Module struct {
	*mio.ModuleBase
	db database.DB
}

func New(b *mio.Bot, db database.DB, logger *zap.Logger) mio.Module {
	return &Module{
		ModuleBase: mio.NewModule(b, "Moderation", logger.Named("moderation")),
		db:         db,
	}
}

func (m *Module) Hook() error {
	m.Bot.Discord.AddEventHandlerOnce(checkWarnInterval(m))
	m.Bot.Discord.AddEventHandler(addAutoRoleOnJoin(m))

	err := m.RegisterPassives([]*mio.ModulePassive{
		newCheckFilterPassive(m),
	})
	if err != nil {
		return err
	}

	return m.RegisterCommands([]*mio.ModuleCommand{
		newBanCommand(m),
		newUnbanCommand(m),
		newHackbanCommand(m),
		newKickCommand(m),
		newWarnCommand(m),
		newClearWarnCommand(m),
		newWarnLogCommand(m),
		newClearAllWarnsCommand(m),
		newWarnCountCommand(m),
		newFilterWordCommand(m),
		newClearFilterCommand(m),
		newFilterWordListCommand(m),
		newModerationSettingsCommand(m),
		newLockdownChannelCommand(m),
		newUnlockChannelCommand(m),
		newMuteCommand(m),
		newUnmuteCommand(m),
		newSetAutoRoleCommand(m),
		newRemoveAutoRoleCommand(m),
		newPruneCommand(m),
	})
}

func checkWarnInterval(m *Module) func(s *discordgo.Session, r *discordgo.Ready) {
	return func(s *discordgo.Session, r *discordgo.Ready) {
		refreshTicker := time.NewTicker(time.Hour)
		go func() {
			for range refreshTicker.C {
				m.Log.Info("running warn check")
				for _, g := range m.Bot.Discord.Guilds() {
					if g.Unavailable {
						continue
					}
					gc, err := m.db.GetGuild(g.ID)
					if err != nil || gc.WarnDuration <= 0 {
						continue
					}

					warns, err := m.db.GetGuildWarnsIfActive(g.ID)
					if err != nil {
						continue
					}

					dur := time.Duration(gc.WarnDuration)
					for _, warn := range warns {
						if time.Since(warn.GivenAt) > dur {
							t := time.Now()
							warn.IsValid = false
							warn.ClearedByID = &m.Bot.Discord.Sess.State.User.ID
							warn.ClearedAt = &t
							if err := m.db.UpdateMemberWarn(warn); err != nil {
								m.Log.Error("could not update warn", zap.Error(err), zap.Int("warn UID", warn.UID))
							}
							//m.db.Exec("UPDATE warn SET is_valid=false, cleared_by_id=$1, cleared_at=$2 WHERE uid=$3",
							//	m.bot.Discord.Sess.State.User.ID, time.Now(), warn.UID)
						}
					}
				}
			}
		}()
	}
}

func newBanCommand(m *Module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:              m,
		Name:             "ban",
		Description:      "Bans a user. Days of messages to be deleted and reason is optional",
		Triggers:         []string{"m?ban", "m?b", ".b", ".ban"},
		Usage:            ".b [user] <days> <reason>",
		Cooldown:         2,
		CooldownScope:    mio.Channel,
		RequiredPerms:    discordgo.PermissionBanMembers,
		CheckBotPerms:    true,
		RequiresUserType: mio.UserTypeAny,
		AllowedTypes:     mio.MessageTypeCreate,
		AllowDMs:         false,
		IsEnabled:        true,
		Run:              m.banCommand,
	}
}

func (m *Module) banCommand(msg *mio.DiscordMessage) {
	if len(msg.Args()) < 2 {
		return
	}

	reason := ""
	pruneDays := 0

	targetUser, err := msg.GetUserAtArg(1)
	if err != nil {
		_, _ = msg.Reply("Could not find that user!")
		return
	}

	if targetUser.ID == msg.Sess.State.User.ID {
		_, _ = msg.Reply("no (i can not ban myself)")
		return
	}
	if targetUser.ID == msg.AuthorID() {
		_, _ = msg.Reply("no (you can not ban yourself)")
		return
	}

	if len(msg.Args()) > 2 {
		reason = strings.Join(msg.RawArgs()[2:], " ")
		pruneDays, err = strconv.Atoi(msg.Args()[2])
		if err == nil {
			reason = strings.Join(msg.RawArgs()[3:], " ")
			if pruneDays > 7 {
				pruneDays = 7
			} else if pruneDays < 0 {
				pruneDays = 0
			}
		}
	}

	topUserRole := msg.Discord.HighestRolePosition(msg.Message.GuildID, msg.Message.Author.ID)
	topTargetRole := msg.Discord.HighestRolePosition(msg.Message.GuildID, targetUser.ID)
	topBotRole := msg.Discord.HighestRolePosition(msg.Message.GuildID, msg.Sess.State.User.ID)

	if topUserRole <= topTargetRole || topBotRole <= topTargetRole {
		_, _ = msg.Reply("no (you can only ban users who are below you and me in the role hierarchy)")
		return
	}

	// if user is in the server
	if topTargetRole > 0 {
		userChannel, userChErr := msg.Discord.Sess.UserChannelCreate(targetUser.ID)
		if userChErr == nil {
			g, err := msg.Discord.Guild(msg.Message.GuildID)
			if err != nil {
				return
			}

			if reason == "" {
				_, _ = msg.Sess.ChannelMessageSend(userChannel.ID, fmt.Sprintf("You have been banned from %v", g.Name))
			} else {
				_, _ = msg.Sess.ChannelMessageSend(userChannel.ID, fmt.Sprintf("You have been banned from %v for the following reason:\n%v", g.Name, reason))
			}
		}
	}

	err = msg.Discord.Sess.GuildBanCreateWithReason(msg.Message.GuildID, targetUser.ID, fmt.Sprintf("%v - %v", msg.Message.Author.String(), reason), pruneDays)
	if err != nil {
		_, _ = msg.Reply("I could not ban that user!")
		return
	}

	warns, err := m.db.GetMemberWarns(msg.GuildID(), targetUser.ID)
	if err != nil {
		_, _ = msg.Reply("There was an issue, please try again!")
		return
	}

	t := time.Now()
	for _, warn := range warns {
		warn.IsValid = false
		warn.ClearedByID = &msg.Sess.State.User.ID
		warn.ClearedAt = &t
		if err := m.db.UpdateMemberWarn(warn); err != nil {
			m.Log.Error("could not update warn", zap.Error(err), zap.Int("warn ID", warn.UID))
		}
	}

	embed := helpers.NewEmbed().
		WithTitle("User banned").
		WithOkColor().
		AddField("Username", targetUser.Mention(), true).
		AddField("ID", targetUser.ID, true)
	_, _ = msg.ReplyEmbed(embed.Build())
}

func newUnbanCommand(m *Module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:              m,
		Name:             "unban",
		Description:      "Unbans a user",
		Triggers:         []string{"m?unban", "m?ub", ".ub", ".unban"},
		Usage:            ".unban [userID]",
		Cooldown:         2,
		CooldownScope:    mio.Channel,
		RequiredPerms:    discordgo.PermissionBanMembers,
		CheckBotPerms:    true,
		RequiresUserType: mio.UserTypeAny,
		AllowedTypes:     mio.MessageTypeCreate,
		AllowDMs:         false,
		IsEnabled:        true,
		Run:              m.unbanCommand,
	}
}

func (m *Module) unbanCommand(msg *mio.DiscordMessage) {
	if len(msg.Args()) < 2 {
		return
	}

	_, err := strconv.Atoi(msg.Args()[1])
	if err != nil {
		return
	}

	err = msg.Discord.Sess.GuildBanDelete(msg.Message.GuildID, msg.Args()[1])
	if err != nil {
		return
	}

	targetUser, err := msg.GetUserAtArg(1)
	if err != nil {
		return
	}

	embed := helpers.NewEmbed().
		WithDescription(fmt.Sprintf("**Unbanned** %v - %v#%v (%v)", targetUser.Mention(), targetUser.Username, targetUser.Discriminator, targetUser.ID)).
		WithOkColor()
	_, _ = msg.ReplyEmbed(embed.Build())
}

func newHackbanCommand(m *Module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:              m,
		Name:             "hackban",
		Description:      "Hackbans one or several users. Prunes 7 days. Only accepts user IDs.",
		Triggers:         []string{"m?hackban", "m?hb"},
		Usage:            "m?hb [userID] <additional userIDs...>",
		Cooldown:         3,
		CooldownScope:    mio.Channel,
		RequiredPerms:    discordgo.PermissionBanMembers,
		CheckBotPerms:    true,
		RequiresUserType: mio.UserTypeAny,
		AllowedTypes:     mio.MessageTypeCreate,
		AllowDMs:         false,
		IsEnabled:        true,
		Run: func(msg *mio.DiscordMessage) {
			if len(msg.Args()) < 2 {
				return
			}

			badBans, badIDs := 0, 0
			for _, arg := range msg.Args()[1:] {
				if !utils.IsNumber(arg) {
					badIDs++
					continue
				}
				err := msg.Discord.Sess.GuildBanCreateWithReason(msg.Message.GuildID, arg, fmt.Sprintf("[%v] - Hackban", msg.Message.Author), 7)
				if err != nil {
					badBans++
					continue
				}
			}
			_, _ = msg.Reply(fmt.Sprintf("Banned %v out of %v users provided.", len(msg.Args())-1-badBans-badIDs, len(msg.Args())-1-badIDs))
		},
	}
}

func newKickCommand(m *Module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:              m,
		Name:             "kick",
		Description:      "Kicks a user. Reason is optional",
		Triggers:         []string{"m?kick", "m?k", ".kick", ".k"},
		Usage:            "m?k [user] <reason>",
		Cooldown:         2,
		CooldownScope:    mio.Channel,
		RequiredPerms:    discordgo.PermissionKickMembers,
		CheckBotPerms:    true,
		RequiresUserType: mio.UserTypeAny,
		AllowedTypes:     mio.MessageTypeCreate,
		AllowDMs:         false,
		IsEnabled:        true,
		Run:              m.kickCommand,
	}
}

func (m *Module) kickCommand(msg *mio.DiscordMessage) {
	if len(msg.Args()) < 2 {
		return
	}

	targetUser, err := msg.GetMemberAtArg(1)
	if err != nil {
		_, _ = msg.Reply("Could not find that user!")
		return
	}

	if targetUser.User.ID == msg.Sess.State.User.ID {
		_, _ = msg.Reply("no (I can not kick myself)")
		return
	}

	if targetUser.User.ID == msg.AuthorID() {
		_, _ = msg.Reply("no (you can not kick yourself)")
		return
	}

	reason := ""
	if len(msg.Args()) > 2 {
		reason = strings.Join(msg.RawArgs()[2:], " ")
	}

	topUserRole := msg.Discord.HighestRolePosition(msg.Message.GuildID, msg.Message.Author.ID)
	topTargetRole := msg.Discord.HighestRolePosition(msg.Message.GuildID, targetUser.User.ID)
	topBotRole := msg.Discord.HighestRolePosition(msg.Message.GuildID, msg.Sess.State.User.ID)

	if topUserRole <= topTargetRole || topBotRole <= topTargetRole {
		_, _ = msg.Reply("no (you can only kick users who are below you and me in the role hierarchy)")
		return
	}

	g, err := msg.Discord.Guild(msg.Message.GuildID)
	if err != nil {
		return
	}

	userCh, userChErr := msg.Sess.UserChannelCreate(targetUser.User.ID)
	if userChErr == nil {
		if reason == "" {
			_, _ = msg.Sess.ChannelMessageSend(userCh.ID, fmt.Sprintf("You have been kicked from %v.", g.Name))
		} else {
			_, _ = msg.Sess.ChannelMessageSend(userCh.ID, fmt.Sprintf("You have been kicked from %v for the following reason: %v", g.Name, reason))
		}
	}

	err = msg.Sess.GuildMemberDeleteWithReason(g.ID, targetUser.User.ID, fmt.Sprintf("%v - %v", msg.Message.Author.String(), reason))
	if err != nil {
		_, _ = msg.Reply("Something went wrong when trying to kick that user, please try again!")
		return
	}

	embed := helpers.NewEmbed().
		WithTitle("User kicked").
		WithOkColor().
		AddField("Username", targetUser.Mention(), true).
		AddField("ID", targetUser.User.ID, true)
	_, _ = msg.ReplyEmbed(embed.Build())
}

func newPruneCommand(m *Module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:              m,
		Name:             "prune",
		Description:      "Prunes all of Meido's messages in the last 100 messages. Amount of messages can be specified, but max 100. If a user is specified, it removes all messages from that user in the last 100 messages.",
		Triggers:         []string{"m?prune"},
		Usage:            "m?prune <user> <amount>",
		Cooldown:         2,
		CooldownScope:    mio.Channel,
		RequiredPerms:    discordgo.PermissionManageMessages,
		CheckBotPerms:    true,
		RequiresUserType: mio.UserTypeAny,
		AllowedTypes:     mio.MessageTypeCreate,
		AllowDMs:         false,
		IsEnabled:        true,
		Run:              m.pruneCommand,
	}
}

func (m *Module) pruneCommand(msg *mio.DiscordMessage) {
	if len(msg.Args()) == 1 {
		pruneMessages(msg, msg.Discord.BotUser().ID, 100)
	} else if len(msg.Args()) == 2 {
		if member, err := msg.GetMemberAtArg(1); err == nil {
			pruneMessages(msg, member.User.ID, 100)
			return
		}

		if !utils.IsNumber(msg.Args()[1]) {
			return
		}
		num, _ := strconv.Atoi(msg.Args()[1])
		pruneMessages(msg, "", num)
	} else if len(msg.Args()) == 3 {
		member, err := msg.GetMemberAtArg(1)
		if err != nil {
			return
		}
		if !utils.IsNumber(msg.Args()[2]) {
			return
		}
		num, _ := strconv.Atoi(msg.Args()[2])
		pruneMessages(msg, member.User.ID, num+1) // +1 because the command itself adds 1 new message
	}
}

// pruneMessages prunes messages in the DiscordMessage channel. It only prunes
// messages where the author ID corresponds to the ID given. If the ID given is
// empty, it prunes all messages. It prunes the amount of messages that the
// Session.State allows, or how many are available, or the amount given.
func pruneMessages(msg *mio.DiscordMessage, memberID string, amount int) {
	ch, err := msg.Discord.Channel(msg.ChannelID())
	if err != nil {
		return
	}
	targets := []string{}
	for i := len(ch.Messages) - 1; i >= 0; i-- {
		msg := ch.Messages[i]
		if memberID == "" || (msg.Author != nil && memberID == msg.Author.ID) {
			targets = append(targets, msg.ID)
		}
		if len(targets) == amount {
			break
		}
	}
	if len(targets) == 0 {
		return
	}
	if err = msg.Discord.Sess.ChannelMessagesBulkDelete(msg.ChannelID(), targets); err != nil {
		_, _ = msg.Reply("There was an issue, please try again!")
		return
	}
	_, _ = msg.ReplyAndDelete(fmt.Sprintf("Pruned %v messages!", len(targets)), time.Second*2)
}
