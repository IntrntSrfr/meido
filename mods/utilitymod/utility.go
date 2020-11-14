package utilitymod

import (
	"bytes"
	"fmt"
	"github.com/andersfylling/disgord"
	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/intrntsrfr/meidov2"
	"github.com/jmoiron/sqlx"
	"image"
	"image/color"
	"image/png"
	"math"
	"runtime"
	"strconv"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

type UtilityMod struct {
	Name string
	sync.Mutex
	cl        chan *meidov2.DiscordMessage
	commands  map[string]meidov2.ModCommand
	startTime time.Time
	db        *sqlx.DB
}

func New(name string) meidov2.Mod {
	return &UtilityMod{
		startTime: time.Now(),
		Name:      name,
		commands:  make(map[string]meidov2.ModCommand),
	}
}

func (m *UtilityMod) Save() error {
	return nil
}

func (m *UtilityMod) Load() error {
	return nil
}

func (m *UtilityMod) Settings(msg *meidov2.DiscordMessage) {

}

func (m *UtilityMod) Help(msg *meidov2.DiscordMessage) {

}

func (m *UtilityMod) Commands() map[string]meidov2.ModCommand {
	return nil
}

func (m *UtilityMod) Hook(b *meidov2.Bot) error {
	m.cl = b.CommandLog
	m.db = b.DB

	b.Discord.Sess.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		statusTimer := time.NewTicker(time.Second * 15)
		oldMemCount := 0
		oldSrvCount := 0
		go func() {
			for range statusTimer.C {
				memCount := 0
				srvCount := 0
				for _, sess := range b.Discord.Sessions {
					srvCount++
					for _, g := range sess.State.Guilds {
						memCount += g.MemberCount
					}
				}

				if memCount == oldMemCount && srvCount == oldSrvCount {
					continue
				}

				s.UpdateStatusComplex(discordgo.UpdateStatusData{
					Game: &discordgo.Game{
						Name: fmt.Sprintf("BEING REWORKED, %v MEMBERS, %v SERVERS", memCount, srvCount),
						Type: discordgo.GameTypeGame,
					},
				})
				oldMemCount = memCount
				oldSrvCount = srvCount
			}
		}()
	})

	m.RegisterCommand(NewAvatarCommand(m))
	m.RegisterCommand(NewAboutCommand(m))
	m.RegisterCommand(NewServerCommand(m))
	m.RegisterCommand(NewServerBannerCommand(m))
	m.RegisterCommand(NewServerSplashCommand(m))
	m.RegisterCommand(NewColorCommand(m))

	//m.commands = append(m.commands, m.Avatar, m.About, m.Server, m.ServerBanner, m.ServerSplash)

	return nil
}

func (m *UtilityMod) RegisterCommand(cmd meidov2.ModCommand) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.commands[cmd.Name()]; ok {
		panic(fmt.Sprintf("command '%v' already exists in %v", cmd.Name(), m.Name))
	}
	m.commands[cmd.Name()] = cmd
}

func (m *UtilityMod) Message(msg *meidov2.DiscordMessage) {
	if msg.Type != meidov2.MessageTypeCreate {
		return
	}
	for _, c := range m.commands {
		go c.Run(msg)
	}
}

type AvatarCommand struct {
	m       *UtilityMod
	Enabled bool
}

func NewAvatarCommand(m *UtilityMod) meidov2.ModCommand {
	return &AvatarCommand{
		m:       m,
		Enabled: true,
	}
}
func (c *AvatarCommand) Name() string {
	return "Avatar"
}
func (c *AvatarCommand) Description() string {
	return "Displays the profile picture of either the author or whoever is mentioned"
}
func (c *AvatarCommand) Triggers() []string {
	return []string{"m?avatar", "m?av", ">av"}
}
func (c *AvatarCommand) Usage() string {
	return ">av\n>av 123123123123\nm?av\nm?avatar"
}
func (c *AvatarCommand) Cooldown() int {
	return 10
}
func (c *AvatarCommand) RequiredPerms() int {
	return 0
}
func (c *AvatarCommand) RequiresOwner() bool {
	return false
}
func (c *AvatarCommand) IsEnabled() bool {
	return c.Enabled
}

