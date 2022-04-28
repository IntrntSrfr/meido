package utilitymod

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/intrntsrfr/meido/base"
	"github.com/intrntsrfr/meido/database"
	"github.com/intrntsrfr/meido/utils"
)

type UtilityMod struct {
	sync.Mutex
	name         string
	commands     map[string]*base.ModCommand
	allowedTypes base.MessageType
	allowDMs     bool
	bot          *base.Bot
	startTime    time.Time
	db           *database.DB
}

func New(b *base.Bot, db *database.DB) base.Mod {
	return &UtilityMod{
		startTime:    time.Now(),
		name:         "Utility",
		commands:     make(map[string]*base.ModCommand),
		allowedTypes: base.MessageTypeCreate,
		allowDMs:     true,
		bot:          b,
		db:           db,
	}
}

func (m *UtilityMod) Name() string {
	return m.name
}
func (m *UtilityMod) Passives() []*base.ModPassive {
	return []*base.ModPassive{}
}
func (m *UtilityMod) Commands() map[string]*base.ModCommand {
	return m.commands
}
func (m *UtilityMod) AllowedTypes() base.MessageType {
	return m.allowedTypes
}
func (m *UtilityMod) AllowDMs() bool {
	return m.allowDMs
}
func (m *UtilityMod) Hook() error {
	m.bot.Discord.AddEventHandler(m.StatusLoop())

	m.RegisterCommand(NewPingCommand(m))
	m.RegisterCommand(NewAvatarCommand(m))
	m.RegisterCommand(NewBannerCommand(m))
	m.RegisterCommand(NewMemberAvatarCommand(m))
	m.RegisterCommand(NewAboutCommand(m))
	m.RegisterCommand(NewServerCommand(m))
	m.RegisterCommand(NewServerAvatarCommand(m))
	m.RegisterCommand(NewServerBannerCommand(m))
	m.RegisterCommand(NewServerSplashCommand(m))
	m.RegisterCommand(NewColorCommand(m))
	m.RegisterCommand(NewInviteCommand(m))
	//m.RegisterCommand(NewUserPermsCommand(m))
	m.RegisterCommand(NewUserInfoCommand(m))

	m.RegisterCommand(NewHelpCommand(m))

	return nil
}
func (m *UtilityMod) RegisterCommand(cmd *base.ModCommand) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.commands[cmd.Name]; ok {
		panic(fmt.Sprintf("command '%v' already exists in %v", cmd.Name, m.name))
	}
	m.commands[cmd.Name] = cmd
}

// NewPingCommand returns a new ping command.
func NewPingCommand(m *UtilityMod) *base.ModCommand {
	return &base.ModCommand{
		Mod:           m,
		Name:          "ping",
		Description:   "Checks the bot ping against Discord",
		Triggers:      []string{"m?ping"},
		Usage:         "m?ping",
		Cooldown:      2,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  base.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run:           m.pingCommand,
	}
}

func (m *UtilityMod) pingCommand(msg *base.DiscordMessage) {
	if msg.LenArgs() < 1 {
		return
	}

	startTime := time.Now()

	first, err := msg.Reply("Ping")
	if err != nil {
		return
	}

	now := time.Now()
	discordLatency := now.Sub(startTime)

	msg.Sess.ChannelMessageEdit(msg.Message.ChannelID, first.ID,
		fmt.Sprintf("Pong!\nDelay: %s", discordLatency))
}

func NewServerCommand(m *UtilityMod) *base.ModCommand {
	return &base.ModCommand{
		Mod:           m,
		Name:          "server",
		Description:   "Displays information about the server",
		Triggers:      []string{"m?server"},
		Usage:         "m?server",
		Cooldown:      5,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  base.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           m.serverCommand,
	}
}

