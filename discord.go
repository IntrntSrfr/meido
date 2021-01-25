package meido

import (
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"net/http"
	"runtime/debug"
	"sort"
	"time"
)

// Discord represents the part of the bot that deals with interaction with Discord.
type Discord struct {
	token    string
	Sess     *discordgo.Session
	Sessions []*discordgo.Session
	ownerIds []string

	messageChan chan *DiscordMessage
}

// NewDiscord takes in a token and creates a Discord object.
func NewDiscord(token string) *Discord {
	return &Discord{
		token:       token,
		messageChan: make(chan *DiscordMessage, 256),
	}
}

// Open populates the Discord object with Sessions and returns a DiscordMessage channel.
func (d *Discord) Open() (<-chan *DiscordMessage, error) {

	shardCount, err := recommendedShards(d.token)
	if err != nil {
		panic(err)
	}

	d.Sessions = make([]*discordgo.Session, shardCount)

	for i := 0; i < shardCount; i++ {
		s, err := discordgo.New("Bot " + d.token)
		if err != nil {
			return nil, err
		}

		s.State.TrackVoice = false
		s.State.TrackPresences = false
		s.ShardCount = shardCount
		s.ShardID = i
		s.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAllWithoutPrivileged | discordgo.IntentsGuildMembers)

		s.AddHandler(d.onMessageCreate)
		s.AddHandler(d.onMessageUpdate)
		s.AddHandler(d.onMessageDelete)

		d.Sessions[i] = s
		fmt.Println("created session:", i)
	}
	d.Sess = d.Sessions[0]

	return d.messageChan, nil
}

// recommendedShards asks discord for the recommended shardcount for the bot given the token.
// returns -1 if the request does not go well.
func recommendedShards(token string) (int, error) {
	req, _ := http.NewRequest("GET", "https://discord.com/api/v8/gateway/bot", nil)
	req.Header.Add("Authorization", "Bot "+token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return -1, err
	}
	defer res.Body.Close()

	resp := &discordgo.GatewayBotResponse{}
	err = json.NewDecoder(res.Body).Decode(&resp)
	if err != nil {
		return -1, err
	}

	return resp.Shards, nil
}

// Run opens the Discord sessions.
func (d *Discord) Run() error {
	for _, sess := range d.Sessions {
		sess.Open()
	}
	return nil
}

// Close closes the Discord sessions
func (d *Discord) Close() {
	for _, sess := range d.Sessions {
		sess.Close()
	}
}

// BotRecover
func BotRecover() {
	if r := recover(); r != nil {
		log.Println("Recovery:", r)
		log.Println(string(debug.Stack()))
	}
}

// onMessageCreate is the handler for the *discordgo.MessageCreate event.
// It populates a DiscordMessage object and sends it to Discord.messageChan
func (d *Discord) onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author == nil || m.Message.Author.Bot {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovery:", r)
			log.Println()
			log.Println(string(debug.Stack()))
			log.Println()

			if data, err := json.MarshalIndent(m, "", "\t"); err == nil {
				log.Println(string(data))
				log.Println()
			}
		}
	}()

	var author *discordgo.User
	var member *discordgo.Member

	author = m.Author

	if m.GuildID != "" {
		member = m.Member
		member.User = author
	}

	d.messageChan <- &DiscordMessage{
		Sess:         s,
		Discord:      d,
		Message:      m.Message,
		Author:       author,
		Member:       member,
		Type:         MessageTypeCreate,
		TimeReceived: time.Now(),
		Shard:        s.ShardID,
	}
}

// onMessageUpdate is the handler for the *discordgo.MessageUpdate event.
// It populates a DiscordMessage object and sends it to Discord.messageChan
func (d *Discord) onMessageUpdate(s *discordgo.Session, m *discordgo.MessageUpdate) {
	if m.Author == nil || m.Message.Author.Bot {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovery:", r)
			log.Println()
			log.Println(string(debug.Stack()))
			log.Println()

			if data, err := json.MarshalIndent(m, "", "\t"); err == nil {
				log.Println(string(data))
				log.Println()
			}
		}
	}()

	var author *discordgo.User
	var member *discordgo.Member

	author = m.Author

	if m.GuildID != "" {
		member = m.Member
		member.User = author
	}

	d.messageChan <- &DiscordMessage{
		Sess:         s,
		Discord:      d,
		Message:      m.Message,
		Author:       author,
		Member:       member,
		Type:         MessageTypeUpdate,
		TimeReceived: time.Now(),
		Shard:        s.ShardID,
	}
}

