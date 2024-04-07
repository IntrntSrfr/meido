package utility

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"math/rand"
	"runtime"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/g4s8/hexcolor"
	"github.com/intrntsrfr/meido/internal/database"
	"github.com/intrntsrfr/meido/pkg/mio"
	"github.com/intrntsrfr/meido/pkg/mio/bot"
	"github.com/intrntsrfr/meido/pkg/mio/discord"
	"github.com/intrntsrfr/meido/pkg/utils"
	"github.com/intrntsrfr/meido/pkg/utils/builders"
)

type module struct {
	*bot.ModuleBase
	db        database.DB
	startTime time.Time
}

func New(b *bot.Bot, db database.DB, logger mio.Logger) bot.Module {
	logger = logger.Named("Utility")
	return &module{
		ModuleBase: bot.NewModule(b, "Utility", logger),
		db:         db,
		startTime:  time.Now(),
	}
}

func (m *module) Hook() error {
	if err := m.RegisterApplicationCommands(
		newColorSlash(m),
		newUserInfoUserCommand(m),
		newHelpSlash(m),
		newPingSlash(m),
		newStatsSlash(m),
	); err != nil {
		return err
	}

	if err := m.RegisterMessageComponents(
		newHelpModuleSelectHandler(m),
		newHelpCommandSelectHandler(m),
		newHelpModuleBackHandler(m),
		newHelpCommandBackHandler(m),
	); err != nil {
		return err
	}

	if err := m.RegisterCommands(
		newPingCommand(m),
		newAvatarCommand(m),
		newBannerCommand(m),
		newMemberAvatarCommand(m),
		newServerCommand(m),
		newServerIconCommand(m),
		newServerBannerCommand(m),
		newServerSplashCommand(m),
		newColorCommand(m),
		newIdTimestampCmd(m),
		newInviteCommand(m),
		newUserInfoCommand(m),
		newHelpCommand(m),
	); err != nil {
		return err
	}

	return nil
}

func NewConvertCommand(m *module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "convert",
		Description:      "Converts between units",
		Triggers:         []string{"m?convert"},
		Usage:            "m?convert kg lb 50",
		Cooldown:         0,
		CooldownScope:    bot.CooldownScopeChannel,
		RequiredPerms:    0,
		CheckBotPerms:    false,
		RequiresUserType: bot.UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         true,
		Enabled:          true,
		Execute: func(msg *discord.DiscordMessage) {
			if len(msg.Args()) < 4 {
				return
			}
		},
	}
}

