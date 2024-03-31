package administration

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/pkg/mio"
	"github.com/intrntsrfr/meido/pkg/mio/bot"
	"github.com/intrntsrfr/meido/pkg/mio/discord"
	"github.com/intrntsrfr/meido/pkg/utils"
	"github.com/intrntsrfr/meido/pkg/utils/builders"
)

type module struct {
	*bot.ModuleBase
	dmLogChannels []string
}

func New(b *bot.Bot, logger mio.Logger) bot.Module {
	logger = logger.Named("Administration")
	return &module{
		ModuleBase:    bot.NewModule(b, "Administration", logger),
		dmLogChannels: b.Config.GetStringSlice("dm_log_channels"),
	}
}

func (m *module) Hook() error {
	if err := m.RegisterPassives(newForwardDmsPassive(m)); err != nil {
		return err
	}
	return m.RegisterCommands(
		newToggleCommandCommand(m),
		newMessageCommand(m),
	)
}

func newMessageCommand(m *module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "message",
		Description:      "Sends a message to a channel",
		Triggers:         []string{"m?msg"},
		Usage:            "m?msg [channelID] [message]",
		Cooldown:         0,
		CooldownScope:    bot.CooldownScopeChannel,
		RequiredPerms:    0,
		RequiresUserType: bot.UserTypeBotOwner,
		CheckBotPerms:    false,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         true,
		Enabled:          true,
		Execute: func(msg *discord.DiscordMessage) {
			if len(msg.Args()) < 3 {
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

func newToggleCommandCommand(m *module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "togglecommand",
		Description:      "Enables or disables a command. Bot owner only.",
		Triggers:         []string{"m?togglecommand", "m?tc"},
		Usage:            "m?tc ping",
		Cooldown:         time.Second * 2,
		CooldownScope:    bot.CooldownScopeChannel,
		RequiredPerms:    0,
		CheckBotPerms:    false,
		RequiresUserType: bot.UserTypeBotOwner,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         true,
		Enabled:          true,
		Execute: func(msg *discord.DiscordMessage) {
			if cmd, err := m.Bot.FindCommand(msg.RawContent()); err == nil {
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

func newForwardDmsPassive(m *module) *bot.ModulePassive {
	return &bot.ModulePassive{
		Mod:          m,
		Name:         "forwarddms",
		Description:  "Forwards all received DMs to channels specified by the bot owner",
		AllowedTypes: discord.MessageTypeCreate,
		Enabled:      true,
		Execute: func(msg *discord.DiscordMessage) {
			if !msg.IsDM() {
				return
			}
			embed := builders.NewEmbedBuilder().
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
