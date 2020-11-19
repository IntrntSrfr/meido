package moderationmod

import (
	"database/sql"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meidov2"
	"github.com/jmoiron/sqlx"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ModerationMod struct {
	Name string
	sync.Mutex
	cl chan *meidov2.DiscordMessage
	//commands []func(msg *meidov2.DiscordMessage)
	commands map[string]meidov2.ModCommand
	passives []func(*meidov2.DiscordMessage)
	db       *sqlx.DB
}

func New(name string) meidov2.Mod {
	return &ModerationMod{
		Name:     name,
		commands: make(map[string]meidov2.ModCommand),
	}
}

func (m *ModerationMod) Save() error {
	return nil
}

func (m *ModerationMod) Load() error {
	return nil
}

func (m *ModerationMod) Settings(msg *meidov2.DiscordMessage) {

}

func (m *ModerationMod) Help(msg *meidov2.DiscordMessage) {

}
func (m *ModerationMod) Commands() map[string]meidov2.ModCommand {
	return nil
}

func (m *ModerationMod) Hook(b *meidov2.Bot) error {
	m.cl = b.CommandLog
	m.db = b.DB

	b.Discord.Sess.AddHandler(func(s *discordgo.Session, g *discordgo.GuildCreate) {
		dbg := &DiscordGuild{}
		err := m.db.Get(dbg, "SELECT guild_id FROM guilds WHERE guild_id = $1;", g.Guild.ID)
		if err != nil && err != sql.ErrNoRows {
			fmt.Println(err)
		} else if err == sql.ErrNoRows {
			m.db.Exec("INSERT INTO guilds(guild_id, use_strikes, max_strikes) VALUES($1, $2, $3)", g.Guild.ID, false, 3)
			fmt.Println(fmt.Sprintf("Inserted new guild: %v [%v]", g.Guild.Name, g.Guild.ID))
		}
	})

	// add this later
	/*
		b.Discord.Sess.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
			refreshTicker := time.NewTicker(time.Hour)

			go func() {
				for range refreshTicker.C {
					for _, g := range b.Discord.Sess.State.Guilds {
						dge := &DiscordGuild{}
						err := b.DB.Get(dge, "SELECT * FROM guilds WHERE guild_id=$1", g.ID)
						if err != nil {
							continue
						}

						var warns []*WarnEntry
						err = b.DB.Select(&warns, "SELECT * FROM warns WHERE guild_id=$1", g.ID)
						if err != nil {
							continue
						}

						for _, warn := range warns {
							if warn.GivenAt.Unix() < time.Now().Add(time.Hour*24*30*-1).Unix() {
								b.DB.Exec("DELETE FROM warns WHERE uid=$1", warn.UID)
							}
						}
					}
				}
			}()
		})
	*/

	//m.passives = append(m.passives, m.CheckFilter)

	m.RegisterCommand(NewBanCommand(m))
	m.RegisterCommand(NewUnbanCommand(m))
	m.RegisterCommand(NewHackbanCommand(m))

	//m.RegisterCommand(NewWarnCommand(m))
	//m.RegisterCommand(NewWarnLogCommand(m))
	//m.RegisterCommand(NewRemoveWarnCommand(m))
	//m.RegisterCommand(NewClearWarnsCommand(m))

	//m.RegisterCommand(NewFilterWordCommand(m))
	//m.RegisterCommand(NewClearFilterCommand(m))
	//m.RegisterCommand(NewFilterWordListCommand(m))

	//m.RegisterCommand(NewSetMaxWarnsCommand(m))
	//m.RegisterCommand(NewToggleStrikeCommand(m))

	m.RegisterCommand(NewLockdownChannelCommand(m))
	m.RegisterCommand(NewUnlockChannelCommand(m))

	/*
		m.commands = append(m.commands, m.Unban, m.Hackban)
		m.commands = append(m.commands, m.Warn, m.WarnLog, m.RemoveWarn, m.ClearWarns)
		m.commands = append(m.commands, m.FilterWord, m.ClearFilter, m.FilterWordsList)
		m.commands = append(m.commands, m.SetMaxStrikes, m.ToggleStrikes)
	*/

	return nil
}

func (m *ModerationMod) RegisterCommand(cmd meidov2.ModCommand) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.commands[cmd.Name()]; ok {
		panic(fmt.Sprintf("command '%v' already exists in %v", cmd.Name(), m.Name))
	}
	m.commands[cmd.Name()] = cmd
}