// newPingCommand returns a new ping command.
func newPingCommand(m *module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "ping",
		Description:      "Checks how fast the bot can respond to a command",
		Triggers:         []string{"m?ping"},
		Usage:            "m?ping",
		Cooldown:         time.Second * 2,
		CooldownScope:    bot.CooldownScopeChannel,
		RequiredPerms:    0,
		CheckBotPerms:    false,
		RequiresUserType: bot.UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         true,
		Enabled:          true,
		Execute: func(msg *discord.DiscordMessage) {
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

func newPingSlash(m *module) *bot.ModuleApplicationCommand {
	bld := bot.NewModuleApplicationCommandBuilder(m, "ping").
		Type(discordgo.ChatApplicationCommand).
		Description("Checks how fast the bot can respond to a command")

	run := func(dac *discord.DiscordApplicationCommand) {
		startTime := time.Now()
		dac.Respond("Pong")
		resp := fmt.Sprintf("Pong!\nDelay: %s", time.Since(startTime))
		respData := &discordgo.WebhookEdit{
			Content: &resp,
		}
		dac.Sess.Real().InteractionResponseEdit(dac.Interaction, respData)
	}

	return bld.Execute(run).Build()
}

func newStatsSlash(m *module) *bot.ModuleApplicationCommand {
	bld := bot.NewModuleApplicationCommandBuilder(m, "stats").
		Type(discordgo.ChatApplicationCommand).
		Description("Displays bot statistics")

	exec := func(dac *discord.DiscordApplicationCommand) {
		var memory runtime.MemStats
		runtime.ReadMemStats(&memory)
		guilds, bots, humans, total := getGuildCounts(dac)
		uptime := time.Since(m.startTime)
		countStr := getCommandCountString(m)
		bugs := rand.New(rand.NewSource(m.startTime.UnixNano())).Intn(1000)

		embed := builders.NewEmbedBuilder().WithTitle("Meido statistics").WithOkColor().
			AddField("⏳ Uptime", uptime.String(), true).
			AddField("⌨️ Commands ran", countStr, true).
			AddField("🖥️ Servers", fmt.Sprint(len(guilds)), true).
			AddField("🧑‍🤝‍🧑 Users", fmt.Sprintf("Total: %v\nHumans: %v\nBots: %v", total, humans, bots), true).
			AddField("💾 Memory usage", fmt.Sprintf("%v/%v", humanize.Bytes(memory.Alloc), humanize.Bytes(memory.Sys)), true).
			AddField("🗑️ Garbage collected", humanize.Bytes(memory.TotalAlloc-memory.Alloc), true).
			AddField("👑 Owner IDs", strings.Join(m.Bot.Config.GetStringSlice("owner_ids"), ", "), true).
			AddField("🐞 Bugs", fmt.Sprint(bugs), true)
		dac.RespondEmbed(embed.Build())
	}

	return bld.Execute(exec).Build()
}

func getGuildCounts(dac *discord.DiscordApplicationCommand) ([]*discordgo.Guild, int, int, int) {
	var (
		totalUsers  int
		totalBots   int
		totalHumans int
	)

	guilds := dac.Sess.State().Guilds
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
	return guilds, totalBots, totalHumans, totalUsers
}

func getCommandCountString(m *module) string {
	countStr := "Not available"
	count, err := m.db.GetCommandCount()
	if err == nil {
		countStr = fmt.Sprint(count)
	}
	return countStr
}

func newColorCommand(m *module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "color",
		Description:      "Displays a small image of a provided color hex",
		Triggers:         []string{"m?color"},
		Usage:            "m?color [color hex]",
		Cooldown:         time.Second * 1,
		CooldownScope:    bot.CooldownScopeChannel,
		RequiredPerms:    0,
		CheckBotPerms:    false,
		RequiresUserType: bot.UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         true,
		Enabled:          true,
		Execute: func(msg *discord.DiscordMessage) {
			if len(msg.Args()) < 2 {
				return
			}

			clrStr := msg.Args()[1]
			clrStr = strings.TrimPrefix(clrStr, "#")
			buf, err := generateColorPNG(clrStr)
			if err != nil {
				_, _ = msg.Reply("Invalid hex code")
				return
			}
			_, _ = msg.ReplyComplex(&discordgo.MessageSend{File: &discordgo.File{Name: "color.png", Reader: buf}})
		},
	}
}

func newColorSlash(m *module) *bot.ModuleApplicationCommand {
	cmd := bot.NewModuleApplicationCommandBuilder(m, "color").
		Type(discordgo.ChatApplicationCommand).
		Description("Show the color of a provided hex").
		Cooldown(time.Second, bot.CooldownScopeChannel).
		AddOption(&discordgo.ApplicationCommandOption{
			Name:        "hex",
			Description: "The hex string of the desired color",
			Required:    true,
			Type:        discordgo.ApplicationCommandOptionString,
		})

	run := func(d *discord.DiscordApplicationCommand) {
		if len(d.Data.Options) < 1 {
			return
		}

		clrStrOpt, ok := d.Option("hex")
		if !ok {
			return
		}
		clrStr := clrStrOpt.StringValue()
		if !strings.HasPrefix(clrStr, "#") {
			clrStr = "#" + strings.TrimSpace(clrStr)
		}
		buf, err := generateColorPNG(clrStr)
		if err != nil {
			d.RespondEphemeral("Invalid hex code")
			return
		}
		d.RespondFile(
			fmt.Sprintf("Color hex: `%v`", strings.ToUpper(clrStr)),
			"color.png",
			buf,
		)
	}

	return cmd.Execute(run).Build()
}

func generateColorPNG(clrStr string) (*bytes.Buffer, error) {
	clr, err := hexcolor.Parse(clrStr)
	if err != nil {
		return nil, err
	}

	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: clr}, image.Point{}, draw.Src)
	buf := bytes.Buffer{}
	err = png.Encode(&buf, img)
	return &buf, err
}