// true and (not m?av and not >av)
func (c *AvatarCommand) Run(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 1 || (msg.Args()[0] != "m?av" && msg.Args()[0] != ">av") {
		return
	}

	c.m.cl <- msg

	var targetUser *discordgo.User
	var err error

	if msg.LenArgs() > 1 {
		if len(msg.Message.Mentions) >= 1 {
			targetUser = msg.Message.Mentions[0]
		} else {
			targetUser, err = msg.Sess.User(msg.Args()[1])
			if err != nil {
				return
			}
		}
	} else {
		targetUser, err = msg.Discord.Sess.User(msg.Message.Author.ID)
		if err != nil {
			return
		}
	}

	if targetUser == nil {
		return
	}

	if targetUser.Avatar == "" {
		msg.ReplyEmbed(&discordgo.MessageEmbed{
			Color:       0xC80000,
			Description: fmt.Sprintf("%v has no avatar set.", targetUser.String()),
		})
	} else {
		msg.ReplyEmbed(&discordgo.MessageEmbed{
			Color: msg.HighestColor(msg.Message.GuildID, targetUser.ID),
			Title: targetUser.String(),
			Image: &discordgo.MessageEmbedImage{URL: targetUser.AvatarURL("1024")},
		})
	}
}

// only here in case of disgord
func AvatarURL(u *disgord.User, size int) string {
	a, _ := u.AvatarURL(size, true)
	return a
}

type ServerCommand struct {
	m       *UtilityMod
	Enabled bool
}

func NewServerCommand(m *UtilityMod) meidov2.ModCommand {
	return &ServerCommand{
		m:       m,
		Enabled: true,
	}
}
func (c *ServerCommand) Name() string {
	return "Server"
}
func (c *ServerCommand) Description() string {
	return "Displays information about the current server"
}
func (c *ServerCommand) Triggers() []string {
	return []string{"m?server"}
}
func (c *ServerCommand) Usage() string {
	return "m?server"
}
func (c *ServerCommand) Cooldown() int {
	return 10
}
func (c *ServerCommand) RequiredPerms() int {
	return 0
}
func (c *ServerCommand) RequiresOwner() bool {
	return false
}
func (c *ServerCommand) IsEnabled() bool {
	return c.Enabled
}
func (c *ServerCommand) Run(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 1 || msg.Args()[0] != "m?server" {
		return
	}
	if msg.IsDM() {
		return
	}
	c.m.cl <- msg

	g, err := msg.Discord.Sess.State.Guild(msg.Message.GuildID)
	if err != nil {
		msg.Reply("Error getting guild data")
		return
	}

	tc := 0
	vc := 0

	for _, ch := range g.Channels {
		if ch.Type == discordgo.ChannelTypeGuildText {
			tc++
		} else if ch.Type == discordgo.ChannelTypeGuildVoice {
			vc++
		}
	}

	owner, err := msg.Discord.Sess.GuildMember(g.ID, g.OwnerID)
	if err != nil {
		msg.Reply("Error getting guild data")
		return
	}

	id, err := strconv.ParseInt(g.ID, 10, 64)
	if err != nil {
		return
	}

	id = ((id >> 22) + 1420070400000) / 1000

	dur := time.Since(time.Unix(id, 0))

	ts := time.Unix(id, 0)

	embed := discordgo.MessageEmbed{
		Color: 0xFFFFFF,
		Author: &discordgo.MessageEmbedAuthor{
			Name: g.Name,
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Owner",
				Value:  fmt.Sprintf("%v\n(%v)", owner.Mention(), owner.User.ID),
				Inline: true,
			},
			{
				Name:  "Creation date",
				Value: fmt.Sprintf("%v\n%v days ago", ts.Format(time.RFC1123), math.Floor(dur.Hours()/24.0)),
			},
			{
				Name:   "Members",
				Value:  fmt.Sprintf("%v members", g.MemberCount),
				Inline: true,
			},
			{
				Name:   "Channels",
				Value:  fmt.Sprintf("Total: %v\nText: %v\nVoice: %v", len(g.Channels), tc, vc),
				Inline: true,
			},
			{
				Name:   "Roles",
				Value:  fmt.Sprintf("%v roles", len(g.Roles)),
				Inline: true,
			},
		},
	}
	if g.Icon != "" {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: fmt.Sprintf("https://cdn.discordapp.com/icons/%v/%v.png", g.ID, g.Icon),
		}
		embed.Author.IconURL = fmt.Sprintf("https://cdn.discordapp.com/icons/%v/%v.png", g.ID, g.Icon)
	}

	msg.ReplyEmbed(&embed)
}

