package administration

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/pkg/mio"
	"github.com/intrntsrfr/meido/pkg/utils/builders"
)

type module struct {
	*mio.ModuleBase
	dmLogChannels []string
}

func New(b *mio.Bot, logger mio.Logger) mio.Module {
	logger = logger.Named("Administration")
	return &module{
		ModuleBase:    mio.NewModule(b, "Administration", logger),
		dmLogChannels: b.Config.GetStringSlice("dm_log_channels"),
	}
}

func (m *module) Hook() error {
	if err := m.RegisterPassives(
		newForwardDmsPassive(m),
	); err != nil {
		return err
	}

	if err := m.RegisterCommands(
		newToggleCommandCommand(m),
	); err != nil {
		return err
	}

	if err := m.RegisterApplicationCommands(
		newSendMessageSlash(m),
	); err != nil {
		return err
	}

	return nil
}

func newSendMessageSlash(m *module) *mio.ModuleApplicationCommand {
	cmd := mio.NewModuleApplicationCommandBuilder(m, "sendmessage").
		Type(discordgo.ChatApplicationCommand).
		RequiresBotOwner().
		Description("Sends a message to a channel").
		AddOption(&discordgo.ApplicationCommandOption{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "channel",
			Description: "Channel to send the message to",
			Required:    true,
		}).
		AddOption(&discordgo.ApplicationCommandOption{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "message",
			Description: "Message to send",
			Required:    true,
		})

	exec := func(d *mio.DiscordApplicationCommand) {
		channelOpt, ok := d.Options("channel")
		if !ok {
			return
		}
		channel := channelOpt.StringValue()

		messageOpt, ok := d.Options("message")
		if !ok {
			return
		}
		message := messageOpt.StringValue()

		// if we suspect its JSON, try to decode it
		complex := strings.HasPrefix(message, "{")
		if complex {
			var data discordgo.MessageSend
			err := json.Unmarshal([]byte(message), &data)
			if err != nil {
				_ = d.RespondEphemeral("There was an issue")
				return
			}

			sentMsg, err := d.Sess.ChannelMessageSendComplex(channel, &data)
			if err != nil {
				_ = d.RespondEphemeral("Could not deliver message")
				return
			}
			_ = d.RespondEphemeral(fmt.Sprintf("Message delivered. Link: https://discord.com/channels/%v/messages/%v", sentMsg.ChannelID, sentMsg.ID))
			return
		}

		sentMsg, err := d.Sess.ChannelMessageSend(channel, message)
		if err != nil {
			_ = d.RespondEphemeral("Could not deliver message")
			return
		}
		_ = d.RespondEphemeral(fmt.Sprintf("Message delivered. Link: https://discord.com/channels/%v/messages/%v", sentMsg.ChannelID, sentMsg.ID))
	}

	return cmd.Execute(exec).Build()
}

func newToggleCommandCommand(m *module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:              m,
		Name:             "togglecommand",
		Description:      "Enables or disables a command. Bot owner only.",
		Triggers:         []string{"m?togglecommand", "m?tc"},
		Usage:            "m?tc ping",
		Cooldown:         time.Second * 2,
		CooldownScope:    mio.CooldownScopeChannel,
		RequiredPerms:    0,
		CheckBotPerms:    false,
		RequiresUserType: mio.UserTypeBotOwner,
		AllowedTypes:     mio.MessageTypeCreate,
		AllowDMs:         true,
		Enabled:          true,
		Execute: func(msg *mio.DiscordMessage) {
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

func newForwardDmsPassive(m *module) *mio.ModulePassive {
	return &mio.ModulePassive{
		Mod:          m,
		Name:         "forwarddms",
		Description:  "Forwards all received DMs to channels specified by the bot owner",
		AllowedTypes: mio.MessageTypeCreate,
		Enabled:      true,
		Execute: func(msg *mio.DiscordMessage) {
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
