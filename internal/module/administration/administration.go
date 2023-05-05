package administration

import (
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/internal/helpers"
	"github.com/intrntsrfr/meido/pkg/mio"
	"github.com/intrntsrfr/meido/pkg/utils"
	"go.uber.org/zap"
	"strings"
	"time"
)

// Module represents the administration mod
type Module struct {
	*mio.ModuleBase
	dmLogChannels []string
}

// New returns a new AdministrationMod.
func New(bot *mio.Bot, logger *zap.Logger) mio.Module {
	return &Module{
		ModuleBase:    mio.NewModule(bot, "Administration", logger.Named("administration")),
		dmLogChannels: bot.Config.GetStringSlice("dm_log_channels"),
	}
}

// Hook will hook the Module into the Bot.
func (m *Module) Hook() error {
	m.Bot.Discord.AddEventHandler(m.StatusLoop())
	if err := m.RegisterPassive(newForwardDmsPassive(m)); err != nil {
		return err
	}
	return m.RegisterCommands([]*mio.ModuleCommand{
		newToggleCommandCommand(m),
		newMessageCommand(m),
	})
}

func newMessageCommand(m *Module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:           m,
		Name:          "message",
		Description:   "Sends a message to a channel",
		Triggers:      []string{"m?msg"},
		Usage:         "m?msg [channelID] [message]",
		Cooldown:      0,
		CooldownUser:  false,
		RequiredPerms: 0,
		RequiresOwner: true,
		CheckBotPerms: false,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run: func(msg *mio.DiscordMessage) {
			if msg.LenArgs() < 3 {
				return
			}
			chID := msg.Args()[1]
			text := strings.Join(msg.Args()[1:], " ")

			if !utils.IsNumber(chID) {
				return
			}

			var data discordgo.MessageSend
			err := json.Unmarshal([]byte(text), &data)
			if err != nil {
				_, _ = msg.Reply("There was an issue")
				return
			}

			if _, err := msg.Sess.ChannelMessageSendComplex(chID, &data); err != nil {
				_, _ = msg.Reply("Could not deliver message")
				return
			}
			_, _ = msg.Reply("Message delivered")
		},
	}
}

// newToggleCommandCommand returns a new ping command.
func newToggleCommandCommand(m *Module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:           m,
		Name:          "togglecommand",
		Description:   "Enables or disables a command. Bot owner only.",
		Triggers:      []string{"m?togglecommand", "m?tc"},
		Usage:         "m?tc ping",
		Cooldown:      2,
		CooldownUser:  false,
		RequiredPerms: 0,
		RequiresOwner: true,
		CheckBotPerms: false,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run: func(msg *mio.DiscordMessage) {
			for _, mod := range m.Bot.Modules {
				cmd, ok := mio.FindCommand(mod, msg.Args())
				if !ok {
					return
				}
				if cmd.Name == "togglecommand" {
					return
				}
				cmd.Enabled = !cmd.Enabled
				if cmd.Enabled {
					_, _ = msg.Reply(fmt.Sprintf("Enabled command %v", cmd.Name))
					return
				}
				_, _ = msg.Reply(fmt.Sprintf("Disabled command %v", cmd.Name))
			}
		},
	}
}

func newForwardDmsPassive(m *Module) *mio.ModulePassive {
	return &mio.ModulePassive{
		Mod:          m,
		Name:         "forwarddms",
		Description:  "Forwards all received DMs to channels specified by the bot owner",
		AllowedTypes: mio.MessageTypeCreate,
		Enabled:      true,
		Run: func(msg *mio.DiscordMessage) {
			if !msg.IsDM() {
				return
			}

			embed := helpers.NewEmbed().
				WithTitle(fmt.Sprintf("Message from %v", msg.Message.Author.String())).
				WithOkColor().
				WithDescription(msg.Message.Content).
				WithFooter(msg.Message.Author.ID, "").
				WithTimestamp(msg.Message.Timestamp.Format(time.RFC3339))
			if len(msg.Message.Attachments) > 0 {
				embed.WithImageUrl(msg.Message.Attachments[0].URL)
			}
			for _, id := range m.dmLogChannels {
				_, _ = msg.Sess.ChannelMessageSendEmbed(id, embed.Build())
			}
		},
	}
}

func (m *Module) StatusLoop() func(s *discordgo.Session, r *discordgo.Ready) {
	statusTimer := time.NewTicker(time.Second * 15)
	return func(s *discordgo.Session, r *discordgo.Ready) {
		display := true
		go func() {
			for range statusTimer.C {
				if display {
					srvCount := 0
					for range m.Bot.Discord.Guilds() {
						srvCount++
					}
					_ = s.UpdateStatusComplex(discordgo.UpdateStatusData{
						Activities: []*discordgo.Activity{
							{
								Name: fmt.Sprintf("over %v servers", srvCount),
								Type: 3,
							},
						},
					})
				} else {
					_ = s.UpdateStatusComplex(discordgo.UpdateStatusData{
						Activities: []*discordgo.Activity{
							{
								Name: fmt.Sprintf("m?help"),
								Type: discordgo.ActivityTypeGame,
							},
						},
					})
				}
				display = !display
			}
		}()
	}
}