func newIdTimestampCmd(m *module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "idtimestamp",
		Description:      "Converts a Discord ID to a timestamp",
		Triggers:         []string{"m?idt", "m?idts", "m?ts", "m?idtimestamp"},
		Usage:            "m?idt [ID]",
		Cooldown:         0,
		CooldownScope:    bot.CooldownScopeChannel,
		RequiredPerms:    0,
		CheckBotPerms:    false,
		RequiresUserType: bot.UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         true,
		Enabled:          true,
		Execute: func(msg *discord.DiscordMessage) {
			id := msg.AuthorID()
			if len(msg.Args()) > 1 {
				id = msg.Args()[1]
			}
			_, _ = msg.Reply(fmt.Sprintf("<t:%v>", utils.IDToTimestamp(id).Unix()))
		},
	}

}

func newInviteCommand(m *module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "invite",
		Description:      "Sends a bot invite link and support server invite link",
		Triggers:         []string{"m?invite"},
		Usage:            "m?invite",
		Cooldown:         time.Second * 1,
		CooldownScope:    bot.CooldownScopeChannel,
		RequiredPerms:    0,
		CheckBotPerms:    false,
		RequiresUserType: bot.UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         true,
		Enabled:          true,
		Execute: func(msg *discord.DiscordMessage) {
			botLink := "<https://discordapp.com/oauth2/authorize?client_id=" + msg.Sess.State().User.ID + "&scope=bot>"
			serverLink := "https://discord.gg/KgMEGK3"
			_, _ = msg.Reply(fmt.Sprintf("Invite me to your server: %v\nSupport server: %v", botLink, serverLink))
		},
	}
}

const (
	helpModuleSelect  = "help_module_select"
	helpCommandSelect = "help_command_select"
	helpModuleBack    = "help_module_back"
	helpCommandBack   = "help_command_back"
)

func newHelpSlash(m *module) *bot.ModuleApplicationCommand {
	cmd := bot.NewModuleApplicationCommandBuilder(m, "help").
		Type(discordgo.ChatApplicationCommand).
		Description("Displays helpful things").
		Cooldown(time.Second, bot.CooldownScopeChannel).
		AddOption(&discordgo.ApplicationCommandOption{
			Name:        "command",
			Description: "The command to show help for",
			Required:    false,
			Type:        discordgo.ApplicationCommandOptionString,
		})

	run := func(d *discord.DiscordApplicationCommand) {
		// if no options are provided, show the help menu
		if len(d.Data.Options) < 1 {
			menu := createHelpMenu(m, d.Sess, d.AuthorID())
			_ = d.RespondComplex(menu, discordgo.InteractionResponseChannelMessageWithSource)
			return
		}

		// otherwise, show the help for the provided command
		cmdOpt, ok := d.Option("command")
		if !ok {
			d.RespondEphemeral("Command not found")
			return
		}

		inp := cmdOpt.StringValue()
		if cmd, err := m.Bot.FindCommand(inp); err == nil {
			showCommandHelp(m, d, cmd)
			return
		}
		d.RespondEphemeral("Command not found")
	}

	return cmd.Execute(run).Build()
}

