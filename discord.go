package meidov2

import (
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"net/http"
	"sort"
	"time"
)

type Discord struct {
	token    string
	Sess     *discordgo.Session
	Sessions []*discordgo.Session
	ownerIds []string

	messageChan chan *DiscordMessage
}

func NewDiscord(token string) *Discord {
	return &Discord{
		token:       token,
		messageChan: make(chan *DiscordMessage, 256),
	}
}

func (d *Discord) Open() (<-chan *DiscordMessage, error) {
	req, _ := http.NewRequest("GET", "https://discord.com/api/v8/gateway/bot", nil)
	req.Header.Add("Authorization", "Bot "+d.token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	resp := &discordgo.GatewayBotResponse{}
	err = json.NewDecoder(res.Body).Decode(&resp)
	if err != nil {
		panic(err)
	}

	shardCount := resp.Shards
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

func (d *Discord) Run() error {
	for _, sess := range d.Sessions {
		sess.Open()
	}
	return nil
}

func (d *Discord) Close() {
	for _, sess := range d.Sessions {
		sess.Close()
	}
}

func (d *Discord) onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author == nil || m.Message.Author.Bot {
		return
	}

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

func (d *Discord) onMessageUpdate(s *discordgo.Session, m *discordgo.MessageUpdate) {
	if m.Author == nil || m.Message.Author.Bot {
		return
	}

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

func (d *Discord) HasPermissions(channelID string, perm int) bool {
	uPerms, err := d.UserChannelPermissions(d.Sess.State.User.ID, channelID)
	if err != nil {
		return false
	}
	return uPerms&perm != 0 || uPerms&discordgo.PermissionAdministrator != 0
}

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
func (d *Discord) HighestRole(gid, uid string) int {

	g, err := d.Guild(gid)
	if err != nil {
		return -1
	}
	mem, err := d.Member(gid, uid)
	if err != nil {
		return -1
	}

	gRoles := g.Roles

	sort.Sort(RoleByPos(gRoles))

	for _, gr := range gRoles {
		for _, r := range mem.Roles {
			if r == gr.ID {
				return gr.Position
			}
		}
	}

	return -1
}
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

type RoleByPos []*discordgo.Role

func (a RoleByPos) Len() int           { return len(a) }
func (a RoleByPos) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a RoleByPos) Less(i, j int) bool { return a[i].Position > a[j].Position }

func (d *Discord) Guilds() []*discordgo.Guild {
	var guilds []*discordgo.Guild
	for _, sess := range d.Sessions {
		guilds = append(guilds, sess.State.Guilds...)
	}
	return guilds
}
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

func (d *Discord) IsOwner(msg *DiscordMessage) bool {
	isOwner := false
	for _, id := range d.ownerIds {
		if msg.Author.ID == id {
			isOwner = true
		}
	}
	return isOwner
}

func (d *Discord) StartTyping(channelID string) error {
	return d.Sess.ChannelTyping(channelID)
}
