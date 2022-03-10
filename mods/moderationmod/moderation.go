package moderationmod

import (
	"database/sql"
	"fmt"
	"github.com/bwmarrin/discordgo"
	base2 "github.com/intrntsrfr/meido/base"
	database2 "github.com/intrntsrfr/meido/database"
	"github.com/intrntsrfr/meido/utils"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ModerationMod struct {
	sync.Mutex
	name         string
	commands     map[string]*base2.ModCommand
	passives     []*base2.ModPassive
	db           *database2.DB
	allowedTypes base2.MessageType
	allowDMs     bool
	bot          *base2.Bot
}

func New(name string) base2.Mod {
	return &ModerationMod{
		name:         name,
		commands:     make(map[string]*base2.ModCommand),
		allowedTypes: base2.MessageTypeCreate | base2.MessageTypeUpdate,
		allowDMs:     false,
	}
}

func (m *ModerationMod) Name() string {
	return m.name
}
func (m *ModerationMod) Save() error {
	return nil
}
func (m *ModerationMod) Load() error {
	return nil
}
func (m *ModerationMod) Passives() []*base2.ModPassive {
	return m.passives
}
func (m *ModerationMod) Commands() map[string]*base2.ModCommand {
	return m.commands
}
func (m *ModerationMod) AllowedTypes() base2.MessageType {
	return m.allowedTypes
}
func (m *ModerationMod) AllowDMs() bool {
	return m.allowDMs
}
func (m *ModerationMod) Hook(b *base2.Bot) error {
	m.bot = b
	m.db = b.DB

	b.Discord.Sess.AddHandler(func(s *discordgo.Session, g *discordgo.GuildCreate) {
		dbg := &database2.Guild{}
		err := m.db.Get(dbg, "SELECT guild_id FROM guild WHERE guild_id = $1;", g.Guild.ID)
		if err != nil && err != sql.ErrNoRows {
			fmt.Println(err)
		} else if err == sql.ErrNoRows {
			m.db.Exec("INSERT INTO guild(guild_id, use_warns, max_warns) VALUES($1, $2, $3)", g.Guild.ID, false, 3)
			fmt.Println(fmt.Sprintf("Inserted new guild: %v [%v]", g.Guild.Name, g.Guild.ID))
		}
	})

	b.Discord.Sess.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		refreshTicker := time.NewTicker(time.Hour)

		go func() {
			for range refreshTicker.C {
				for _, g := range b.Discord.Guilds() {
					if g.Unavailable {
						continue
					}
					dge := &database2.Guild{}
					err := b.DB.Get(dge, "SELECT * FROM guild WHERE guild_id=$1", g.ID)
					if err != nil {
						continue
					}

					if dge.WarnDuration <= 0 {
						continue
					}

					var warns []*database2.Warn
					err = b.DB.Select(&warns, "SELECT * FROM warn WHERE guild_id=$1 AND is_valid", g.ID)
					if err != nil {
						continue
					}

					dur := time.Duration(dge.WarnDuration)
					for _, warn := range warns {
						if warn.GivenAt.Unix() < time.Now().Add(-1*time.Hour*24*dur).Unix() {
							b.DB.Exec("UPDATE warn SET is_valid=false, cleared_by_id=$1, cleared_at=$2 WHERE id=$3",
								b.Discord.Sess.State.User.ID, time.Now(), warn.UID)
						}
					}
				}
			}
		}()
	})

	b.Discord.Sess.AddHandler(func(s *discordgo.Session, g *discordgo.GuildMemberAdd) {

		a := &database2.AutoRole{}
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
			if r.ID == a.RoleID {
				found = true
			}
		}

		if !found {
			// if its not found, the role should probably get set to be empty
			return
		}

		s.GuildMemberRoleAdd(g.GuildID, g.User.ID, a.RoleID)
	})

	m.passives = append(m.passives, NewCheckFilterPassive(m))

	m.RegisterCommand(NewBanCommand(m))
	m.RegisterCommand(NewUnbanCommand(m))
	m.RegisterCommand(NewHackbanCommand(m))
	m.RegisterCommand(NewKickCommand(m))

	m.RegisterCommand(NewWarnCommand(m))
	m.RegisterCommand(NewClearWarnCommand(m))
	m.RegisterCommand(NewWarnLogCommand(m))
	m.RegisterCommand(NewClearAllWarnsCommand(m))
	m.RegisterCommand(NewWarnCountCommand(m))

	m.RegisterCommand(NewFilterWordCommand(m))
	m.RegisterCommand(NewClearFilterCommand(m))
	m.RegisterCommand(NewFilterWordListCommand(m))

	m.RegisterCommand(NewModerationSettingsCommand(m))

	m.RegisterCommand(NewLockdownChannelCommand(m))
	m.RegisterCommand(NewUnlockChannelCommand(m))

	//m.RegisterCommand(NewAutoRoleCommand(m))

	return nil
}