func (m *ModerationMod) Message(msg *meidov2.DiscordMessage) {
	// moderation only is for servers, so dms are ignored
	if msg.IsDM() {
		return
	}
	if msg.Type == meidov2.MessageTypeDelete {
		return
	}
	for _, c := range m.passives {
		go c(msg)
	}

	for _, c := range m.commands {
		go c.Run(msg)
	}
}

type BanCommand struct {
	m       *ModerationMod
	Enabled bool
}

func NewBanCommand(m *ModerationMod) meidov2.ModCommand {
	return &BanCommand{
		m:       m,
		Enabled: true,
	}
}
func (c *BanCommand) Name() string {
	return "Ban"
}
func (c *BanCommand) Description() string {
	return "Bans a user. Days of messages to be deleted and reason is optional"
}
func (c *BanCommand) Triggers() []string {
	return []string{"m?ban", "m?b", ".b", ".ban"}
}
func (c *BanCommand) Usage() string {
	return ".b @internet surfer#0001\n.b 163454407999094786\n.b 163454407999094786 being very mean\n.b 163454407999094786 1 being very mean\n.b 163454407999094786 1"
}
func (c *BanCommand) Cooldown() int {
	return 10
}
func (c *BanCommand) RequiredPerms() int {
	return discordgo.PermissionBanMembers
}
func (c *BanCommand) RequiresOwner() bool {
	return false
}
func (c *BanCommand) IsEnabled() bool {
	return c.Enabled
}
func (c *BanCommand) Run(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 2 || (msg.Args()[0] != ".ban" && msg.Args()[0] != ".b" && msg.Args()[0] != "m?ban" && msg.Args()[0] != "m?b") {
		return
	}
	if msg.Type != meidov2.MessageTypeCreate {
		return
	}

	uPerms, err := msg.Discord.UserChannelPermissions(msg.Member, msg.Message.ChannelID)
	if err != nil {
		fmt.Println(err)
		return
	}
	if uPerms&discordgo.PermissionBanMembers == 0 && uPerms&discordgo.PermissionAdministrator == 0 {
		return
	}

	botPerms, err := msg.Discord.Sess.State.UserChannelPermissions(msg.Sess.State.User.ID, msg.Message.ChannelID)
	if err != nil {
		return
	}
	if botPerms&discordgo.PermissionBanMembers == 0 && botPerms&discordgo.PermissionAdministrator == 0 {
		return
	}

	c.m.cl <- msg

	var (
		targetUser *discordgo.User
		reason     string
		pruneDays  int
	)

	switch la := msg.LenArgs(); {
	case la == 2:
		pruneDays = 0
		reason = ""
	case la >= 3:
		pruneDays, err = strconv.Atoi(msg.Args()[2])
		if err != nil {
			pruneDays = 0
			reason = strings.Join(msg.Args()[2:], " ")
		} else {
			reason = strings.Join(msg.Args()[3:], " ")
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
		msg.Reply("no")
		return
	}
	if targetUser.ID == msg.Message.Author.ID {
		msg.Reply("no")
		return
	}

	topUserRole := msg.HighestRole(msg.Message.GuildID, msg.Message.Author.ID)
	topTargetRole := msg.HighestRole(msg.Message.GuildID, targetUser.ID)
	topBotRole := msg.HighestRole(msg.Message.GuildID, msg.Sess.State.User.ID)

	if topUserRole <= topTargetRole || topBotRole <= topTargetRole {
		msg.Reply("no")
		return
	}

	if topTargetRole > 0 {
		userChannel, userChErr := msg.Discord.Sess.UserChannelCreate(targetUser.ID)
		if userChErr == nil {
			g, err := msg.Discord.Sess.State.Guild(msg.Message.GuildID)
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

	_, err = c.m.db.Exec("UPDATE warns SET is_valid=false, cleared_by_id=$1, cleared_at=$2 WHERE guild_id=$3 AND user_id=$4 and is_valid",
		msg.Sess.State.User.ID, time.Now(), msg.Message.GuildID, targetUser.ID)

	embed := &discordgo.MessageEmbed{
		Title: "User banned",
		Color: 0xC80000,
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

type UnbanCommand struct {
	m       *ModerationMod
	Enabled bool
}

func NewUnbanCommand(m *ModerationMod) meidov2.ModCommand {
	return &UnbanCommand{
		m:       m,
		Enabled: true,
	}
}

func (c *UnbanCommand) Name() string {
	return "Unban"
}

func (c *UnbanCommand) Description() string {
	return "Unbans a user."
}

func (c *UnbanCommand) Triggers() []string {
	return []string{"m?unban", "m?ub", ".ub", ".unban"}
}

func (c *UnbanCommand) Usage() string {
	return ".unban 163454407999094786"
}

func (c *UnbanCommand) Cooldown() int {
	return 10
}

func (c *UnbanCommand) RequiredPerms() int {
	return discordgo.PermissionBanMembers
}

func (c *UnbanCommand) RequiresOwner() bool {
	return false
}

func (c *UnbanCommand) IsEnabled() bool {
	return c.Enabled
}

func (c *UnbanCommand) Run(msg *meidov2.DiscordMessage) {

	if msg.LenArgs() < 2 || (msg.Args()[0] != ".unban" && msg.Args()[0] != ".ub" && msg.Args()[0] != "m?unban" && msg.Args()[0] != "m?ub") {
		return
	}
	if msg.Type != meidov2.MessageTypeCreate {
		return
	}

	uPerms, err := msg.Discord.UserChannelPermissions(msg.Member, msg.Message.ChannelID)
	if err != nil {
		fmt.Println(err)
		return
	}
	if uPerms&discordgo.PermissionBanMembers == 0 && uPerms&discordgo.PermissionAdministrator == 0 {
		return
	}

	botPerms, err := msg.Discord.Sess.State.UserChannelPermissions(msg.Sess.State.User.ID, msg.Message.ChannelID)
	if err != nil {
		fmt.Println(err)
		return
	}
	if botPerms&discordgo.PermissionBanMembers == 0 && botPerms&discordgo.PermissionAdministrator == 0 {
		return
	}

	c.m.cl <- msg

	_, err = strconv.ParseUint(msg.Args()[1], 10, 64)
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
		Color:       0x00C800,
	}

	msg.ReplyEmbed(embed)
}

type HackbanCommand struct {
	m       *ModerationMod
	Enabled bool
}

func NewHackbanCommand(m *ModerationMod) meidov2.ModCommand {
	return &HackbanCommand{
		m:       m,
		Enabled: true,
	}
}

func (c *HackbanCommand) Name() string {
	return "Hackban"
}

func (c *HackbanCommand) Description() string {
	return "Hackbans one or several users. Prunes 7 days."
}

func (c *HackbanCommand) Triggers() []string {
	return []string{"m?hackban", "m?hb"}
}

func (c *HackbanCommand) Usage() string {
	return "m?hb 123 123 12 31 23 123"
}

func (c *HackbanCommand) Cooldown() int {
	return 10
}

func (c *HackbanCommand) RequiredPerms() int {
	return discordgo.PermissionBanMembers
}

func (c *HackbanCommand) RequiresOwner() bool {
	return false
}

func (c *HackbanCommand) IsEnabled() bool {
	return c.Enabled
}

func (c *HackbanCommand) Run(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 2 || (msg.Args()[0] != "m?hackban" && msg.Args()[0] != "m?hb") {
		return
	}
	if msg.Type != meidov2.MessageTypeCreate {
		return
	}

	uPerms, err := msg.Discord.UserChannelPermissions(msg.Member, msg.Message.ChannelID)
	if err != nil {
		fmt.Println(err)
		return
	}
	if uPerms&discordgo.PermissionBanMembers == 0 && uPerms&discordgo.PermissionAdministrator == 0 {
		return
	}

	botPerms, err := msg.Discord.Sess.State.UserChannelPermissions(msg.Sess.State.User.ID, msg.Message.ChannelID)
	if err != nil {
		fmt.Println(err)
		return
	}
	if botPerms&discordgo.PermissionBanMembers == 0 && botPerms&discordgo.PermissionAdministrator == 0 {
		return
	}

	c.m.cl <- msg

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
		/*
			err = msg.Discord.Client.BanMember(context.Background(), msg.Message.GuildID, disgord.Snowflake(userID), &disgord.BanMemberParams{
				DeleteMessageDays: 7,
				Reason:            fmt.Sprintf("[%v] - Hackban", msg.Message.Author.Tag()),
			})
		*/
		if err != nil {
			fmt.Println(err)
			badBans++
			continue
		}
	}
	msg.Reply(fmt.Sprintf("Banned %v out of %v users provided.", len(userList)-badBans-badIDs, len(userList)-badIDs))
}

type KickCommand struct {
	m       *ModerationMod
	Enabled bool
}

func (c *KickCommand) Name() string {
	return "Kick"
}

func (c *KickCommand) Description() string {
	return "Kicks a user. Reason is optional"
}

func (c *KickCommand) Triggers() []string {
	return []string{"m?kick", "m?k", ".kick", ".k"}
}

func (c *KickCommand) Usage() string {
	return "m?k @internet surfer#0001\n.k 163454407999094786"
}

func (c *KickCommand) Cooldown() int {
	return 10
}

func (c *KickCommand) RequiredPerms() int {
	return discordgo.PermissionKickMembers
}

func (c *KickCommand) RequiresOwner() bool {
	return false
}

func (c *KickCommand) IsEnabled() bool {
	return c.Enabled
}

func (c *KickCommand) Run(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 2 || (msg.Args()[0] != ".kick" && msg.Args()[0] != ".k" && msg.Args()[0] != "m?kick" && msg.Args()[0] != "m?k") {
		return
	}
	if msg.Type != meidov2.MessageTypeCreate {
		return
	}

	uPerms, err := msg.Discord.UserChannelPermissions(msg.Member, msg.Message.ChannelID)
	if err != nil {
		fmt.Println(err)
		return
	}
	if uPerms&discordgo.PermissionKickMembers == 0 && uPerms&discordgo.PermissionAdministrator == 0 {
		return
	}

	botPerms, err := msg.Discord.Sess.State.UserChannelPermissions(msg.Sess.State.User.ID, msg.Message.ChannelID)
	if err != nil {
		fmt.Println(err)
		return
	}
	if botPerms&discordgo.PermissionKickMembers == 0 && botPerms&discordgo.PermissionAdministrator == 0 {
		return
	}

	c.m.cl <- msg

	var targetUser *discordgo.Member

	reason := ""
	if msg.LenArgs() > 2 {
		reason = strings.Join(msg.Args()[2:], " ")
	}

	if len(msg.Message.Mentions) >= 1 {
		targetUser, err = msg.Sess.GuildMember(msg.Message.GuildID, msg.Message.Mentions[0].ID)
		if err != nil {
			msg.Reply("that person isnt even here wtf :(")
			return
		}
	} else {
		targetUser, err = msg.Sess.GuildMember(msg.Message.GuildID, msg.Args()[1])
		if err != nil {
			msg.Reply("that person isnt even here wtf :(")
			return
		}
	}

	if targetUser.User.ID == msg.Sess.State.User.ID {
		msg.Reply("no")
		return
	}
	if targetUser.User.ID == msg.Message.Author.ID {
		msg.Reply("no")
		return
	}

	topUserRole := msg.HighestRole(msg.Message.GuildID, msg.Message.Author.ID)
	topTargetRole := msg.HighestRole(msg.Message.GuildID, targetUser.User.ID)
	topBotRole := msg.HighestRole(msg.Message.GuildID, msg.Sess.State.User.ID)

	if topUserRole <= topTargetRole || topBotRole <= topTargetRole {
		msg.Reply("no")
		return
	}

	g, err := msg.Sess.State.Guild(msg.Message.GuildID)
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
		msg.Reply("failed to kick user")
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
		Color: 0xC80000,
	}

	msg.ReplyEmbed(embed)
}
