package moderationmod

import (
	"database/sql"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/internal/database"
	"github.com/intrntsrfr/meido/pkg/mio"
	"github.com/intrntsrfr/meido/pkg/utils"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"time"
)

type ModerationMod struct {
	*mio.ModuleBase
	bot *mio.Bot
	db  database.DB
	log *zap.Logger
}

func New(b *mio.Bot, db *database.PsqlDB, log *zap.Logger) mio.Module {
	return &ModerationMod{
		ModuleBase: mio.NewModule("Moderation"),
		bot:        b,
		db:         db,
		log:        log,
	}
}

func (m *ModerationMod) Hook() error {
	m.bot.Discord.Sess.AddHandler(func(s *discordgo.Session, g *discordgo.GuildCreate) {
		dbg := &database.Guild{}
		err := m.db.Get(dbg, "SELECT guild_id FROM guild WHERE guild_id = $1;", g.Guild.ID)
		if err != nil && err != sql.ErrNoRows {
			fmt.Println(err)
		} else if err == sql.ErrNoRows {
			m.db.Exec("INSERT INTO guild VALUES($1)", g.Guild.ID)
			fmt.Println(fmt.Sprintf("Inserted new guild: %v [%v]", g.Guild.Name, g.Guild.ID))
		}
	})

	m.bot.Discord.Sess.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		refreshTicker := time.NewTicker(time.Hour)

		go func() {
			for range refreshTicker.C {
				for _, g := range m.bot.Discord.Guilds() {
					if g.Unavailable {
						continue
					}
					dge := &database.Guild{}
					err := m.db.Get(dge, "SELECT * FROM guild WHERE guild_id=$1", g.ID)
					if err != nil {
						continue
					}

					if dge.WarnDuration <= 0 {
						continue
					}

					var warns []*database.Warn
					err = m.db.Select(&warns, "SELECT * FROM warn WHERE guild_id=$1 AND is_valid", g.ID)
					if err != nil {
						continue
					}

					dur := time.Duration(dge.WarnDuration)
					for _, warn := range warns {
						if warn.GivenAt.Unix() < time.Now().Add(-1*time.Hour*24*dur).Unix() {
							m.db.Exec("UPDATE warn SET is_valid=false, cleared_by_id=$1, cleared_at=$2 WHERE uid=$3",
								m.bot.Discord.Sess.State.User.ID, time.Now(), warn.UID)
						}
					}
				}
			}
		}()
	})
	/*
		b.Discord.Sess.AddHandler(func(s *discordgo.Session, g *discordgo.GuildMemberAdd) {

			a := &database.AutoRole{}
			err := m.db.Get(a, "SELECT * FROM guild WHERE guild_id=$1", g.GuildID)
			if err != nil {
				return
			}

			if a.RoleID == "" {
				return
			}

			guild, err := b.Discord.Guild(g.GuildID)
			if err != nil {
				fmt.Println(err)
				return
			}

			found := false
			for _, r := range guild.Roles {
				if r.UID == a.RoleID {
					found = true
				}
			}

			if !found {
				// if its not found, the role should probably get set to be empty
				return
			}

			s.GuildMemberRoleAdd(g.GuildID, g.User.UID, a.RoleID)
		})
	*/
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
		//NewAutoRoleCommand(m),
	})
}

func NewBanCommand(m *ModerationMod) *mio.ModuleCommand {
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

func (m *ModerationMod) banCommand(msg *mio.DiscordMessage) {
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
		msg.Reply("I could not ban that user!")
		return
	}

	// clear all active warns
	err = m.db.ClearActiveUserWarns(msg.GuildID(), targetUser.ID, msg.AuthorID())
	if err != nil {
		m.log.Error("could not clear user warns", zap.Error(err))
	}

	//_, err = m.db.Exec("UPDATE warn SET is_valid=false, cleared_by_id=$1, cleared_at=$2 WHERE guild_id=$3 AND user_id=$4 and is_valid",
	//	msg.Sess.State.User.UID, time.Now(), msg.Message.GuildID, targetUser.UID)

	embed := &discordgo.MessageEmbed{
		Title: "User banned",
		Color: utils.ColorCritical,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Username",
				Value:  fmt.Sprintf("%v", targetUser.Mention()),
				Inline: true,
			},
			{
				Name:   "UID",
				Value:  fmt.Sprintf("%v", targetUser.ID),
				Inline: true,
			},
		},
	}

	_, _ = msg.ReplyEmbed(embed)
}

func NewUnbanCommand(m *ModerationMod) *mio.ModuleCommand {
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

func (m *ModerationMod) unbanCommand(msg *mio.DiscordMessage) {
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

	embed := &discordgo.MessageEmbed{
		Description: fmt.Sprintf("**Unbanned** %v - %v#%v (%v)", targetUser.Mention(), targetUser.Username, targetUser.Discriminator, targetUser.ID),
		Color:       utils.ColorGreen,
	}

	_, _ = msg.ReplyEmbed(embed)
}

func NewHackbanCommand(m *ModerationMod) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:           m,
		Name:          "hackban",
		Description:   "Hackbans one or several users. Prunes 7 days.",
		Triggers:      []string{"m?hackban", "m?hb"},
		Usage:         "m?hb [userID] <userID>...",
		Cooldown:      3,
		RequiredPerms: discordgo.PermissionBanMembers,
		RequiresOwner: false,
		CheckBotPerms: true,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           m.hackbanCommand,
	}
}

func (m *ModerationMod) hackbanCommand(msg *mio.DiscordMessage) {
	if msg.LenArgs() < 2 {
		return
	}

	var userList []string
	for _, mention := range msg.Message.Mentions {
		userList = append(userList, fmt.Sprint(mention.ID))
	}

	for _, userID := range msg.Args()[1:] {
		userList = append(userList, userID)
	}

	badBans, badIDs := 0, 0
	for _, userIDString := range userList {
		_, err := strconv.Atoi(userIDString)
		if err != nil {
			badIDs++
			continue
		}
		err = msg.Discord.Sess.GuildBanCreateWithReason(msg.Message.GuildID, userIDString, fmt.Sprintf("[%v] - Hackban", msg.Message.Author.String()), 7)
		if err != nil {
			fmt.Println(err)
			badBans++
			continue
		}
	}
	_, _ = msg.Reply(fmt.Sprintf("Banned %v out of %v users provided.", len(userList)-badBans-badIDs, len(userList)-badIDs))
}

func NewKickCommand(m *ModerationMod) *mio.ModuleCommand {
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

func (m *ModerationMod) kickCommand(msg *mio.DiscordMessage) {
	if msg.LenArgs() < 2 {
		return
	}

	targetUser, err := msg.GetMemberAtArg(1)
	if err != nil {
		_, _ = msg.Reply("Could not find that user!")
		return
	}

	if targetUser.User.ID == msg.Sess.State.User.ID {
		_, _ = msg.Reply("no (i can not kick myself)")
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

	embed := &discordgo.MessageEmbed{
		Title: "User kicked",
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Username",
				Value:  fmt.Sprintf("%v", targetUser.Mention()),
				Inline: true,
			},
			{
				Name:   "UID",
				Value:  fmt.Sprintf("%v", targetUser.User.ID),
				Inline: true,
			},
		},
		Color: utils.ColorCritical,
	}

	_, _ = msg.ReplyEmbed(embed)
}
