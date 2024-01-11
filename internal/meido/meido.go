package meido

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/internal/database"
	"github.com/intrntsrfr/meido/internal/module/administration"
	"github.com/intrntsrfr/meido/internal/module/fishing"
	"github.com/intrntsrfr/meido/internal/module/fun"
	"github.com/intrntsrfr/meido/internal/module/mediaconvertmod"
	"github.com/intrntsrfr/meido/internal/module/search"
	"github.com/intrntsrfr/meido/internal/module/utility"
	"github.com/intrntsrfr/meido/internal/structs"
	"github.com/intrntsrfr/meido/pkg/mio"
	"go.uber.org/zap"
)

type Meido struct {
	Bot    *mio.Bot
	db     database.DB
	logger *zap.Logger
}

func New(config mio.Configurable, db database.DB, log *zap.Logger) *Meido {
	return &Meido{
		Bot:    mio.NewBot(config, log.Named("mio")),
		db:     db,
		logger: log,
	}
}

func (m *Meido) Run(useDefHandlers bool) error {
	if err := m.Bot.Open(useDefHandlers); err != nil {
		return err
	}

	// register modules here
	m.registerModules()
	// register mio event handlers here
	m.registerMioHandlers()
	// register discord event handlers here
	m.registerDiscordHandlers()

	if err := m.Bot.Run(); err != nil {
		return err
	}
	return nil
}

func (m *Meido) Close() {
	m.Bot.Close()
}

func (m *Meido) registerModules() {
	m.Bot.RegisterModule(administration.New(m.Bot, m.logger))
	//m.Bot.RegisterModule(testing.New(m.Bot, m.logger))
	m.Bot.RegisterModule(fun.New(m.Bot, m.logger))
	m.Bot.RegisterModule(fishing.New(m.Bot, m.db, m.logger))
	m.Bot.RegisterModule(utility.New(m.Bot, m.db, m.logger))
	//m.Bot.RegisterModule(moderation.New(m.Bot, m.Bot.DB, m.logger))
	//m.Bot.RegisterModule(customrole.New(m.Bot, m.Bot.DB, m.logger))
	m.Bot.RegisterModule(search.New(m.Bot, m.logger))
	m.Bot.RegisterModule(mediaconvertmod.New(m.Bot, m.logger))
	//m.Bot.RegisterModule(aimod.New(gptClient, config.GPT3Engine))
}

func (m *Meido) registerMioHandlers() {
	m.Bot.AddEventHandler("command_ran", logCommand(m))
	m.Bot.AddEventHandler("command_panicked", logCommandPanicked(m))
}

func logCommand(m *Meido) func(i interface{}) {
	return func(i interface{}) {
		cmd, _ := i.(*mio.CommandRan)
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
			m.logger.Error("error logging command", zap.Error(err))
		}
	}
}

func logCommandPanicked(m *Meido) func(i interface{}) {
	return func(i interface{}) {
		cmd, _ := i.(*mio.CommandPanicked)
		m.logger.Error("command panicked",
			zap.Any("command", cmd.Command),
			zap.Any("message", cmd.Message),
			zap.String("stack trace", cmd.StackTrace),
		)
	}
}

func (m *Meido) registerDiscordHandlers() {
	m.Bot.Discord.AddEventHandler(insertGuild(m))
	m.Bot.Discord.AddEventHandlerOnce(statusLoop(m))
}

func insertGuild(m *Meido) func(s *discordgo.Session, g *discordgo.GuildCreate) {
	return func(s *discordgo.Session, g *discordgo.GuildCreate) {
		if _, err := m.db.GetGuild(g.Guild.ID); err != nil && err == sql.ErrNoRows {
			if err = m.db.CreateGuild(g.Guild.ID); err != nil {
				m.logger.Error("could not create new guild", zap.Error(err), zap.String("guild ID", g.ID))
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
