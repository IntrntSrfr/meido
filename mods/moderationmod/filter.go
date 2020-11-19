package moderationmod

import (
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
	return "FilterWord"
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
	return "FilterWordList"
}
func (c *FilterWordListCommand) Description() string {
	return "Displays a list of all filtered phrases for this server"
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
	msg.Reply(filterListBuilder.String())

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
	if dge.UseStrikes {
		c.m.db.Exec("UPDATE guilds SET use_strikes=false WHERE guild_id=$1 AND use_strikes=true", dge.GuildID)
		msg.Reply("Strike system is now DISABLED")
	} else {
		c.m.db.Exec("UPDATE guilds SET use_strikes=true WHERE guild_id=$1 AND use_strikes=false", dge.GuildID)
		msg.Reply("Strike system is now ENABLED")
	}
}

type SetMaxWarnsCommand struct {
	m       *ModerationMod
	Enabled bool
}

func NewSetMaxWarnsCommand(m *ModerationMod) meidov2.ModCommand {
	return &SetMaxWarnsCommand{
		m:       m,
		Enabled: true,
	}
}
func (c *SetMaxWarnsCommand) Name() string {
	return "SetMaxWarns"
}
func (c *SetMaxWarnsCommand) Description() string {
	return "Sets the amount of warns a user needs to acquire to get banned."
}
func (c *SetMaxWarnsCommand) Triggers() []string {
	return []string{"m?setmaxwarns"}
}
func (c *SetMaxWarnsCommand) Usage() string {
	return "m?setmaxwarns 5"
}
func (c *SetMaxWarnsCommand) Cooldown() int {
	return 10
}
func (c *SetMaxWarnsCommand) RequiredPerms() int {
	return discordgo.PermissionAdministrator
}
func (c *SetMaxWarnsCommand) RequiresOwner() bool {
	return false
}
func (c *SetMaxWarnsCommand) IsEnabled() bool {
	return c.Enabled
}
func (c *SetMaxWarnsCommand) Run(msg *meidov2.DiscordMessage) {

	if msg.LenArgs() < 2 || msg.Args()[0] != "m?maxstrikes" {
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

	n, err := strconv.Atoi(msg.Args()[1])
	if err != nil {
		return
	}
	if n < 0 {
		n = 0
	} else if n > 10 {
		n = 10
	}

	_, err = c.m.db.Exec("UPDATE guilds SET max_strikes=$1 WHERE guild_id=$2", n, msg.Message.GuildID)
	if err != nil {
		msg.Reply("error setting max strikes")
		return
	}
	msg.Reply(fmt.Sprintf("set max strikes to %v", n))
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

	dge := &DiscordGuild{}
	err = m.db.Get(dge, "SELECT use_strikes, max_strikes FROM guilds WHERE guild_id=$1", msg.Message.GuildID)
	if err != nil {
		return
	}
	if !dge.UseStrikes {
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
	if warnCount+1 >= dge.MaxStrikes {

		if userChError == nil {
			msg.Discord.Sess.ChannelMessageSend(userChannel.ID, fmt.Sprintf("You have been banned from %v for acquiring %v warns.\nLast warning was: %v",
				g.Name, dge.MaxStrikes, reason))
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
				g.Name, reason, warnCount+1, dge.MaxStrikes))
		}
		/*
			_, err = m.db.Exec("INSERT INTO warns VALUES(DEFAULT, $1, $2, $3, $4, $5, $6)",
				msg.Message.GuildID, msg.Message.Author.ID, reason, cu.ID, time.Now(), true)
		*/
		msg.Reply(fmt.Sprintf("%v has been warned\nThey are currently at warn %v/%v", msg.Message.Author.Mention(), warnCount+1, dge.MaxStrikes))
	}
}