// createHelpMenu creates a help menu with a select menu for selecting a module.
func createHelpMenu(m *module, sess discord.DiscordSession, userID string) *discordgo.InteractionResponseData {
	var (
		options  = []discordgo.SelectMenuOption{}
		modNames = []string{}
	)

	for _, mod := range m.Bot.Modules {
		options = append(options, discordgo.SelectMenuOption{
			Label: mod.Name(),
			Value: mod.Name(),
		})
		modNames = append(modNames, fmt.Sprintf("`%v`", mod.Name()))
	}

	return &discordgo.InteractionResponseData{
		Embeds: []*discordgo.MessageEmbed{
			builders.NewEmbedBuilder().
				WithTitle("Meido Help").
				WithDescription(fmt.Sprintf("Select a module to see its commands.\n\nModules: %v", strings.Join(modNames, ", "))).
				WithThumbnail(sess.State().User.AvatarURL("256")).
				WithOkColor().
				Build(),
		},
		Components: []discordgo.MessageComponent{
			builders.NewActionRowBuilder().
				AddSelectMenu("Select a module", options, utils.JoinStrings(":", helpModuleSelect, userID)).
				Build(),
		},
	}
}

// createHelpModuleMenu creates a help menu with a select menu for selecting a command from a module.
// It also includes a back button to go back to the module selection menu.
func createHelpModuleMenu(m bot.Module, sess discord.DiscordSession, userID string) *discordgo.InteractionResponseData {
	var (
		options  = []discordgo.SelectMenuOption{}
		cmdNames = []string{}
	)

	for _, cmd := range m.Commands() {
		options = append(options, discordgo.SelectMenuOption{
			Label: cmd.Name,
			Value: cmd.Name,
		})
		cmdNames = append(cmdNames, fmt.Sprintf("`%v`", cmd.Name))
	}

	for _, cmd := range m.ApplicationCommands() {
		options = append(options, discordgo.SelectMenuOption{
			Label: "/" + cmd.Name,
			Value: "/" + cmd.Name,
		})
		cmdNames = append(cmdNames, fmt.Sprintf("`/%v`", cmd.Name))
	}

	return &discordgo.InteractionResponseData{
		Embeds: []*discordgo.MessageEmbed{
			builders.NewEmbedBuilder().
				WithTitle(fmt.Sprintf("Meido Help - %v", m.Name())).
				WithDescription(fmt.Sprintf("Select a command to see its info.\n\nCommands: %v", strings.Join(cmdNames, ", "))).
				WithThumbnail(sess.State().User.AvatarURL("256")).
				WithOkColor().
				Build(),
		},
		Components: []discordgo.MessageComponent{
			builders.NewActionRowBuilder().
				AddSelectMenu("Select a command", options, utils.JoinStrings(":", helpCommandSelect, userID)).
				Build(),
			builders.NewActionRowBuilder().
				AddButton("Back", discordgo.DangerButton, utils.JoinStrings(":", helpModuleBack, userID)).
				Build(),
		},
	}
}

func newHelpModuleSelectHandler(m *module) *bot.ModuleMessageComponent {
	return &bot.ModuleMessageComponent{
		Mod:           m,
		Name:          helpModuleSelect,
		Cooldown:      0,
		CooldownScope: bot.CooldownScopeChannel,
		Permissions:   0,
		UserType:      bot.UserTypeAny,
		CheckBotPerms: false,
		Enabled:       true,
		Execute: func(dmc *discord.DiscordMessageComponent) {
			parts := strings.Split(dmc.Data.CustomID, ":")
			if len(parts) < 2 || parts[1] != dmc.AuthorID() {
				dmc.RespondEmpty()
				return
			}

			selected := dmc.Interaction.MessageComponentData().Values[0]
			mod, err := m.Bot.FindModule(selected)
			if err != nil {
				dmc.UpdateRespose("You are using an old version of the help menu, please try again.")
				return
			}

			// show the commands for the selected module
			commands := mod.Commands()
			if len(commands) == 0 {
				dmc.UpdateRespose(fmt.Sprintf("No commands found for module %v", mod.Name()))
				return
			}

			menu := createHelpModuleMenu(mod, dmc.Sess, dmc.AuthorID())
			_ = dmc.RespondComplex(menu, discordgo.InteractionResponseUpdateMessage)
		},
	}
}

