package utilitymod

import (
	"bytes"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/intrntsrfr/meido/internal/database"
	"github.com/intrntsrfr/meido/internal/mods"
	"github.com/intrntsrfr/meido/pkg/mio"
	"github.com/intrntsrfr/meido/pkg/utils"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

type UtilityMod struct {
	sync.Mutex
	name         string
	commands     map[string]*mio.ModCommand
	allowedTypes mio.MessageType
	allowDMs     bool
	bot          *mio.Bot
	startTime    time.Time
	db           database.DB
}

func New(b *mio.Bot, db database.DB) mio.Mod {
	return &UtilityMod{
		startTime:    time.Now(),
		name:         "Utility",
		commands:     make(map[string]*mio.ModCommand),
		allowedTypes: mio.MessageTypeCreate,
		allowDMs:     true,
		bot:          b,
		db:           db,
	}
}

func (m *UtilityMod) Name() string {
	return m.name
}
func (m *UtilityMod) Passives() []*mio.ModPassive {
	return []*mio.ModPassive{}
}
func (m *UtilityMod) Commands() map[string]*mio.ModCommand {
	return m.commands
}
func (m *UtilityMod) AllowedTypes() mio.MessageType {
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
	m.RegisterCommand(NewServerIconCommand(m))
	m.RegisterCommand(NewServerBannerCommand(m))
	m.RegisterCommand(NewServerSplashCommand(m))
	m.RegisterCommand(NewColorCommand(m))
	m.RegisterCommand(NewInviteCommand(m))
	//m.RegisterCommand(NewUserPermsCommand(m))
	m.RegisterCommand(NewUserInfoCommand(m))

	m.RegisterCommand(NewHelpCommand(m))

	return nil
}
func (m *UtilityMod) RegisterCommand(cmd *mio.ModCommand) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.commands[cmd.Name]; ok {
		panic(fmt.Sprintf("command '%v' already exists in %v", cmd.Name, m.name))
	}
	m.commands[cmd.Name] = cmd
}

func NewWeatherCommand(m *UtilityMod) *mio.ModCommand {
	return &mio.ModCommand{
		Mod:           m,
		Name:          "weather",
		Description:   "Finds the weather at a provided location",
		Triggers:      []string{"m?weather"},
		Usage:         "m?weather Oslo",
		Cooldown:      0,
		CooldownUser:  false,
		RequiredPerms: 0,
		RequiresOwner: false,
		CheckBotPerms: false,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run: func(msg *mio.DiscordMessage) {
			// utilize open weather api?
		},
	}
}

func NewConvertCommand(m *UtilityMod) *mio.ModCommand {
	return &mio.ModCommand{
		Mod:           m,
		Name:          "convert",
		Description:   "Converts between units",
		Triggers:      []string{"m?convert"},
		Usage:         "m?convert kg lb 50",
		Cooldown:      0,
		CooldownUser:  false,
		RequiredPerms: 0,
		RequiresOwner: false,
		CheckBotPerms: false,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run: func(msg *mio.DiscordMessage) {
			if msg.LenArgs() < 4 {
				return
			}

		},
	}
}

// NewPingCommand returns a new ping command.
func NewPingCommand(m *UtilityMod) *mio.ModCommand {
	return &mio.ModCommand{
		Mod:           m,
		Name:          "ping",
		Description:   "Checks the bot ping against Discord",
		Triggers:      []string{"m?ping"},
		Usage:         "m?ping",
		Cooldown:      2,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run:           m.pingCommand,
	}
}

func (m *UtilityMod) pingCommand(msg *mio.DiscordMessage) {
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
	_, _ = msg.Sess.ChannelMessageEdit(msg.Message.ChannelID, first.ID,
		fmt.Sprintf("Pong!\nDelay: %s", discordLatency))
}

func NewAboutCommand(m *UtilityMod) *mio.ModCommand {
	return &mio.ModCommand{
		Mod:           m,
		Name:          "about",
		Description:   "Displays Meido statistics",
		Triggers:      []string{"m?about"},
		Usage:         "m?about",
		Cooldown:      5,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowDMs:      true,
		AllowedTypes:  mio.MessageTypeCreate,
		Enabled:       true,
		Run: func(msg *mio.DiscordMessage) {
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

			embed := &discordgo.MessageEmbed{}
			embed = mods.SetEmbedTitle(embed, "About")
			embed.Color = utils.ColorInfo
			embed = mods.AddEmbedField(embed, "Uptime", uptime.String(), true)
			embed = mods.AddEmbedField(embed, "Total commands ran", fmt.Sprint(count), true)
			embed = mods.AddEmbedField(embed, "Guilds", fmt.Sprint(len(guilds)), false)
			embed = mods.AddEmbedField(embed, "Users", fmt.Sprintf("%v users | %v humans | %v bots", totalUsers, totalHumans, totalBots), true)
			embed = mods.AddEmbedField(embed, "Memory use", fmt.Sprintf("%v/%v", humanize.Bytes(memory.Alloc), humanize.Bytes(memory.Sys)), false)
			embed = mods.AddEmbedField(embed, "Garbage collected", humanize.Bytes(memory.TotalAlloc-memory.Alloc), true)
			_, _ = msg.ReplyEmbed(embed)
		},
	}
}

func NewColorCommand(m *UtilityMod) *mio.ModCommand {
	return &mio.ModCommand{
		Mod:           m,
		Name:          "color",
		Description:   "Displays a hex color",
		Triggers:      []string{"m?color"},
		Usage:         "m?color [hex color]",
		Cooldown:      1,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run:           m.colorCommand,
	}
}
func (m *UtilityMod) colorCommand(msg *mio.DiscordMessage) {
	if msg.LenArgs() < 2 {
		return
	}

	clrStr := msg.Args()[1]

	if clrStr[0] == '#' {
		clrStr = clrStr[1:]
	}

	clr, err := strconv.ParseInt(clrStr, 16, 32)
	if err != nil || clr < 0 || clr > 0xffffff {
		_, _ = msg.Reply("invalid color")
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
	_, _ = msg.Sess.ChannelFileSend(msg.Message.ChannelID, "color.png", &buf)
}

func NewInviteCommand(m *UtilityMod) *mio.ModCommand {
	return &mio.ModCommand{
		Mod:           m,
		Name:          "invite",
		Description:   "Sends an invite link for Meido, as well as support server",
		Triggers:      []string{"m?invite"},
		Usage:         "m?invite",
		Cooldown:      1,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run: func(msg *mio.DiscordMessage) {
			botLink := "<https://discordapp.com/oauth2/authorize?client_id=" + m.bot.Discord.Sess.State.User.ID + "&scope=bot>"
			serverLink := "https://discord.gg/KgMEGK3"
			_, _ = msg.Reply(fmt.Sprintf("Invite me to your server: %v\nSupport server: %v", botLink, serverLink))
		},
	}
}

func NewHelpCommand(m *UtilityMod) *mio.ModCommand {
	return &mio.ModCommand{
		Mod:           m,
		Name:          "help",
		Description:   "Displays helpful things",
		Triggers:      []string{"m?help", "m?h"},
		Usage:         "m?help | m?help about",
		Cooldown:      1,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run:           m.helpCommand,
	}
}

func (m *UtilityMod) helpCommand(msg *mio.DiscordMessage) {

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
				info.WriteString(fmt.Sprintf("\n\nRequired permissions: %v", mio.PermMap[cmd.RequiredPerms]))
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
