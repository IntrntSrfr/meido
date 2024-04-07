package fun

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/gol"
	"github.com/intrntsrfr/meido/pkg/mio"
	"github.com/intrntsrfr/meido/pkg/mio/bot"
	"github.com/intrntsrfr/meido/pkg/mio/discord"
)

type module struct {
	*bot.ModuleBase
}

func New(b *bot.Bot, logger mio.Logger) bot.Module {
	logger = logger.Named("Fun")
	return &module{
		ModuleBase: bot.NewModule(b, "Fun", logger),
	}
}

func (m *module) Hook() error {
	if err := m.RegisterApplicationCommands(newLifeSlash(m)); err != nil {
		return err
	}
	if err := m.RegisterCommands(newLifeCommand(m)); err != nil {
		return err
	}
	return nil
}

func newLifeCommand(m *module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "life",
		Description:      "Shows a gif of Conway's Game of Life. If no seed is provided, it uses your user ID",
		Triggers:         []string{"m?life"},
		Usage:            "m?life | m?life <seed | user>",
		Cooldown:         time.Second * 5,
		CooldownScope:    bot.CooldownScopeChannel,
		RequiredPerms:    0,
		CheckBotPerms:    false,
		RequiresUserType: bot.UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         true,
		Enabled:          true,
		Execute: func(msg *discord.DiscordMessage) {
			_ = msg.Discord.StartTyping(msg.ChannelID())
			seedStr := msg.AuthorID()
			if len(msg.Args()) > 1 {
				seedStr = strings.Join(msg.Args()[1:], " ")
			}

			buf, err := generateGif(seedStr)
			if err != nil {
				_, _ = msg.Reply("There was an issue, please try again!")
				return
			}

			_, _ = msg.ReplyFile("", "game.gif", buf)
		},
	}
}

func newLifeSlash(m *module) *bot.ModuleApplicationCommand {
	cmd := bot.NewModuleApplicationCommandBuilder(m, "life").
		Type(discordgo.ChatApplicationCommand).
		Description("Show Conway's Game of Life").
		Cooldown(time.Second*5, bot.CooldownScopeChannel).
		AddOption(&discordgo.ApplicationCommandOption{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "seed",
			Description: "Random generator seed",
		})

	exec := func(dac *discord.DiscordApplicationCommand) {
		seed := dac.AuthorID()
		if seedOpt, ok := dac.Option("seed"); ok {
			seed = seedOpt.StringValue()
		}

		buf, err := generateGif(seed)
		if err != nil {
			_ = dac.Respond("Generation failed")
			return
		}

		_ = dac.RespondFile("", "game.gif", buf)
	}
	return cmd.Execute(exec).Build()
}

func generateGif(seedStr string) (*bytes.Buffer, error) {
	ye := sha1.New()
	_, err := ye.Write([]byte(seedStr))
	if err != nil {
		return nil, err
	}
	seed := int64(binary.BigEndian.Uint64(ye.Sum(nil)[:8]))
	game, _ := gol.NewGame(seed, 100, 100, true)
	game.Run(100, 50, false, true, "game.gif", 2)
	buf := bytes.Buffer{}
	_ = game.Export(&buf) // no need to check error, because export will always be populated
	return &buf, nil
}