func (m *UtilityMod) serverCommand(msg *base.DiscordMessage) {
	if msg.LenArgs() < 1 {
		return
	}

	g, err := msg.Discord.Guild(msg.Message.GuildID)
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

	users := 0
	bots := 0

	for _, mem := range g.Members {
		if mem.User.Bot {
			bots++
		} else {
			users++
		}
	}

	owner, err := msg.Discord.Member(g.ID, g.OwnerID)
	if err != nil {
		msg.Reply("Error getting guild data")
		return
	}

	ts := utils.IDToTimestamp(g.ID)
	dur := time.Since(ts)

	embed := discordgo.MessageEmbed{
		Color: utils.ColorInfo,
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
				Value: fmt.Sprintf("%v | %v day(s) ago", ts.Format(time.RFC1123), math.Floor(dur.Hours()/24.0)),
			},
			{
				Name:   "Members",
				Value:  fmt.Sprintf("%v members\n%v users\n%v bots", g.MemberCount, users, bots),
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
	if hash := g.IconURL(); hash != "" {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: fmt.Sprintf("%v?size=1024", hash),
		}
		embed.Author.IconURL = fmt.Sprintf("%v?size=256", hash)
	}

	msg.ReplyEmbed(&embed)
}

func NewAboutCommand(m *UtilityMod) *base.ModCommand {
	return &base.ModCommand{
		Mod:           m,
		Name:          "about",
		Description:   "Displays Meido statistics",
		Triggers:      []string{"m?about"},
		Usage:         "m?about",
		Cooldown:      5,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowDMs:      true,
		AllowedTypes:  base.MessageTypeCreate,
		Enabled:       true,
		Run:           m.aboutCommand,
	}
}

func (m *UtilityMod) aboutCommand(msg *base.DiscordMessage) {
	if msg.LenArgs() < 1 {
		return
	}

	var (
		totalUsers  int
		totalBots   int
		totalHumans int
		memory      runtime.MemStats
	)
	runtime.ReadMemStats(&memory)
	guilds := msg.Discord.Guilds()
	for _, guild := range guilds {
		for _, mem := range guild.Members {
			if mem.User.Bot {
				totalBots++
			} else {
				totalHumans++
			}
		}

		totalUsers += guild.MemberCount
	}

	uptime := time.Now().Sub(m.startTime)
	count, err := m.db.GetCommandCount()
	if err != nil {
		return
	}

	msg.ReplyEmbed(&discordgo.MessageEmbed{
		Title: "About",
		Color: utils.ColorInfo,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Uptime",
				Value:  uptime.String(),
				Inline: true,
			},
			{
				Name:   "Total commands ran",
				Value:  strconv.Itoa(count),
				Inline: true,
			},
			{
				Name:   "Guilds",
				Value:  strconv.Itoa(len(guilds)),
				Inline: false,
			},
			{
				Name:   "Users",
				Value:  fmt.Sprintf("%v users | %v humans | %v bots", totalUsers, totalHumans, totalBots),
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
		},
	})
}

