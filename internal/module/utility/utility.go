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
	iutils "github.com/intrntsrfr/meido/internal/utils"
	"github.com/intrntsrfr/meido/pkg/mio/bot"
	"github.com/intrntsrfr/meido/pkg/mio/discord"
	"github.com/intrntsrfr/meido/pkg/utils"
	"go.uber.org/zap"
)

type Module struct {
	*bot.ModuleBase
	db        database.DB
	startTime time.Time
}

func New(b *bot.Bot, db database.DB, logger *zap.Logger) bot.Module {
	logger = logger.Named("Utility")
	return &Module{
		ModuleBase: bot.NewModule(b, "Utility", logger),
		db:         db,
		startTime:  time.Now(),
	}
}

func (m *Module) Hook() error {
	return m.RegisterCommands(
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
	)
}

func NewConvertCommand(m *Module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "convert",
		Description:      "Converts between units",
		Triggers:         []string{"m?convert"},
		Usage:            "m?convert kg lb 50",
		Cooldown:         0,
		CooldownScope:    bot.Channel,
		RequiredPerms:    0,
		CheckBotPerms:    false,
		RequiresUserType: bot.UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         true,
		Enabled:          true,
		Run: func(msg *discord.DiscordMessage) {
			if len(msg.Args()) < 4 {
				return
			}
		},
	}
}

// newPingCommand returns a new ping command.
func newPingCommand(m *Module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "ping",
		Description:      "Checks the bot ping against Discord",
		Triggers:         []string{"m?ping"},
		Usage:            "m?ping",
		Cooldown:         2,
		CooldownScope:    bot.Channel,
		RequiredPerms:    0,
		CheckBotPerms:    false,
		RequiresUserType: bot.UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         true,
		Enabled:          true,
		Run: func(msg *discord.DiscordMessage) {
			if len(msg.Args()) < 1 {
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

func newAboutCommand(m *Module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "about",
		Description:      "Displays Meido statistics",
		Triggers:         []string{"m?about"},
		Usage:            "m?about",
		Cooldown:         5,
		CooldownScope:    bot.Channel,
		RequiredPerms:    0,
		CheckBotPerms:    false,
		RequiresUserType: bot.UserTypeAny,
		AllowDMs:         true,
		AllowedTypes:     discord.MessageTypeCreate,
		Enabled:          true,
		Run: func(msg *discord.DiscordMessage) {
			if len(msg.Args()) < 1 {
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
			embed := iutils.NewEmbed().
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

func newColorCommand(m *Module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "color",
		Description:      "Displays a small image of a provided color hex",
		Triggers:         []string{"m?color"},
		Usage:            "m?color [color hex]",
		Cooldown:         1,
		CooldownScope:    bot.Channel,
		RequiredPerms:    0,
		CheckBotPerms:    false,
		RequiresUserType: bot.UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         true,
		Enabled:          true,
		Run: func(msg *discord.DiscordMessage) {
			if len(msg.Args()) < 2 {
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

func newIdTimestampCmd(m *Module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "idtimestamp",
		Description:      "Converts a Discord ID to a timestamp",
		Triggers:         []string{"m?idt", "m?idts", "m?ts", "m?idtimestamp"},
		Usage:            "m?idt [ID]",
		Cooldown:         0,
		CooldownScope:    bot.Channel,
		RequiredPerms:    0,
		CheckBotPerms:    false,
		RequiresUserType: bot.UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         true,
		Enabled:          true,
		Run: func(msg *discord.DiscordMessage) {
			id := msg.AuthorID()
			if len(msg.Args()) > 1 {
				id = msg.Args()[1]
			}
			_, _ = msg.Reply(fmt.Sprintf("<t:%v>", utils.IDToTimestamp(id).Unix()))
		},
	}

}

func newInviteCommand(m *Module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "invite",
		Description:      "Sends a bot invite link and support server invite link",
		Triggers:         []string{"m?invite"},
		Usage:            "m?invite",
		Cooldown:         1,
		CooldownScope:    bot.Channel,
		RequiredPerms:    0,
		CheckBotPerms:    false,
		RequiresUserType: bot.UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         true,
		Enabled:          true,
		Run: func(msg *discord.DiscordMessage) {
			botLink := "<https://discordapp.com/oauth2/authorize?client_id=" + msg.Sess.State().User.ID + "&scope=bot>"
			serverLink := "https://discord.gg/KgMEGK3"
			_, _ = msg.Reply(fmt.Sprintf("Invite me to your server: %v\nSupport server: %v", botLink, serverLink))
		},
	}
}

func newHelpCommand(m *Module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "help",
		Description:      "Displays helpful things",
		Triggers:         []string{"m?help", "m?h"},
		Usage:            "m?help <module | command | passive>",
		Cooldown:         1,
		CooldownScope:    bot.Channel,
		RequiredPerms:    0,
		CheckBotPerms:    false,
		RequiresUserType: bot.UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         true,
		Enabled:          true,
		Run:              m.helpCommand,
	}
}

func (m *Module) helpCommand(msg *discord.DiscordMessage) {
	embed := iutils.NewEmbed().
		WithOkColor().
		WithFooter("Use m?help [module] to see module commands.\nUse m?help [command] to see command info.\nArguments in [square brackets] are required, while arguments in <angle brackets> are optional.", "").
		WithThumbnail(msg.Sess.State().User.AvatarURL("256"))

	if len(msg.Args()) == 1 {
		desc := strings.Builder{}
		for _, mod := range m.Bot.Modules {
			desc.WriteString(fmt.Sprintf("- %v\n", mod.Name()))
		}
		embed.WithTitle("Modules")
		embed.WithDescription(desc.String())
		_, _ = msg.ReplyEmbed(embed.Build())
		return
	}

	// if only m?help
	if len(msg.Args()) < 2 {
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
				list.WriteString(fmt.Sprintf("- `%v`\n", pas.Name))
			}
		}

		if len(mod.Commands()) > 0 {
			list.WriteString("\nCommands:\n")
			for _, cmd := range mod.Commands() {
				list.WriteString(fmt.Sprintf("- `%v`\n", cmd.Name))
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
		info.WriteString(fmt.Sprintf("\n**Required permissions**: %v", discord.PermMap[cmd.RequiredPerms]))
		if !cmd.AllowDMs {
			info.WriteString(fmt.Sprintf("\n%v", "Cannot be used in DMs"))
		}

		embed.WithTitle(fmt.Sprintf("Command - %v", cmd.Name))
		embed.WithDescription(info.String())
		_, _ = msg.ReplyEmbed(embed.Build())
		return
	}
}