func newHelpModuleBackHandler(m *module) *bot.ModuleMessageComponent {
	return &bot.ModuleMessageComponent{
		Mod:           m,
		Name:          helpModuleBack,
		Cooldown:      0,
		CooldownScope: bot.CooldownScopeChannel,
		Permissions:   0,
		UserType:      bot.UserTypeAny,
		CheckBotPerms: false,
		Enabled:       true,
		Execute: func(dmc *discord.DiscordMessageComponent) {
			parts := strings.Split(dmc.Data.CustomID, ":")
			if len(parts) < 2 || parts[1] != dmc.AuthorID() {
				dmc.RespondEmpty()
				return
			}

			menu := createHelpMenu(m, dmc.Sess, dmc.AuthorID())
			_ = dmc.RespondComplex(menu, discordgo.InteractionResponseUpdateMessage)
		},
	}
}

func newHelpCommandSelectHandler(m *module) *bot.ModuleMessageComponent {
	return &bot.ModuleMessageComponent{
		Mod:           m,
		Name:          helpCommandSelect,
		Cooldown:      0,
		CooldownScope: bot.CooldownScopeChannel,
		Permissions:   0,
		UserType:      bot.UserTypeAny,
		CheckBotPerms: false,
		Enabled:       true,
		Execute: func(dmc *discord.DiscordMessageComponent) {
			parts := strings.Split(dmc.Data.CustomID, ":")
			if len(parts) < 2 || parts[1] != dmc.AuthorID() {
				dmc.RespondEmpty()
				return
			}

			var (
				embed *discordgo.MessageEmbed
				mod   bot.Module
			)
			selected := dmc.Interaction.MessageComponentData().Values[0]
			if strings.HasPrefix(selected, "/") {
				cmd, err := m.Bot.FindApplicationCommand(selected[1:])
				if err != nil {
					dmc.UpdateRespose("You are using an old version of the help menu, please try again.")

					return
				}
				embed = getApplicationCommandEmbed(cmd, dmc.Sess.State().User.AvatarURL("256"))
				mod = cmd.Mod
			} else {
				cmd, err := m.Bot.FindCommand(selected)
				if err != nil {
					dmc.UpdateRespose("You are using an old version of the help menu, please try again.")
					return
				}
				embed = getCommandEmbed(cmd, dmc.Sess.State().User.AvatarURL("256"))
				mod = cmd.Mod
			}

			resp := &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
				Components: []discordgo.MessageComponent{
					builders.NewActionRowBuilder().
						AddButton("Back", discordgo.DangerButton, utils.JoinStrings(":", helpCommandBack, mod.Name(), dmc.AuthorID())).
						Build(),
				},
			}
			_ = dmc.RespondComplex(resp, discordgo.InteractionResponseUpdateMessage)
		},
	}
}

func newHelpCommandBackHandler(m *module) *bot.ModuleMessageComponent {
	return &bot.ModuleMessageComponent{
		Mod:           m,
		Name:          helpCommandBack,
		Cooldown:      0,
		CooldownScope: bot.CooldownScopeChannel,
		Permissions:   0,
		UserType:      bot.UserTypeAny,
		CheckBotPerms: false,
		Enabled:       true,
		Execute: func(dmc *discord.DiscordMessageComponent) {
			parts := strings.Split(dmc.Data.CustomID, ":")
			if len(parts) < 3 || parts[2] != dmc.AuthorID() {
				dmc.RespondEmpty()
				return
			}

			mod, err := m.Bot.FindModule(parts[1])
			if err != nil {
				dmc.UpdateRespose("You are using an old version of the help menu, please try again.")
				return
			}

			menu := createHelpModuleMenu(mod, dmc.Sess, dmc.AuthorID())
			_ = dmc.RespondComplex(menu, discordgo.InteractionResponseUpdateMessage)
		},
	}
}