func NewServerSplashCommand(m *UtilityMod) *base.ModCommand {
	return &base.ModCommand{
		Mod:           m,
		Name:          "serversplash",
		Description:   "Displays server splash if one exists",
		Triggers:      []string{"m?serversplash"},
		Usage:         "m?serversplash",
		Cooldown:      5,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  base.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           m.serverSplashCommand,
	}
}
func (m *UtilityMod) serverSplashCommand(msg *base.DiscordMessage) {
	if msg.LenArgs() < 1 {
		return
	}

	g, err := msg.Discord.Guild(msg.Message.GuildID)
	if err != nil {
		return
	}

	if g.Splash == "" {
		msg.Reply("This server doesn't have a splash!")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title: g.Name,
		Color: utils.ColorInfo,
		Image: &discordgo.MessageEmbedImage{
			URL: fmt.Sprintf("%v?size=2048", discordgo.EndpointGuildSplash(g.ID, g.Splash)),
		},
	}
	msg.ReplyEmbed(embed)
}
func NewServerAvatarCommand(m *UtilityMod) *base.ModCommand {
	return &base.ModCommand{
		Mod:           m,
		Name:          "serveravatar",
		Description:   "Displays server avatar if one exists",
		Triggers:      []string{"m?serveravatar", "m?servericon", "m?sav", ">sav"},
		Usage:         "m?servericon",
		Cooldown:      5,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  base.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           m.serverIconCommand,
	}
}
func (m *UtilityMod) serverIconCommand(msg *base.DiscordMessage) {
	if msg.LenArgs() < 1 {
		return
	}

	g, err := msg.Discord.Guild(msg.Message.GuildID)
	if err != nil {
		return
	}

	if g.Icon == "" {
		msg.Reply("This server doesn't have an icon!")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title: g.Name,
		Color: utils.ColorInfo,
		Image: &discordgo.MessageEmbedImage{
			URL: fmt.Sprintf("%v?size=2048", g.IconURL()),
		},
	}
	msg.ReplyEmbed(embed)
}

func NewServerBannerCommand(m *UtilityMod) *base.ModCommand {
	return &base.ModCommand{
		Mod:           m,
		Name:          "serverbanner",
		Description:   "Displays server banner if one exists",
		Triggers:      []string{"m?serverbanner"},
		Usage:         "m?serverbanner",
		Cooldown:      5,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  base.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           m.serverBannerCommand,
	}
}
func (m *UtilityMod) serverBannerCommand(msg *base.DiscordMessage) {
	if msg.LenArgs() < 1 {
		return
	}

	g, err := msg.Discord.Guild(msg.Message.GuildID)
	if err != nil {
		return
	}

	hash := g.BannerURL()
	if hash == "" {
		msg.Reply("This server doesn't have a banner!")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title: g.Name,
		Color: utils.ColorInfo,
		Image: &discordgo.MessageEmbedImage{
			URL: fmt.Sprintf("%v?size=2048", hash),
		},
	}
	msg.ReplyEmbed(embed)
}

func NewColorCommand(m *UtilityMod) *base.ModCommand {
	return &base.ModCommand{
		Mod:           m,
		Name:          "color",
		Description:   "Displays a hex color",
		Triggers:      []string{"m?color"},
		Usage:         "m?color [hex color]",
		Cooldown:      1,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  base.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run:           m.colorCommand,
	}
}
func (m *UtilityMod) colorCommand(msg *base.DiscordMessage) {
	if msg.LenArgs() < 2 {
		return
	}

	clrStr := msg.Args()[1]

	if clrStr[0] == '#' {
		clrStr = clrStr[1:]
	}

	clr, err := strconv.ParseInt(clrStr, 16, 32)
	if err != nil || clr < 0 || clr > 0xffffff {
		msg.Reply("invalid color")
		return
	}

	red := clr >> 16
	green := (clr >> 8) & 0xff
	blue := clr & 0xff

	img := image.NewRGBA(image.Rect(0, 0, 64, 64))

	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.RGBA{R: uint8(red), G: uint8(green), B: uint8(blue), A: 255}}, image.Point{}, draw.Src)

	buf := bytes.Buffer{}

	err = png.Encode(&buf, img)
	if err != nil {
		return
	}

	msg.Sess.ChannelFileSend(msg.Message.ChannelID, "color.png", &buf)
}

func NewInviteCommand(m *UtilityMod) *base.ModCommand {
	return &base.ModCommand{
		Mod:           m,
		Name:          "invite",
		Description:   "Sends an invite link for Meido, as well as support server",
		Triggers:      []string{"m?invite"},
		Usage:         "m?invite",
		Cooldown:      1,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  base.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run:           m.inviteCommand,
	}
}
func (m *UtilityMod) inviteCommand(msg *base.DiscordMessage) {
	botLink := "<https://discordapp.com/oauth2/authorize?client_id=" + m.bot.Discord.Sess.State.User.ID + "&scope=bot>"
	serverLink := "https://discord.gg/KgMEGK3"
	msg.Reply(fmt.Sprintf("Invite me to your server: %v\nSupport server: %v", botLink, serverLink))
}

