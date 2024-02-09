package meido

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/internal/database"
	"github.com/intrntsrfr/meido/internal/module/administration"
	"github.com/intrntsrfr/meido/internal/module/customrole"
	"github.com/intrntsrfr/meido/internal/module/fishing"
	"github.com/intrntsrfr/meido/internal/module/fun"
	"github.com/intrntsrfr/meido/internal/module/mediatransform"
	"github.com/intrntsrfr/meido/internal/module/moderation"
	"github.com/intrntsrfr/meido/internal/module/search"
	"github.com/intrntsrfr/meido/internal/module/testing"
	"github.com/intrntsrfr/meido/internal/module/utility"
	"github.com/intrntsrfr/meido/internal/structs"
	"github.com/intrntsrfr/meido/pkg/mio"
	"github.com/intrntsrfr/meido/pkg/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Meido struct {
	Bot    *mio.Bot
	db     database.DB
	logger *zap.Logger
}

func New(config utils.Configurable, db database.DB) *Meido {
	logger := newLogger().Named("Meido")
	return &Meido{
		Bot:    mio.NewBot(config, logger),
		db:     db,
		logger: logger,
	}
}

func newLogger() *zap.Logger {
	loggerConfig := zap.NewDevelopmentConfig()
	loggerConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, _ := loggerConfig.Build()
	return logger
}

func (m *Meido) Run(ctx context.Context, useDefHandlers bool) error {
	m.Bot.UseDefaultHandlers()
	go m.listenMioEvents(ctx)
	m.registerModules()
	m.registerDiscordHandlers()
	return m.Bot.Run(ctx)
}

func (m *Meido) Close() {
	m.Bot.Close()
}

func (m *Meido) registerModules() {
	m.Bot.RegisterModule(administration.New(m.Bot, m.logger))
	m.Bot.RegisterModule(testing.New(m.Bot, m.logger))
	m.Bot.RegisterModule(fun.New(m.Bot, m.logger))
	m.Bot.RegisterModule(fishing.New(m.Bot, m.db, m.logger))
	m.Bot.RegisterModule(utility.New(m.Bot, m.db, m.logger))
	m.Bot.RegisterModule(moderation.New(m.Bot, m.db, m.logger))
	m.Bot.RegisterModule(customrole.New(m.Bot, m.db, m.logger))
	m.Bot.RegisterModule(search.New(m.Bot, m.logger))
	m.Bot.RegisterModule(mediatransform.New(m.Bot, m.logger))
}

func (m *Meido) listenMioEvents(ctx context.Context) {
	for {
		select {
		case evt := <-m.Bot.Events():
			switch evt.Type {
			case mio.BotEventCommandRan:
				m.logCommand(evt.Data.(*mio.CommandRan))
			case mio.BotEventCommandPanicked:
				m.logCommandPanicked(evt.Data.(*mio.CommandPanicked))
			case mio.BotEventPassivePanicked:
				m.logPassivePanicked(evt.Data.(*mio.PassivePanicked))
			}
		case <-ctx.Done():
			return
		}
	}
}

func (m *Meido) logCommand(cmd *mio.CommandRan) {
	err := m.db.CreateCommandLogEntry(&structs.CommandLogEntry{
		Command:   cmd.Command.Name,
		Args:      strings.Join(cmd.Message.Args(), " "),
		UserID:    cmd.Message.AuthorID(),
		GuildID:   cmd.Message.GuildID(),
		ChannelID: cmd.Message.ChannelID(),
		MessageID: cmd.Message.Message.ID,
		SentAt:    time.Now(),
	})
	if err != nil {
		m.logger.Error("Command write to DB failed", zap.Error(err))
	}
}

func (m *Meido) logCommandPanicked(cmd *mio.CommandPanicked) {
	m.logger.Error("Command panic",
		zap.Any("command", cmd.Command),
		zap.Any("message", cmd.Message),
		zap.String("stack trace", cmd.StackTrace),
	)
}

func (m *Meido) logPassivePanicked(pas *mio.PassivePanicked) {
	m.logger.Error("Passive panic",
		zap.Any("passive", pas.Passive),
		zap.Any("message", pas.Message),
		zap.String("stack trace", pas.StackTrace),
	)
}

func (m *Meido) registerDiscordHandlers() {
	m.Bot.Discord.AddEventHandler(insertGuild(m))
	m.Bot.Discord.AddEventHandlerOnce(statusLoop(m))
}

func insertGuild(m *Meido) func(s *discordgo.Session, g *discordgo.GuildCreate) {
	return func(s *discordgo.Session, g *discordgo.GuildCreate) {
		if _, err := m.db.GetGuild(g.Guild.ID); err != nil && err == sql.ErrNoRows {
			if err = m.db.CreateGuild(g.Guild.ID); err != nil {
				m.logger.Error("New guild write to DB failed", zap.Error(err), zap.String("guildID", g.ID))
			}
		}
	}
}

const totalStatusDisplays = 6

func statusLoop(m *Meido) func(s *discordgo.Session, r *discordgo.Ready) {
	statusTimer := time.NewTicker(time.Second * 15)
	return func(s *discordgo.Session, r *discordgo.Ready) {
		display := 0
		go func() {
			for range statusTimer.C {
				var (
					name       string
					statusType discordgo.ActivityType
				)
				switch display {
				case 0:
					srvCount := m.Bot.Discord.GuildCount()
					name = fmt.Sprintf("over %v servers", srvCount)
					statusType = discordgo.ActivityTypeWatching
				case 1:
					name = "m?help"
					statusType = discordgo.ActivityTypeGame
				case 2:
					name = "Remember to stay sane"
					statusType = discordgo.ActivityTypeGame
				case 3:
					name = "m?fish"
					statusType = discordgo.ActivityTypeGame
				case 4:
					name = "Changed custom role commands"
					statusType = discordgo.ActivityTypeGame
				case 5:
					name = "Auto roles are added!! Wow!!"
					statusType = discordgo.ActivityTypeGame
				}

				_ = s.UpdateStatusComplex(discordgo.UpdateStatusData{
					Activities: []*discordgo.Activity{{
						Name: name,
						Type: statusType,
					}},
				})
				display = (display + 1) % totalStatusDisplays
			}
		}()
	}
}