type AboutCommand struct {
	m       *UtilityMod
	Enabled bool
}

func NewAboutCommand(m *UtilityMod) meidov2.ModCommand {
	return &AboutCommand{
		m:       m,
		Enabled: true,
	}
}
func (c *AboutCommand) Name() string {
	return "About"
}

func (c *AboutCommand) Description() string {
	return "Displays current Meido statistics"
}

func (c *AboutCommand) Triggers() []string {
	return []string{"m?about"}
}
func (c *AboutCommand) Usage() string {
	return "m?about"
}
func (c *AboutCommand) Cooldown() int {
	return 30
}
func (c *AboutCommand) RequiredPerms() int {
	return 0
}
func (c *AboutCommand) RequiresOwner() bool {
	return false
}
func (c *AboutCommand) IsEnabled() bool {
	return c.Enabled
}
func (c *AboutCommand) Run(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 1 || msg.Args()[0] != "m?about" {
		return
	}
	c.m.cl <- msg

	var (
		totalUsers int
		/*
			totalBots   int
			totalHumans int
		*/
		memory        runtime.MemStats
		totalCommands int
	)
	runtime.ReadMemStats(&memory)
	guilds := msg.Discord.Sess.State.Guilds

	for _, guild := range guilds {
		totalUsers += guild.MemberCount
	}

	uptime := time.Now().Sub(c.m.startTime)
	err := c.m.db.Get(&totalCommands, "SELECT COUNT(*) FROM commandlog;")
	if err != nil {
		msg.Reply("Error getting data")
		return
	}

	msg.ReplyEmbed(&discordgo.MessageEmbed{
		Title: "About",
		Color: 0xFEFEFE,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Uptime",
				Value:  uptime.String(),
				Inline: true,
			},
			{
				Name:   "Total commands ran",
				Value:  strconv.Itoa(totalCommands),
				Inline: true,
			},
			{
				Name:   "Guilds",
				Value:  strconv.Itoa(len(guilds)),
				Inline: false,
			},
			{
				Name:   "Users",
				Value:  strconv.Itoa(totalUsers),
				Inline: true,
			},
			{
				Name:   "Current memory use",
				Value:  fmt.Sprintf("%v/%v", humanize.Bytes(memory.Alloc), humanize.Bytes(memory.Sys)),
				Inline: false,
			},
			{
				Name:   "Garbage collected",
				Value:  humanize.Bytes(memory.TotalAlloc - memory.Alloc),
				Inline: true,
			},
			{
				Name:   "Allocs | Frees",
				Value:  fmt.Sprintf("%v | %v", memory.Mallocs, memory.Frees),
				Inline: false,
			},
		},
	})
}

type ServerSplashCommand struct {
	m       *UtilityMod
	Enabled bool
}

func NewServerSplashCommand(m *UtilityMod) meidov2.ModCommand {
	return &ServerSplashCommand{
		m:       m,
		Enabled: true,
	}
}
func (c *ServerSplashCommand) Name() string {
	return "ServerSplash"
}
func (c *ServerSplashCommand) Description() string {
	return "Displays the server splash if one exists"
}
func (c *ServerSplashCommand) Triggers() []string {
	return []string{"m?serversplash"}
}
func (c *ServerSplashCommand) Usage() string {
	return "m?serversplash"
}
func (c *ServerSplashCommand) Cooldown() int {
	return 10
}
func (c *ServerSplashCommand) RequiredPerms() int {
	return 0
}
func (c *ServerSplashCommand) RequiresOwner() bool {
	return false
}
func (c *ServerSplashCommand) IsEnabled() bool {
	return c.Enabled
}
func (c *ServerSplashCommand) Run(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 1 || msg.Args()[0] != "m?serversplash" {
		return
	}
	if msg.IsDM() {
		return
	}

	c.m.cl <- msg

	g, err := msg.Discord.Sess.State.Guild(msg.Message.GuildID)
	if err != nil {
		return
	}

	if g.Splash == "" {
		msg.Reply("this server has no splash")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title: g.Name,
		Color: 0xFFFFFF,
		Image: &discordgo.MessageEmbedImage{
			URL: fmt.Sprintf("https://cdn.discordapp.com/splashes/%v/%v.png", g.ID, g.Splash),
		},
	}
	msg.ReplyEmbed(embed)
}