func (m *ModerationMod) RegisterCommand(cmd *base2.ModCommand) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.commands[cmd.Name]; ok {
		panic(fmt.Sprintf("command '%v' already exists in %v", cmd.Name, m.Name()))
	}
	m.commands[cmd.Name] = cmd
}

func NewBanCommand(m *ModerationMod) *base2.ModCommand {
	return &base2.ModCommand{
		Mod:           m,
		Name:          "ban",
		Description:   "Bans a user. Days of messages to be deleted and reason is optional",
		Triggers:      []string{"m?ban", "m?b", ".b", ".ban"},
		Usage:         ".b @internet surfer#0001\n.b 163454407999094786\n.b 163454407999094786 being very mean\n.b 163454407999094786 1 being very mean\n.b 163454407999094786 1",
		Cooldown:      2,
		RequiredPerms: discordgo.PermissionBanMembers,
		RequiresOwner: false,
		CheckBotPerms: true,
		AllowedTypes:  base2.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           m.banCommand,
	}
}

func (m *ModerationMod) banCommand(msg *base2.DiscordMessage) {
	if msg.LenArgs() < 2 {
		return
	}

	var (
		targetUser *discordgo.User
		reason     string
		pruneDays  int
		err        error
	)

	switch la := msg.LenArgs(); {
	case la == 2:
		pruneDays = 0
		reason = ""
	case la >= 3:
		pruneDays, err = strconv.Atoi(msg.Args()[2])
		if err != nil {
			pruneDays = 0
			reason = strings.Join(msg.RawArgs()[2:], " ")
		} else {
			reason = strings.Join(msg.RawArgs()[3:], " ")
		}

		//pruneDays = int(math.Max(float64(0), float64(pruneDays)))
		if pruneDays > 7 {
			pruneDays = 7
		} else if pruneDays < 0 {
			pruneDays = 0
		}
	}

	if len(msg.Message.Mentions) > 0 {
		targetUser = msg.Message.Mentions[0]
	} else {
		targetUser, err = msg.Discord.Sess.User(msg.Args()[1])
		if err != nil {
			return
		}
	}

	if targetUser.ID == msg.Sess.State.User.ID {
		msg.Reply("no (i can not ban myself)")
		return
	}
	if targetUser.ID == msg.Message.Author.ID {
		msg.Reply("no (you can not ban yourself)")
		return
	}

	topUserRole := msg.Discord.HighestRolePosition(msg.Message.GuildID, msg.Message.Author.ID)
	topTargetRole := msg.Discord.HighestRolePosition(msg.Message.GuildID, targetUser.ID)
	topBotRole := msg.Discord.HighestRolePosition(msg.Message.GuildID, msg.Sess.State.User.ID)

	if topUserRole <= topTargetRole || topBotRole <= topTargetRole {
		msg.Reply("no (you can only ban users who are below you and me in the role hierarchy)")
		return
	}

	if topTargetRole > 0 {
		userChannel, userChErr := msg.Discord.Sess.UserChannelCreate(targetUser.ID)
		if userChErr == nil {
			g, err := msg.Discord.Guild(msg.Message.GuildID)
			if err != nil {
				return
			}

			if reason == "" {
				msg.Sess.ChannelMessageSend(userChannel.ID, fmt.Sprintf("You have been banned from %v", g.Name))
			} else {
				msg.Sess.ChannelMessageSend(userChannel.ID, fmt.Sprintf("You have been banned from %v for the following reason:\n%v", g.Name, reason))
			}
		}
	}

	err = msg.Discord.Sess.GuildBanCreateWithReason(msg.Message.GuildID, targetUser.ID, fmt.Sprintf("%v - %v", msg.Message.Author.String(), reason), pruneDays)
	if err != nil {
		msg.Reply(err.Error())
		return
	}

	// clear all active warns
	_, err = m.db.Exec("UPDATE warn SET is_valid=false, cleared_by_id=$1, cleared_at=$2 WHERE guild_id=$3 AND user_id=$4 and is_valid",
		msg.Sess.State.User.ID, time.Now(), msg.Message.GuildID, targetUser.ID)

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
				Name:   "ID",
				Value:  fmt.Sprintf("%v", targetUser.ID),
				Inline: true,
			},
		},
	}

	msg.ReplyEmbed(embed)
}

func NewUnbanCommand(m *ModerationMod) *base2.ModCommand {
	return &base2.ModCommand{
		Mod:           m,
		Name:          "unban",
		Description:   "Unbans a user",
		Triggers:      []string{"m?unban", "m?ub", ".ub", ".unban"},
		Usage:         ".unban 163454407999094786",
		Cooldown:      2,
		RequiredPerms: discordgo.PermissionBanMembers,
		RequiresOwner: false,
		CheckBotPerms: true,
		AllowedTypes:  base2.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           m.unbanCommand,
	}
}