// onMessageDelete is the handler for the *discordgo.MessageDelete event.
// It populates a DiscordMessage object and sends it to Discord.messageChan
func (d *Discord) onMessageDelete(s *discordgo.Session, m *discordgo.MessageDelete) {
	d.messageChan <- &DiscordMessage{
		Sess:         s,
		Discord:      d,
		Message:      m.Message,
		Type:         MessageTypeDelete,
		TimeReceived: time.Now(),
		Shard:        s.ShardID,
	}
}

// UserChannelPermissionsDirect finds permissions based on the *discordgo.Member object directly
func (d *Discord) UserChannelPermissionsDirect(m *discordgo.Member, channelID string) (apermissions int, err error) {

	channel, err := d.Channel(channelID)
	if err != nil {
		return
	}

	guild, err := d.Guild(channel.GuildID)
	if err != nil {
		return
	}

	if m.User.ID == guild.OwnerID {
		apermissions = discordgo.PermissionAll
		return
	}

	return memberPermissions(guild, channel, m), nil
}

// UserChannelPermissions finds member permissions the usual way, using just the IDs.
func (d *Discord) UserChannelPermissions(userID, channelID string) (int, error) {
	var (
		err         error
		permissions int
	)
	for _, s := range d.Sessions {
		permissions, err = s.State.UserChannelPermissions(userID, channelID)
		if err == nil {
			return permissions, nil
		}
	}
	return permissions, err
}

// HasPermissions finds if the bot user has permissions in a channel.
func (d *Discord) HasPermissions(channelID string, perm int) bool {
	uPerms, err := d.UserChannelPermissions(d.Sess.State.User.ID, channelID)
	if err != nil {
		return false
	}
	return uPerms&perm != 0 || uPerms&discordgo.PermissionAdministrator != 0
}

// yoinked from discordgo lol. I dont remember why at this point.
// Calculates the permissions for a member.
// https://support.discord.com/hc/en-us/articles/206141927-How-is-the-permission-hierarchy-structured-
func memberPermissions(guild *discordgo.Guild, channel *discordgo.Channel, member *discordgo.Member) (apermissions int) {
	userID := member.User.ID

	if userID == guild.OwnerID {
		apermissions = discordgo.PermissionAll
		return
	}

	for _, role := range guild.Roles {
		if role.ID == guild.ID {
			apermissions |= role.Permissions
			break
		}
	}

	for _, role := range guild.Roles {
		for _, roleID := range member.Roles {
			if role.ID == roleID {
				apermissions |= role.Permissions
				break
			}
		}
	}

	if apermissions&discordgo.PermissionAdministrator == discordgo.PermissionAdministrator {
		apermissions |= discordgo.PermissionAll
	}

	// Apply @everyone overrides from the channel.
	for _, overwrite := range channel.PermissionOverwrites {
		if guild.ID == overwrite.ID {
			apermissions &= ^overwrite.Deny
			apermissions |= overwrite.Allow
			break
		}
	}

	denies := 0
	allows := 0

	// Member overwrites can override role overrides, so do two passes
	for _, overwrite := range channel.PermissionOverwrites {
		for _, roleID := range member.Roles {
			if overwrite.Type == "role" && roleID == overwrite.ID {
				denies |= overwrite.Deny
				allows |= overwrite.Allow
				break
			}
		}
	}

	apermissions &= ^denies
	apermissions |= allows

	for _, overwrite := range channel.PermissionOverwrites {
		if overwrite.Type == "member" && overwrite.ID == userID {
			apermissions &= ^overwrite.Deny
			apermissions |= overwrite.Allow
			break
		}
	}

	if apermissions&discordgo.PermissionAdministrator == discordgo.PermissionAdministrator {
		apermissions |= discordgo.PermissionAllChannel
	}

	return apermissions
}

