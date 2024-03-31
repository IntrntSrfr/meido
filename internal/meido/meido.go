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
	"github.com/intrntsrfr/meido/internal/module/moderation"
	"github.com/intrntsrfr/meido/internal/module/search"
	"github.com/intrntsrfr/meido/internal/module/testing"
	"github.com/intrntsrfr/meido/internal/module/utility"
	"github.com/intrntsrfr/meido/pkg/mio"
	"github.com/intrntsrfr/meido/pkg/mio/bot"
	"github.com/intrntsrfr/meido/pkg/utils"
	"go.uber.org/zap"
)

type Meido struct {
	Bot    *bot.Bot
	db     database.DB
	logger mio.Logger
	config *utils.Config
}

func New(config *utils.Config, db database.DB) *Meido {
	logger := newLogger("Meido")

	b := bot.NewBotBuilder(config).
		WithDefaultHandlers().
		WithLogger(logger).
		Build()

	return &Meido{
		Bot:    b,
		db:     db,
		logger: logger,
		config: config,
	}
}

func (m *Meido) Run(ctx context.Context, useDefHandlers bool) error {
	go m.listenMioEvents(ctx)
	m.registerModules()
	m.registerDiscordHandlers()
	return m.Bot.Run(ctx)
}

func (m *Meido) Close() {
	m.Bot.Close()
}

func (m *Meido) registerModules() {
	modules := []bot.Module{
		administration.New(m.Bot, m.logger),
		testing.New(m.Bot, m.logger),
		fun.New(m.Bot, m.logger),
		fishing.New(m.Bot, m.db, m.logger),
		utility.New(m.Bot, m.db, m.logger),
		moderation.New(m.Bot, m.db, m.logger),
		customrole.New(m.Bot, m.db, m.logger),
		search.New(m.Bot, m.logger),
	}

	excludedModules := m.config.GetStringSlice("excluded_modules")
	for _, mod := range modules {
		if !utils.StringInSlice(strings.ToLower(mod.Name()), excludedModules) {
			m.Bot.RegisterModule(mod)
		}
	}
}

func (m *Meido) registerDiscordHandlers() {
	m.Bot.Discord.AddEventHandler(insertGuild(m))
	m.Bot.Discord.AddEventHandlerOnce(statusLoop(m))
}

func insertGuild(m *Meido) func(s *discordgo.Session, g *discordgo.GuildCreate) {
	return func(s *discordgo.Session, g *discordgo.GuildCreate) {
		if dbg, err := m.db.GetGuild(g.Guild.ID); err != nil && err == sql.ErrNoRows {
			if err = m.db.CreateGuild(g.Guild.ID, g.Guild.JoinedAt); err != nil {
				m.logger.Error("New guild write to DB failed", zap.Error(err), zap.String("guildID", g.ID))
			}
		} else if err == nil {
			dbg.JoinedAt = &g.Guild.JoinedAt
			if err := m.db.UpdateGuild(dbg); err != nil {
				m.logger.Error("Update guild joinedAt failed")
			}
		}
	}
}

const totalStatusDisplays = 3

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
					name = fmt.Sprintf("%v servers", srvCount)
					statusType = discordgo.ActivityTypeWatching
				case 1:
					name = "for /help | m?help"
					statusType = discordgo.ActivityTypeWatching
				case 2:
					name = "around with fish"
					statusType = discordgo.ActivityTypeGame
					/*
						case 3:
							name = "m?fish"
							statusType = discordgo.ActivityTypeGame
						case 4:
							name = "Changed custom role commands"
							statusType = discordgo.ActivityTypeGame
						case 5:
							name = "Auto roles are added!! Wow!!"
							statusType = discordgo.ActivityTypeGame
					*/
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