func (m *ModerationMod) unbanCommand(msg *base2.DiscordMessage) {

	if msg.LenArgs() < 2 {
		return
	}

	_, err := strconv.ParseUint(msg.Args()[1], 10, 64)
	if err != nil {
		return
	}

	err = msg.Discord.Sess.GuildBanDelete(msg.Message.GuildID, msg.Args()[1])
	if err != nil {
		return
	}

	targetUser, err := msg.Discord.Sess.User(msg.Args()[1])
	if err != nil {
		return
	}

	embed := &discordgo.MessageEmbed{
		Description: fmt.Sprintf("**Unbanned** %v - %v#%v (%v)", targetUser.Mention(), targetUser.Username, targetUser.Discriminator, targetUser.ID),
		Color:       utils.ColorGreen,
	}

	msg.ReplyEmbed(embed)
}

func NewHackbanCommand(m *ModerationMod) *base2.ModCommand {
	return &base2.ModCommand{
		Mod:           m,
		Name:          "hackban",
		Description:   "Hackbans one or several users. Prunes 7 days.",
		Triggers:      []string{"m?hackban", "m?hb"},
		Usage:         "m?hb 123 123 12 31 23 123",
		Cooldown:      3,
		RequiredPerms: discordgo.PermissionBanMembers,
		RequiresOwner: false,
		CheckBotPerms: true,
		AllowedTypes:  base2.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           m.hackbanCommand,
	}
}

func (m *ModerationMod) hackbanCommand(msg *base2.DiscordMessage) {
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

	badBans := 0
	badIDs := 0

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
	msg.Reply(fmt.Sprintf("Banned %v out of %v users provided.", len(userList)-badBans-badIDs, len(userList)-badIDs))
}

func NewKickCommand(m *ModerationMod) *base2.ModCommand {
	return &base2.ModCommand{
		Mod:           m,
		Name:          "kick",
		Description:   "Kicks a user. Reason is optional",
		Triggers:      []string{"m?kick", "m?k", ".kick", ".k"},
		Usage:         "m?k @internet surfer#0001\n.k 163454407999094786",
		Cooldown:      2,
		RequiredPerms: discordgo.PermissionKickMembers,
		RequiresOwner: false,
		CheckBotPerms: true,
		AllowedTypes:  base2.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           m.kickCommand,
	}
}

func (m *ModerationMod) kickCommand(msg *base2.DiscordMessage) {
	if msg.LenArgs() < 2 {
		return
	}

	var (
		err        error
		targetUser *discordgo.Member
	)

	reason := ""
	if msg.LenArgs() > 2 {
		reason = strings.Join(msg.RawArgs()[2:], " ")
	}

	if len(msg.Message.Mentions) >= 1 {
		targetUser, err = msg.Discord.Member(msg.Message.GuildID, msg.Message.Mentions[0].ID)
		if err != nil {
			msg.Reply("i could not find that member")
			return
		}
	} else {
		targetUser, err = msg.Discord.Member(msg.Message.GuildID, msg.Args()[1])
		if err != nil {
			msg.Reply("i could not find that member")
			return
		}
	}

	if targetUser.User.ID == msg.Sess.State.User.ID {
		msg.Reply("no (i can not ban myself)")
		return
	}
	if targetUser.User.ID == msg.Message.Author.ID {
		msg.Reply("no (you can not ban yourself)")
		return
	}

	topUserRole := msg.Discord.HighestRolePosition(msg.Message.GuildID, msg.Message.Author.ID)
	topTargetRole := msg.Discord.HighestRolePosition(msg.Message.GuildID, targetUser.User.ID)
	topBotRole := msg.Discord.HighestRolePosition(msg.Message.GuildID, msg.Sess.State.User.ID)

	if topUserRole <= topTargetRole || topBotRole <= topTargetRole {
		msg.Reply("no (you can only kick users who are below you and me in the role hierarchy)")
		return
	}

	g, err := msg.Discord.Guild(msg.Message.GuildID)
	if err != nil {
		return
	}

	userCh, userChErr := msg.Sess.UserChannelCreate(targetUser.User.ID)

	if userChErr == nil {
		if reason == "" {
			msg.Sess.ChannelMessageSend(userCh.ID, fmt.Sprintf("You have been kicked from %v.", g.Name))
		} else {
			msg.Sess.ChannelMessageSend(userCh.ID, fmt.Sprintf("You have been kicked from %v for the following reason: %v", g.Name, reason))
		}
	}

	err = msg.Sess.GuildMemberDeleteWithReason(g.ID, targetUser.User.ID, fmt.Sprintf("%v - %v", msg.Message.Author.String(), reason))
	if err != nil {
		msg.Reply("something went wrong when trying to kick user, please try again")
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
				Name:   "ID",
				Value:  fmt.Sprintf("%v", targetUser.User.ID),
				Inline: true,
			},
		},
		Color: utils.ColorCritical,
	}

	msg.ReplyEmbed(embed)
}
