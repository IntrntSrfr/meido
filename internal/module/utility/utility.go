package utility

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/intrntsrfr/meido/internal/database"
	"github.com/intrntsrfr/meido/internal/helpers"
	"github.com/intrntsrfr/meido/pkg/mio"
	"github.com/intrntsrfr/meido/pkg/utils"
	"go.uber.org/zap"
)

type Module struct {
	*mio.ModuleBase
	db        database.DB
	startTime time.Time
}

func New(bot *mio.Bot, db database.DB, logger *zap.Logger) mio.Module {
	return &Module{
		ModuleBase: mio.NewModule(bot, "Utility", logger.Named("utility")),
		db:         db,
		startTime:  time.Now(),
	}
}

func (m *Module) Hook() error {
	return m.RegisterCommands([]*mio.ModuleCommand{
		newPingCommand(m),
		newAvatarCommand(m),
		newBannerCommand(m),
		newMemberAvatarCommand(m),
		newAboutCommand(m),
		newServerCommand(m),
		newServerIconCommand(m),
		newServerBannerCommand(m),
		newServerSplashCommand(m),
		newColorCommand(m),
		newIdTimestampCmd(m),
		newInviteCommand(m),
		newUserInfoCommand(m),
		newHelpCommand(m),
	})
}

func NewConvertCommand(m *Module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
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
		IsEnabled:     true,
		Run: func(msg *mio.DiscordMessage) {
			if msg.LenArgs() < 4 {
				return
			}
		},
	}
}

// newPingCommand returns a new ping command.
func newPingCommand(m *Module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
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
		IsEnabled:     true,
		Run: func(msg *mio.DiscordMessage) {
			if msg.LenArgs() < 1 {
				return
			}
			startTime := time.Now()
			first, err := msg.Reply("Ping")
			if err != nil {
				return
			}
			_, _ = msg.Sess.ChannelMessageEdit(msg.Message.ChannelID, first.ID,
				fmt.Sprintf("Pong!\nDelay: %s", time.Since(startTime)))
		},
	}
}

func newAboutCommand(m *Module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
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
		IsEnabled:     true,
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

			uptime := time.Since(m.startTime)
			count, err := m.db.GetCommandCount()
			if err != nil {
				return
			}
			embed := helpers.NewEmbed().
				WithTitle("About").
				WithOkColor().
				AddField("Uptime", uptime.String(), true).
				AddField("Total commands ran", fmt.Sprint(count), true).
				AddField("Guilds", fmt.Sprint(len(guilds)), false).
				AddField("Users", fmt.Sprintf("%v users | %v humans | %v bots", totalUsers, totalHumans, totalBots), true).
				AddField("Memory use", fmt.Sprintf("%v/%v", humanize.Bytes(memory.Alloc), humanize.Bytes(memory.Sys)), false).
				AddField("Garbage collected", humanize.Bytes(memory.TotalAlloc-memory.Alloc), true).
				AddField("Owners", strings.Join(m.Bot.Config.GetStringSlice("owner_ids"), ", "), true)
			_, _ = msg.ReplyEmbed(embed.Build())
		},
	}
}

func newColorCommand(m *Module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:           m,
		Name:          "color",
		Description:   "Displays a small image of a provided color hex",
		Triggers:      []string{"m?color"},
		Usage:         "m?color [color hex]",
		Cooldown:      1,
		CooldownUser:  false,
		RequiredPerms: 0,
		RequiresOwner: false,
		CheckBotPerms: false,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      true,
		IsEnabled:     true,
		Run: func(msg *mio.DiscordMessage) {
			if msg.LenArgs() < 2 {
				return
			}

			colorStr := msg.Args()[1]
			colorStr = strings.TrimPrefix(colorStr, "#")
			buf, err := generateColorPNG(colorStr)
			if err != nil {
				_, _ = msg.Reply("Invalid hex code")
				return
			}
			_, _ = msg.ReplyComplex(&discordgo.MessageSend{File: &discordgo.File{Name: "color.png", Reader: buf}})
		},
	}
}

func generateColorPNG(clrStr string) (*bytes.Buffer, error) {
	clr, err := strconv.ParseInt(clrStr, 16, 32)
	if err != nil || clr < 0 || clr > 0xffffff {
		return nil, err
	}

	red := clr >> 16
	green := (clr >> 8) & 0xff
	blue := clr & 0xff

	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.RGBA{R: uint8(red), G: uint8(green), B: uint8(blue), A: 255}}, image.Point{}, draw.Src)
	buf := bytes.Buffer{}
	err = png.Encode(&buf, img)
	return &buf, err
}

