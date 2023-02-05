package moderation

import (
	"database/sql"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/internal/database"
	"github.com/intrntsrfr/meido/internal/helpers"
	"github.com/intrntsrfr/meido/pkg/mio"
	"github.com/intrntsrfr/meido/pkg/utils"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"time"
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
	m.Bot.Discord.Sess.AddHandler(func(s *discordgo.Session, g *discordgo.GuildCreate) {
		if _, err := m.db.GetGuild(g.Guild.ID); err != nil && err == sql.ErrNoRows {
			if err = m.db.CreateGuild(g.Guild.ID); err != nil {
				m.Log.Error("could not create new guild", zap.Error(err), zap.String("guild ID", g.ID))
			}
		}
	})

	m.Bot.Discord.Sess.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		refreshTicker := time.NewTicker(time.Hour)
		go func() {
			for range refreshTicker.C {
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
	})

	m.Bot.Discord.AddEventHandler(addAutoRoleOnJoin(m))

	if err := m.RegisterPassive(NewCheckFilterPassive(m)); err != nil {
		return err
	}

	return m.RegisterCommands([]*mio.ModuleCommand{
		NewBanCommand(m),
		NewUnbanCommand(m),
		NewHackbanCommand(m),
		NewKickCommand(m),
		NewWarnCommand(m),
		NewClearWarnCommand(m),
		NewWarnLogCommand(m),
		NewClearAllWarnsCommand(m),
		NewWarnCountCommand(m),
		NewFilterWordCommand(m),
		NewClearFilterCommand(m),
		NewFilterWordListCommand(m),
		NewModerationSettingsCommand(m),
		NewLockdownChannelCommand(m),
		NewUnlockChannelCommand(m),
		NewMuteCommand(m),
		NewUnmuteCommand(m),
		newSetAutoRoleCommand(m),
		newRemoveAutoRoleCommand(m),
	})
}

func NewBanCommand(m *Module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:           m,
		Name:          "ban",
		Description:   "Bans a user. Days of messages to be deleted and reason is optional",
		Triggers:      []string{"m?ban", "m?b", ".b", ".ban"},
		Usage:         ".b @internet surfer#0001\n.b 163454407999094786\n.b 163454407999094786 being very mean\n.b 163454407999094786 1 being very mean\n.b 163454407999094786 1",
		Cooldown:      2,
		RequiredPerms: discordgo.PermissionBanMembers,
		RequiresOwner: false,
		CheckBotPerms: true,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           m.banCommand,
	}
}

func (m *Module) banCommand(msg *mio.DiscordMessage) {
	if msg.LenArgs() < 2 {
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

	if msg.LenArgs() > 2 {
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
		_, _ = msg.Reply("There was an issue, please try again")
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

	//_, err = m.db.Exec("UPDATE warn SET is_valid=false, cleared_by_id=$1, cleared_at=$2 WHERE guild_id=$3 AND user_id=$4 and is_valid",
	//	msg.Sess.State.User.UID, time.Now(), msg.Message.GuildID, targetUser.UID)

	embed := helpers.NewEmbed().
		WithTitle("User banned").
		WithOkColor().
		AddField("Username", targetUser.Mention(), true).
		AddField("ID", targetUser.ID, true)
	_, _ = msg.ReplyEmbed(embed.Build())
}

func NewUnbanCommand(m *Module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:           m,
		Name:          "unban",
		Description:   "Unbans a user",
		Triggers:      []string{"m?unban", "m?ub", ".ub", ".unban"},
		Usage:         ".unban 163454407999094786",
		Cooldown:      2,
		RequiredPerms: discordgo.PermissionBanMembers,
		RequiresOwner: false,
		CheckBotPerms: true,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           m.unbanCommand,
	}
}

func (m *Module) unbanCommand(msg *mio.DiscordMessage) {
	if msg.LenArgs() < 2 {
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

func NewHackbanCommand(m *Module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:           m,
		Name:          "hackban",
		Description:   "Hackbans one or several users. Prunes 7 days. Only accepts user IDs.",
		Triggers:      []string{"m?hackban", "m?hb"},
		Usage:         "m?hb [userID] <userID>...",
		Cooldown:      3,
		RequiredPerms: discordgo.PermissionBanMembers,
		RequiresOwner: false,
		CheckBotPerms: true,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run: func(msg *mio.DiscordMessage) {
			if msg.LenArgs() < 2 {
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
			_, _ = msg.Reply(fmt.Sprintf("Banned %v out of %v users provided.", msg.LenArgs()-1-badBans-badIDs, msg.LenArgs()-1-badIDs))
		},
	}
}

func NewKickCommand(m *Module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:           m,
		Name:          "kick",
		Description:   "Kicks a user. Reason is optional",
		Triggers:      []string{"m?kick", "m?k", ".kick", ".k"},
		Usage:         "m?k @internet surfer#0001\n.k 163454407999094786",
		Cooldown:      2,
		RequiredPerms: discordgo.PermissionKickMembers,
		RequiresOwner: false,
		CheckBotPerms: true,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           m.kickCommand,
	}
}

func (m *Module) kickCommand(msg *mio.DiscordMessage) {
	if msg.LenArgs() < 2 {
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
	if msg.LenArgs() > 2 {
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