// HighestRole finds the highest role a user has in the guild hierarchy.
func (d *Discord) HighestRole(gid, uid string) *discordgo.Role {

	g, err := d.Guild(gid)
	if err != nil {
		return nil
	}
	mem, err := d.Member(gid, uid)
	if err != nil {
		return nil
	}

	gRoles := g.Roles

	sort.Sort(RoleByPos(gRoles))

	for _, gr := range gRoles {
		for _, r := range mem.Roles {
			if r == gr.ID {
				return gr
			}
		}
	}

	return nil
}

// HighestRolePosition gets the highest role position a user has in the guild hierarchy.
func (d *Discord) HighestRolePosition(gid, uid string) int {
	role := d.HighestRole(gid, uid)
	if role == nil {
		return -1
	}
	return role.Position
}

// HighestColor finds the role color for the top-most role where the color is non-zero.
func (d *Discord) HighestColor(gid, uid string) int {

	g, err := d.Guild(gid)
	if err != nil {
		return 0
	}

	mem, err := d.Member(gid, uid)
	if err != nil {
		return 0
	}

	gRoles := g.Roles

	sort.Sort(RoleByPos(gRoles))

	for _, gr := range gRoles {
		for _, r := range mem.Roles {
			if r == gr.ID {
				if gr.Color != 0 {
					return gr.Color
				}
			}
		}
	}

	return 0
}

// RoleByPos is used to sort discordgo.Roles so that it accurately represents how the hierarchy looks.
type RoleByPos []*discordgo.Role

func (a RoleByPos) Len() int           { return len(a) }
func (a RoleByPos) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a RoleByPos) Less(i, j int) bool { return a[i].Position > a[j].Position }

// Guilds returns all the guild objects the bot has in its sessions.
func (d *Discord) Guilds() []*discordgo.Guild {
	var guilds []*discordgo.Guild
	for _, sess := range d.Sessions {
		guilds = append(guilds, sess.State.Guilds...)
	}
	return guilds
}

// Guild takes in an ID and returns a discordgo.Guild if one with that ID exists.
func (d *Discord) Guild(guildID string) (*discordgo.Guild, error) {
	var err error
	for _, sess := range d.Sessions {
		g, err := sess.State.Guild(guildID)
		if err == nil {
			return g, nil
		}
	}
	return nil, err
}

// Channel takes in an ID and returns a discordgo.Channel if one with that ID exists.
func (d *Discord) Channel(channelID string) (*discordgo.Channel, error) {
	var err error
	for _, sess := range d.Sessions {
		ch, err := sess.State.Channel(channelID)
		if err == nil {
			return ch, nil
		}
	}
	return nil, err
}

// Role takes in a guild ID and a role ID and returns a discordgo.Role if one with such IDs exists.
func (d *Discord) Role(guildID, roleID string) (*discordgo.Role, error) {
	var err error
	for _, sess := range d.Sessions {
		r, err := sess.State.Role(guildID, roleID)
		if err == nil {
			return r, nil
		}
	}
	return nil, err
}

// Member takes in a guild ID and a user ID and returns a discordgo.Member if one with such ID exists.
func (d *Discord) Member(guildID, userID string) (*discordgo.Member, error) {
	var err error
	for _, sess := range d.Sessions {
		mem, err := sess.State.Member(guildID, userID)
		if err == nil {
			return mem, nil
		}
	}
	mem, err := d.Sess.GuildMember(guildID, userID)
	if err != nil {
		return nil, err
	}
	return mem, nil
}

// IsOwner returns whether the author of a DiscordMessage is a bot owner by checking
// the IDs in the ownerIDs in the Discord struct.
func (d *Discord) IsOwner(msg *DiscordMessage) bool {
	isOwner := false
	for _, id := range d.ownerIds {
		if msg.Author.ID == id {
			isOwner = true
		}
	}
	return isOwner
}

// StartTyping makes the bot show as 'typing..' in a channel.
func (d *Discord) StartTyping(channelID string) error {
	return d.Sess.ChannelTyping(channelID)
}
