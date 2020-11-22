package moderationmod

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meidov2"
	"strconv"
	"strings"
	"time"
)

type FilterWordCommand struct {
	m       *ModerationMod
	Enabled bool
}

func NewFilterWordCommand(m *ModerationMod) meidov2.ModCommand {
	return &FilterWordCommand{
		m:       m,
		Enabled: true,
	}
}
func (c *FilterWordCommand) Name() string {
	return "filterword"
}
func (c *FilterWordCommand) Description() string {
	return "Adds or removes a word or phrase to the server filter."
}
func (c *FilterWordCommand) Triggers() []string {
	return []string{"m?fw", "m?filterword"}
}
func (c *FilterWordCommand) Usage() string {
	return "m?fw jeff"
}
func (c *FilterWordCommand) Cooldown() int {
	return 3
}
func (c *FilterWordCommand) RequiredPerms() int {
	return discordgo.PermissionManageMessages
}
func (c *FilterWordCommand) RequiresOwner() bool {
	return false
}
func (c *FilterWordCommand) IsEnabled() bool {
	return c.Enabled
}
func (c *FilterWordCommand) Run(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 2 || (msg.Args()[0] != "m?filterword" && msg.Args()[0] != "m?fw") {
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
	if uPerms&discordgo.PermissionManageMessages == 0 && uPerms&discordgo.PermissionAdministrator == 0 {
		return
	}

	c.m.cl <- msg

	phrase := strings.Join(msg.Args()[1:], " ")

	fe := &FilterEntry{}

	err = c.m.db.Get(fe, "SELECT phrase FROM filters WHERE phrase = $1 AND guild_id = $2;", phrase, msg.Message.GuildID)
	switch err {
	case nil:
		c.m.db.Exec("DELETE FROM filters WHERE guild_id=$1 AND phrase=$2;", msg.Message.GuildID, phrase)
		msg.Reply(fmt.Sprintf("Removed `%v` from the filter.", phrase))
	case sql.ErrNoRows:
		c.m.db.Exec("INSERT INTO filters (guild_id, phrase) VALUES ($1,$2);", msg.Message.GuildID, phrase)
		msg.Reply(fmt.Sprintf("Added `%v` to the filter.", phrase))
	default:
		msg.Reply("there was an error, please try again")
	}
}

type FilterWordListCommand struct {
	m       *ModerationMod
	Enabled bool
}

func NewFilterWordListCommand(m *ModerationMod) meidov2.ModCommand {
	return &FilterWordListCommand{
		m:       m,
		Enabled: true,
	}
}
func (c *FilterWordListCommand) Name() string {
	return "filterwordlist"
}
func (c *FilterWordListCommand) Description() string {
	return "Lists of all filtered phrases for this server"
}
func (c *FilterWordListCommand) Triggers() []string {
	return []string{"m?fwl", "m?filterwordlist"}
}
func (c *FilterWordListCommand) Usage() string {
	return "m?fwl"
}
func (c *FilterWordListCommand) Cooldown() int {
	return 10
}
func (c *FilterWordListCommand) RequiredPerms() int {
	return discordgo.PermissionManageMessages
}
func (c *FilterWordListCommand) RequiresOwner() bool {
	return false
}
func (c *FilterWordListCommand) IsEnabled() bool {
	return c.Enabled
}
func (c *FilterWordListCommand) Run(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 1 || (msg.Args()[0] != "m?filterwordlist" && msg.Args()[0] != "m?fwl") {
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
	if uPerms&discordgo.PermissionManageMessages == 0 && uPerms&discordgo.PermissionAdministrator == 0 {
		return
	}

	c.m.cl <- msg

	var fel []*FilterEntry
	err = c.m.db.Select(&fel, "SELECT * FROM filters WHERE guild_id=$1;", msg.Message.GuildID)
	if err != nil {
		fmt.Println(err)
		return
	}
	if len(fel) < 1 {
		msg.Reply("filter is empty")
		return
	}
	filterListBuilder := strings.Builder{}
	filterListBuilder.WriteString("```\nList of currently filtered phrases\n")
	for _, fe := range fel {
		filterListBuilder.WriteString(fmt.Sprintf("- %s\n", fe.Phrase))
	}
	filterListBuilder.WriteString("```")

	if len(filterListBuilder.String()) > 1000 {
		buf := &bytes.Buffer{}
		buf.WriteString(filterListBuilder.String())

		msg.Sess.ChannelFileSend(msg.Message.ChannelID, "filter.txt", buf)
	} else {
		msg.Reply(filterListBuilder.String())
	}
}

type ClearFilterCommand struct {
	m       *ModerationMod
	Enabled bool
}

func NewClearFilterCommand(m *ModerationMod) meidov2.ModCommand {
	return &ClearFilterCommand{
		m:       m,
		Enabled: true,
	}
}
func (c *ClearFilterCommand) Name() string {
	return "ClearFilter"
}

func (c *ClearFilterCommand) Description() string {
	return "Removes all phrases from the server filter"
}

func (c *ClearFilterCommand) Triggers() []string {
	return []string{"m?clearfilter"}
}

func (c *ClearFilterCommand) Usage() string {
	return "m?clearfilter"
}

func (c *ClearFilterCommand) Cooldown() int {
	return 30
}

func (c *ClearFilterCommand) RequiredPerms() int {
	return discordgo.PermissionAdministrator
}

func (c *ClearFilterCommand) RequiresOwner() bool {
	return false
}

func (c *ClearFilterCommand) IsEnabled() bool {
	return c.Enabled
}

func (c *ClearFilterCommand) Run(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 1 || msg.Args()[0] != "m?clearfilter" {
		return
	}

	uPerms, err := msg.Discord.UserChannelPermissions(msg.Member, msg.Message.ChannelID)
	if err != nil {
		fmt.Println(err)
		return
	}
	if uPerms&discordgo.PermissionAdministrator == 0 {
		return
	}

	c.m.cl <- msg

	_, err = c.m.db.Exec("DELETE FROM filters WHERE guild_id=$1", msg.Message.GuildID)
	if err != nil {
		msg.Reply("there was an error")
		return
	}

	msg.Reply("Filter was cleared")
}

type ToggleStrikeCommand struct {
	m       *ModerationMod
	Enabled bool
}

func NewToggleStrikeCommand(m *ModerationMod) meidov2.ModCommand {
	return &ToggleStrikeCommand{
		m:       m,
		Enabled: true,
	}
}
func (c *ToggleStrikeCommand) Name() string {
	return "ToggleStrikes"
}
func (c *ToggleStrikeCommand) Description() string {
	return "Toggles strike system for the server"
}
func (c *ToggleStrikeCommand) Triggers() []string {
	return []string{"m?togglestrikes"}
}
func (c *ToggleStrikeCommand) Usage() string {
	return "m?togglestrikes"
}
func (c *ToggleStrikeCommand) Cooldown() int {
	return 30
}
func (c *ToggleStrikeCommand) RequiredPerms() int {
	return discordgo.PermissionAdministrator
}
func (c *ToggleStrikeCommand) RequiresOwner() bool {
	return false
}
func (c *ToggleStrikeCommand) IsEnabled() bool {
	return c.Enabled
}
func (c *ToggleStrikeCommand) Run(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 1 || msg.Args()[0] != "m?togglestrikes" {
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
	if uPerms&discordgo.PermissionAdministrator == 0 {
		return
	}

	c.m.cl <- msg

	dge := &DiscordGuild{}
	err = c.m.db.Get(dge, "SELECT * FROM guilds WHERE guild_id = $1", msg.Message.GuildID)
	if err != nil {
		return
	}
	if dge.UseWarns {
		c.m.db.Exec("UPDATE guilds SET use_warns=false WHERE guild_id=$1 AND use_warns=true", dge.GuildID)
		msg.Reply("Strike system is now DISABLED")
	} else {
		c.m.db.Exec("UPDATE guilds SET use_warns=true WHERE guild_id=$1 AND use_warns=false", dge.GuildID)
		msg.Reply("Strike system is now ENABLED")
	}
}

type ModerationSettingsCommand struct {
	m       *ModerationMod
	Enabled bool
}

func NewModerationSettingsCommand(m *ModerationMod) meidov2.ModCommand {
	return &ModerationSettingsCommand{
		m:       m,
		Enabled: true,
	}
}
func (c *ModerationSettingsCommand) Name() string {
	return "moderationsettings"
}
func (c *ModerationSettingsCommand) Description() string {
	return "Moderation settings:\n- Toggle warn system [enable/disable]\n- Set max warns [0-10]\n- Set warn duration [0(infinite)-365]"
}
func (c *ModerationSettingsCommand) Triggers() []string {
	return []string{"m?settings moderation"}
}
func (c *ModerationSettingsCommand) Usage() string {
	return "m?settings moderation warns enable/disable\nm?settings moderation maxwarns [0-10]\nm?settings moderation warnduration [0-365]"
}
func (c *ModerationSettingsCommand) Cooldown() int {
	return 5
}
func (c *ModerationSettingsCommand) RequiredPerms() int {
	return discordgo.PermissionAdministrator
}
func (c *ModerationSettingsCommand) RequiresOwner() bool {
	return false
}
func (c *ModerationSettingsCommand) IsEnabled() bool {
	return c.Enabled
}
func (c *ModerationSettingsCommand) Run(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 4 || strings.Join(msg.Args()[:2], " ") != "m?settings moderation" {
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
	if uPerms&discordgo.PermissionAdministrator == 0 {
		return
	}

	c.m.cl <- msg

	switch msg.Args()[2] {
	case "warns":
		if msg.Args()[3] == "enable" {
			c.m.db.Exec("UPDATE guilds SET use_warns=true WHERE guild_id=$1 AND NOT use_warns", msg.Message.GuildID)
			msg.Reply("Strike system is now ENABLED")

		} else if msg.Args()[3] == "disable" {
			c.m.db.Exec("UPDATE guilds SET use_warns=false WHERE guild_id=$1 AND use_warns", msg.Message.GuildID)
			msg.Reply("Strike system is now DISABLED")
		}
	case "maxwarns":

		n, err := strconv.Atoi(msg.Args()[3])
		if err != nil {
			return
		}

		n = meidov2.Clamp(0, 10, n)

		_, err = c.m.db.Exec("UPDATE guilds SET max_warns=$1 WHERE guild_id=$2", n, msg.Message.GuildID)
		if err != nil {
			msg.Reply("error setting max warns")
			return
		}
		msg.Reply(fmt.Sprintf("set max warns to %v", n))
	case "warnduration":

		n, err := strconv.Atoi(msg.Args()[3])
		if err != nil {
			return
		}

		n = meidov2.Clamp(0, 365, n)

		_, err = c.m.db.Exec("UPDATE guilds SET warn_duration=$1 WHERE guild_id=$2", n, msg.Message.GuildID)
		if err != nil {
			msg.Reply("error setting warn duration")
			return
		}
		msg.Reply(fmt.Sprintf("set warn duration to %v days", n))
	}
}

func (m *ModerationMod) CheckFilter(msg *meidov2.DiscordMessage) {

	isIllegal := false
	trigger := ""

	uPerms, err := msg.Discord.UserChannelPermissions(msg.Member, msg.Message.ChannelID)
	if err != nil {
		return
	}
	if uPerms&discordgo.PermissionManageMessages != 0 || uPerms&discordgo.PermissionAdministrator != 0 {
		return
	}

	var filterEntries []*FilterEntry
	err = m.db.Select(&filterEntries, "SELECT phrase FROM filters WHERE guild_id=$1", msg.Message.GuildID)
	if err != nil {
		return
	}

	for _, entry := range filterEntries {
		if strings.Contains(strings.ToLower(msg.Message.Content), strings.ToLower(entry.Phrase)) {
			trigger = entry.Phrase
			isIllegal = true
			break
		}
	}

	if !isIllegal {
		return
	}

	msg.Sess.ChannelMessageDelete(msg.Message.ChannelID, msg.Message.ID)

	dge := &DiscordGuild{}
	err = m.db.Get(dge, "SELECT use_warns, max_warns FROM guilds WHERE guild_id=$1", msg.Message.GuildID)
	if err != nil {
		return
	}
	if !dge.UseWarns {
		msg.Reply(fmt.Sprintf("%v, you are not allowed to use a banned word/phrase", msg.Message.Author.Mention()))
		return
	}

	reason := "Triggering filter: " + trigger
	warnCount := 0

	err = m.db.Get(&warnCount, "SELECT COUNT(*) FROM warns WHERE user_id=$1 AND guild_id=$2 AND is_valid",
		msg.Message.Author.ID, msg.Message.GuildID)
	if err != nil {
		return
	}

	g, err := msg.Discord.Sess.State.Guild(msg.Message.GuildID)
	if err != nil {
		return
	}
	cu := msg.Discord.Sess.State.User

	_, err = m.db.Exec("INSERT INTO warns VALUES(DEFAULT, $1, $2, $3, $4, $5, $6)",
		msg.Message.GuildID, msg.Message.Author.ID, reason, cu.ID, time.Now(), true)
	if err != nil {
		fmt.Println(err)
		return
	}

	userChannel, userChError := msg.Discord.Sess.UserChannelCreate(msg.Message.Author.ID)

	// 3 / 3 strikes
	if warnCount+1 >= dge.MaxWarns {

		if userChError == nil {
			msg.Discord.Sess.ChannelMessageSend(userChannel.ID, fmt.Sprintf("You have been banned from %v for acquiring %v warns.\nLast warning was: %v",
				g.Name, dge.MaxWarns, reason))
		}
		err = msg.Discord.Sess.GuildBanCreateWithReason(g.ID, msg.Message.Author.ID, reason, 0)
		if err != nil {
			return
		}
		/*
			_, err = m.db.Exec("INSERT INTO warns VALUES(DEFAULT, $1, $2, $3, $4, $5, $6)",
				msg.Message.GuildID, msg.Message.Author.ID, reason, cu.ID, time.Now(), false)
		*/

		_, err = m.db.Exec("UPDATE warns SET is_valid=false, cleared_by_id=$1, cleared_at=$2 WHERE guild_id=$3 AND user_id=$4 and is_valid",
			cu.ID, time.Now(), g.ID, msg.Message.Author.ID)

		msg.Reply(fmt.Sprintf("%v has been banned after acquiring too many warns. miss them.", msg.Message.Author.Mention()))

	} else {
		if userChError == nil {
			msg.Discord.Sess.ChannelMessageSend(userChannel.ID, fmt.Sprintf("You have been warned in %v.\nWarned for: %v\nYou are currently at warn %v/%v",
				g.Name, reason, warnCount+1, dge.MaxWarns))
		}
		/*
			_, err = m.db.Exec("INSERT INTO warns VALUES(DEFAULT, $1, $2, $3, $4, $5, $6)",
				msg.Message.GuildID, msg.Message.Author.ID, reason, cu.ID, time.Now(), true)
		*/
		msg.Reply(fmt.Sprintf("%v has been warned\nThey are currently at warn %v/%v", msg.Message.Author.Mention(), warnCount+1, dge.MaxWarns))
	}
}
