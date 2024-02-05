package fun

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/gol"
	"github.com/intrntsrfr/meido/pkg/mio"
	"go.uber.org/zap"
)

type Module struct {
	*mio.ModuleBase
}

func New(bot *mio.Bot, logger *zap.Logger) mio.Module {
	logger = logger.Named("Fun")
	return &Module{
		ModuleBase: mio.NewModule(bot, "Fun", logger),
	}
}

func (m *Module) Hook() error {
	return m.RegisterCommand(newLifeCommand(m))
}

func newLifeCommand(m *Module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:              m,
		Name:             "life",
		Description:      "Shows a gif of Conway's Game of Life. If no seed is provided, it uses your user ID",
		Triggers:         []string{"m?life"},
		Usage:            "m?life | m?life <seed | user>",
		Cooldown:         5,
		CooldownScope:    mio.Channel,
		RequiredPerms:    0,
		CheckBotPerms:    false,
		RequiresUserType: mio.UserTypeAny,
		AllowedTypes:     mio.MessageTypeCreate,
		AllowDMs:         true,
		IsEnabled:        true,
		Run: func(msg *mio.DiscordMessage) {
			_ = msg.Discord.StartTyping(msg.ChannelID())
			seedStr := msg.AuthorID()
			if len(msg.Args()) > 1 {
				seedStr = strings.Join(msg.Args()[1:], " ")
			}

			buf, seed, err := generateGif(seedStr)
			if err != nil {
				_, _ = msg.Reply("There was an issue, please try again!")
				return
			}

			_, _ = msg.ReplyComplex(&discordgo.MessageSend{
				Content: fmt.Sprintf("Here you go! Seed: `%v`", seed),
				File: &discordgo.File{
					Name:   "game.gif",
					Reader: buf,
				},
				Reference: &discordgo.MessageReference{
					MessageID: msg.Message.ID,
					ChannelID: msg.ChannelID(),
					GuildID:   msg.GuildID(),
				},
				AllowedMentions: &discordgo.MessageAllowedMentions{},
			})
		},
	}
}

func generateGif(seedStr string) (*bytes.Buffer, int64, error) {
	ye := sha1.New()
	_, err := ye.Write([]byte(seedStr))
	if err != nil {
		return nil, 0, err
	}
	seed := int64(binary.BigEndian.Uint64(ye.Sum(nil)[:8]))
	game, err := gol.NewGame(seed, 100, 100, true)
	game.Run(100, 50, false, true, "game.gif", 2)
	buf := bytes.Buffer{}
	_ = game.Export(&buf) // no need to check error, because export will always be populated
	return &buf, seed, nil
}