func NewHelpCommand(m *UtilityMod) *base.ModCommand {
	return &base.ModCommand{
		Mod:           m,
		Name:          "help",
		Description:   "Displays helpful things",
		Triggers:      []string{"m?help", "m?h"},
		Usage:         "m?help | m?help about",
		Cooldown:      1,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  base.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run:           m.helpCommand,
	}
}

func (m *UtilityMod) helpCommand(msg *base.DiscordMessage) {

	emb := &discordgo.MessageEmbed{
		Color: utils.ColorInfo,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Use m?help [plugin] to see plugin commands.\nUse m?help [command] to see command info.",
		},
	}
	switch msg.LenArgs() {
	case 1:
		desc := strings.Builder{}
		for _, mod := range m.bot.Mods {
			desc.WriteString(fmt.Sprintf("- %v\n", mod.Name()))
		}
		emb.Title = "Meido plugins"
		emb.Description = desc.String()
		msg.ReplyEmbed(emb)
	case 2:

		inp := strings.ToLower(msg.Args()[1])

		for _, mod := range m.bot.Mods {
			if strings.ToLower(mod.Name()) == strings.ToLower(inp) {
				// this can maybe be replaced by making a helptext method for every mod, so they have more control
				// over what they want to display, if they even want to display anything.

				list := strings.Builder{}

				list.WriteString("\nPassives:\n")
				for _, pas := range mod.Passives() {
					list.WriteString(fmt.Sprintf("- %v\n", pas.Name))
				}

				list.WriteString("\nCommands:\n")
				for _, cmd := range mod.Commands() {
					list.WriteString(fmt.Sprintf("- %v\n", cmd.Name))
				}
				dms := "Does not work in DMs."
				if mod.AllowDMs() {
					dms = "Works in DMs."
				}
				list.WriteString(dms)

				emb.Title = fmt.Sprintf("Commands for %v plugin", mod.Name())
				emb.Description = list.String()

				msg.ReplyEmbed(emb)
				return
			}

			for _, pas := range mod.Passives() {
				if strings.ToLower(pas.Name) == strings.ToLower(inp) {

					emb.Title = fmt.Sprintf("Passive - %v", pas.Name)
					emb.Description = pas.Description + "\n"
					msg.ReplyEmbed(emb)

					return
				}
			}
			for _, cmd := range mod.Commands() {
				isCmd := false
				if strings.ToLower(cmd.Name) == strings.ToLower(inp) {
					isCmd = true
				}

				for _, trig := range cmd.Triggers {
					if strings.ToLower(trig) == strings.ToLower(inp) {
						isCmd = true
					}
				}

				if !isCmd {
					continue
				}

				emb.Title = fmt.Sprintf("Command - %v", cmd.Name)

				dmText := map[bool]string{true: "This works in DMs", false: "This does not work in DMs"}

				info := strings.Builder{}
				info.WriteString(fmt.Sprintf("\n\n%v", cmd.Description))
				info.WriteString(fmt.Sprintf("\n\nAliases: %v", strings.Join(cmd.Triggers, ", ")))
				info.WriteString(fmt.Sprintf("\n\nUsage: %v", cmd.Usage))
				info.WriteString(fmt.Sprintf("\n\nCooldown: %v second(s)", cmd.Cooldown))
				info.WriteString(fmt.Sprintf("\n\nRequired permissions: %v", base.PermMap[cmd.RequiredPerms]))
				info.WriteString(fmt.Sprintf("\n\n%v", dmText[cmd.AllowDMs]))
				emb.Description = info.String()

				msg.ReplyEmbed(emb)

				return
			}
		}
	default:
		return
	}
}