func newIdTimestampCmd(m *Module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:           m,
		Name:          "idtimestamp",
		Description:   "Converts a Discord ID to a timestamp",
		Triggers:      []string{"m?idt", "m?idts", "m?ts", "m?idtimestamp"},
		Usage:         "m?idt [ID]",
		Cooldown:      0,
		CooldownUser:  false,
		RequiredPerms: 0,
		RequiresOwner: false,
		CheckBotPerms: false,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      true,
		IsEnabled:     true,
		Run: func(msg *mio.DiscordMessage) {
			id := msg.AuthorID()
			if msg.LenArgs() > 1 {
				id = msg.Args()[1]
			}
			ts, err := utils.IDToTimestamp2(id)
			if err != nil {
				return
			}
			_, _ = msg.Reply(fmt.Sprintf("<t:%v>", ts.Unix()))
		},
	}

}

func newInviteCommand(m *Module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:           m,
		Name:          "invite",
		Description:   "Sends a bot invite link and support server invite link",
		Triggers:      []string{"m?invite"},
		Usage:         "m?invite",
		Cooldown:      1,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      true,
		IsEnabled:     true,
		Run: func(msg *mio.DiscordMessage) {
			botLink := "<https://discordapp.com/oauth2/authorize?client_id=" + msg.Sess.State.User.ID + "&scope=bot>"
			serverLink := "https://discord.gg/KgMEGK3"
			_, _ = msg.Reply(fmt.Sprintf("Invite me to your server: %v\nSupport server: %v", botLink, serverLink))
		},
	}
}

func newHelpCommand(m *Module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:           m,
		Name:          "help",
		Description:   "Displays helpful things",
		Triggers:      []string{"m?help", "m?h"},
		Usage:         "m?help <module | command | passive>",
		Cooldown:      1,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      true,
		IsEnabled:     true,
		Run:           m.helpCommand,
	}
}

func (m *Module) helpCommand(msg *mio.DiscordMessage) {
	embed := helpers.NewEmbed().
		WithOkColor().
		WithFooter("Use m?help [module] to see module commands.\nUse m?help [command] to see command info.\nArguments in [square brackets] are required, while arguments in <angle brackets> are optional.", "").
		WithThumbnail(msg.Sess.State.User.AvatarURL("256"))

	if msg.LenArgs() == 1 {
		desc := strings.Builder{}
		for _, mod := range m.Bot.Modules {
			desc.WriteString(fmt.Sprintf("- %v\n", mod.Name()))
		}
		embed.WithTitle("Plugins")
		embed.WithDescription(desc.String())
		_, _ = msg.ReplyEmbed(embed.Build())
		return
	}

	// if only m?help
	if msg.LenArgs() < 2 {
		return
	}

	inp := strings.Join(msg.Args()[1:], "")
	if mod, err := m.Bot.FindModule(inp); err == nil {
		// this can maybe be replaced by making a helptext method for every mod, so they have more control
		// over what they want to display, if they even want to display anything.
		list := strings.Builder{}
		if len(mod.Passives()) > 0 {
			list.WriteString("\nPassives:\n")
			for _, pas := range mod.Passives() {
				list.WriteString(fmt.Sprintf("- %v\n", pas.Name))
			}
		}

		if len(mod.Commands()) > 0 {
			list.WriteString("\nCommands:\n")
			for _, cmd := range mod.Commands() {
				list.WriteString(fmt.Sprintf("- %v\n", cmd.Name))
			}
		}

		if !mod.AllowDMs() {
			list.WriteString("\nCannot be used in DMs")
		}

		embed.WithTitle(fmt.Sprintf("%v module", mod.Name()))
		embed.WithDescription(list.String())
		_, _ = msg.ReplyEmbed(embed.Build())
		return
	}

	if pas, err := m.Bot.FindPassive(inp); err == nil {
		embed.WithTitle(fmt.Sprintf("Passive - %v", pas.Name))
		embed.WithDescription(fmt.Sprintf("%v\n", pas.Description))
		_, _ = msg.ReplyEmbed(embed.Build())
		return
	}

	if cmd, err := m.Bot.FindCommand(inp); err == nil {
		info := strings.Builder{}
		info.WriteString(fmt.Sprintf("%v\n", cmd.Description))
		info.WriteString(fmt.Sprintf("\n**Usage**: %v", cmd.Usage))
		info.WriteString(fmt.Sprintf("\n**Aliases**: %v", strings.Join(cmd.Triggers, ", ")))
		info.WriteString(fmt.Sprintf("\n**Cooldown**: %v second(s)", cmd.Cooldown))
		info.WriteString(fmt.Sprintf("\n**Required permissions**: %v", mio.PermMap[cmd.RequiredPerms]))
		if !cmd.AllowDMs {
			info.WriteString(fmt.Sprintf("\n%v", "Cannot be used in DMs"))
		}

		embed.WithTitle(fmt.Sprintf("Command - %v", cmd.Name))
		embed.WithDescription(info.String())
		_, _ = msg.ReplyEmbed(embed.Build())
		return
	}
}