func showCommandHelp(_ *module, d *discord.DiscordApplicationCommand, cmd *bot.ModuleCommand) {
	embed := getCommandEmbed(cmd, d.Sess.State().User.AvatarURL("256"))
	_ = d.RespondEmbed(embed)
}

func getCommandEmbed(cmd *bot.ModuleCommand, avatarUrl string) *discordgo.MessageEmbed {
	text := strings.Builder{}
	text.WriteString(fmt.Sprintf("%v\n", cmd.Description))
	text.WriteString(fmt.Sprintf("\n**Usage**: %v", cmd.Usage))
	text.WriteString(fmt.Sprintf("\n**Aliases**: %v", strings.Join(cmd.Triggers, ", ")))
	if cmd.Cooldown > 0 {
		text.WriteString(fmt.Sprintf("\n**Cooldown**: %v", cmd.Cooldown))
	}
	text.WriteString(fmt.Sprintf("\n**Required permissions**: %v", discord.PermMap[cmd.RequiredPerms]))
	if !cmd.AllowDMs {
		text.WriteString(fmt.Sprintf("\n%v", "Cannot be used in DMs"))
	}

	embed := builders.NewEmbedBuilder().
		WithTitle(fmt.Sprintf("Command - %v", cmd.Name)).
		WithDescription(text.String()).
		WithOkColor().
		WithFooter("Arguments in [square brackets] are required, while arguments in <angle brackets> are optional.", "").
		WithThumbnail(avatarUrl).
		Build()
	return embed
}

func getApplicationCommandEmbed(cmd *bot.ModuleApplicationCommand, avatarUrl string) *discordgo.MessageEmbed {
	text := strings.Builder{}
	text.WriteString(fmt.Sprintf("%v\n", cmd.Description))
	if cmd.Cooldown > 0 {
		text.WriteString(fmt.Sprintf("\n**Cooldown**: %v", cmd.Cooldown))
	}
	if cmd.DefaultMemberPermissions != nil {
		text.WriteString(fmt.Sprintf("\n**Required permissions**: %v", discord.PermMap[*cmd.DefaultMemberPermissions]))
	}
	if !cmd.Mod.AllowDMs() {
		text.WriteString(fmt.Sprintf("\n%v", "Cannot be used in DMs"))
	}

	if len(cmd.Options) > 0 {
		text.WriteString("\n**Options**: ")
		for _, opt := range cmd.Options {
			optName := opt.Name
			if opt.Required {
				optName = "*" + optName
			}
			text.WriteString(fmt.Sprintf("\n\\- `%v`: %v", optName, opt.Description))
		}
	}

	embed := builders.NewEmbedBuilder().
		WithTitle(fmt.Sprintf("Command - %v", cmd.Name)).
		WithDescription(text.String()).
		WithOkColor().
		WithThumbnail(avatarUrl).
		WithFooter("Options marked with * are required.", "").
		Build()
	return embed
}

func newHelpCommand(m *module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "help",
		Description:      "Displays helpful things",
		Triggers:         []string{"m?help", "m?h"},
		Usage:            "m?help <module | command | passive>",
		Cooldown:         time.Second * 1,
		CooldownScope:    bot.CooldownScopeChannel,
		RequiredPerms:    0,
		CheckBotPerms:    false,
		RequiresUserType: bot.UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         true,
		Enabled:          true,
		Execute:          m.helpCommand,
	}
}

func (m *module) helpCommand(msg *discord.DiscordMessage) {
	embed := builders.NewEmbedBuilder().
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
		if cmd.Cooldown > 0 {
			info.WriteString(fmt.Sprintf("\n**Cooldown**: %v second(s)", cmd.Cooldown))
		}
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