type ServerBannerCommand struct {
	m       *UtilityMod
	Enabled bool
}

func NewServerBannerCommand(m *UtilityMod) meidov2.ModCommand {
	return &ServerBannerCommand{
		m:       m,
		Enabled: true,
	}
}
func (c *ServerBannerCommand) Name() string {
	return "ServerBanner"
}
func (c *ServerBannerCommand) Description() string {
	return "Displays the server banner if one exists"
}
func (c *ServerBannerCommand) Triggers() []string {
	return []string{"m?serverbanner"}
}
func (c *ServerBannerCommand) Usage() string {
	return "m?serverbanner"
}
func (c *ServerBannerCommand) Cooldown() int {
	return 10
}
func (c *ServerBannerCommand) RequiredPerms() int {
	return 0
}
func (c *ServerBannerCommand) RequiresOwner() bool {
	return false
}
func (c *ServerBannerCommand) IsEnabled() bool {
	return c.Enabled
}
func (c *ServerBannerCommand) Run(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 1 || msg.Args()[0] != "m?serverbanner" {
		return
	}
	if msg.IsDM() {
		return
	}

	c.m.cl <- msg

	g, err := msg.Discord.Sess.State.Guild(msg.Message.GuildID)
	if err != nil {
		return
	}

	if g.Splash == "" {
		msg.Reply("this server has no banner")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title: g.Name,
		Color: 0xFFFFFF,
		Image: &discordgo.MessageEmbedImage{
			URL: fmt.Sprintf("https://cdn.discordapp.com/banners/%v/%v.png", g.ID, g.Splash),
		},
	}
	msg.ReplyEmbed(embed)
}

type ColorCommand struct {
	m       *UtilityMod
	Enabled bool
}

func NewColorCommand(m *UtilityMod) meidov2.ModCommand {
	return &ColorCommand{
		m:       m,
		Enabled: true,
	}
}
func (c *ColorCommand) Name() string {
	return "Color"
}
func (c *ColorCommand) Description() string {
	return "Displays a provided hex color"
}
func (c *ColorCommand) Triggers() []string {
	return []string{"m?color"}
}
func (c *ColorCommand) Usage() string {
	return "m?color #c0ffee\nm?color c0ffee"
}
func (c *ColorCommand) Cooldown() int {
	return 3
}
func (c *ColorCommand) RequiredPerms() int {
	return 0
}
func (c *ColorCommand) RequiresOwner() bool {
	return false
}
func (c *ColorCommand) IsEnabled() bool {
	return c.Enabled
}
func (c *ColorCommand) Run(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 2 || msg.Args()[0] != "m?color" {
		return
	}

	clrStr := msg.Args()[1]

	if clrStr[0] == '#' {
		clrStr = clrStr[1:]
	}

	clr, err := strconv.ParseInt(clrStr, 16, 32)
	if err != nil {
		msg.Reply("invalid color")
		return
	}
	if clr < 0 || clr > 0xffffff {
		msg.Reply("invalid color")
		return
	}

	red := clr >> 16
	green := (clr >> 8) & 0xff
	blue := clr & 0xff

	img := image.NewRGBA(image.Rect(0, 0, 64, 64))

	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			img.Set(x, y, color.RGBA{R: uint8(red), G: uint8(green), B: uint8(blue), A: 255})
		}
	}

	buf := &bytes.Buffer{}

	err = png.Encode(buf, img)
	if err != nil {
		return
	}

	msg.Sess.ChannelFileSend(msg.Message.ChannelID, "color.png", buf)
}
