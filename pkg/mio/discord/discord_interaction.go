package discord

import (
	"io"
	"time"

	"github.com/bwmarrin/discordgo"
)

type DiscordInteraction struct {
	Sess         DiscordSession `json:"-"`
	Discord      *Discord       `json:"-"`
	Interaction  *discordgo.Interaction
	TimeReceived time.Time
	Shard        int
}

func (it *DiscordInteraction) ID() string {
	return it.Interaction.ID
}

func (it *DiscordInteraction) ChannelID() string {
	return it.Interaction.ChannelID
}

func (it *DiscordInteraction) AuthorID() string {
	if it.Interaction.GuildID == "" {
		return it.Interaction.User.ID
	}
	return it.Interaction.Member.User.ID
}

func (it *DiscordInteraction) GuildID() string {
	return it.Interaction.GuildID
}

func (it *DiscordInteraction) IsDM() bool {
	return it.Interaction.GuildID == ""
}

func (it *DiscordInteraction) AppPermissions() int64 {
	return it.Interaction.AppPermissions
}

func (it *DiscordInteraction) RespondComplex(data *discordgo.InteractionResponseData, responseType discordgo.InteractionResponseType) error {
	resp := &discordgo.InteractionResponse{
		Type: responseType,
		Data: data,
	}
	return it.Sess.InteractionRespond(it.Interaction, resp)
}

func (it *DiscordInteraction) Respond(text string) error {
	resp := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: text,
		},
	}
	return it.Sess.InteractionRespond(it.Interaction, resp)
}

func (it *DiscordInteraction) RespondEmpty() error {
	resp := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{},
	}
	return it.Sess.InteractionRespond(it.Interaction, resp)
}

func (it *DiscordInteraction) UpdateRespose(text string) error {
	resp := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content: text,
		},
	}
	return it.Sess.InteractionRespond(it.Interaction, resp)
}

func (it *DiscordInteraction) RespondEmbed(embed *discordgo.MessageEmbed) error {
	resp := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	}
	return it.Sess.InteractionRespond(it.Interaction, resp)
}

func (it *DiscordInteraction) UpdateResposeEmbed(embed *discordgo.MessageEmbed) error {
	resp := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	}
	return it.Sess.InteractionRespond(it.Interaction, resp)
}

func (it *DiscordInteraction) RespondEphemeral(text string) error {
	resp := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: text,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	}
	return it.Sess.InteractionRespond(it.Interaction, resp)
}

func (it *DiscordInteraction) RespondFile(text, name string, reader io.Reader) error {
	resp := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: text,
			Files: []*discordgo.File{{
				Name:   name,
				Reader: reader,
			}},
		},
	}
	return it.Sess.InteractionRespond(it.Interaction, resp)
}

type DiscordApplicationCommand struct {
	*DiscordInteraction
	Data    discordgo.ApplicationCommandInteractionData
	options map[string]*discordgo.ApplicationCommandInteractionDataOption
}

func (d *DiscordApplicationCommand) Name() string {
	return d.Data.Name
}

// SubCommand returns the name of the subcommand if the command is a subcommand
func (d *DiscordApplicationCommand) SubCommand() string {
	if d.Data.Options == nil {
		return ""
	}

	for _, opt := range d.Data.Options {
		if opt.Type == discordgo.ApplicationCommandOptionSubCommand {
			return opt.Name
		}
	}
	return ""
}

// SubCommandGroup returns the name of the subcommand group if the command is a subcommand group
func (d *DiscordApplicationCommand) SubCommandGroup() string {
	if d.Data.Options == nil {
		return ""
	}

	for _, opt := range d.Data.Options {
		if opt.Type == discordgo.ApplicationCommandOptionSubCommandGroup {
			return opt.Name
		}
	}
	return ""
}

// Options returns a map[string]*discordgo.ApplicationCommandInteractionDataOption
// where the key is the name of the option. This is a flattened map of all options
func (d *DiscordApplicationCommand) Options() map[string]*discordgo.ApplicationCommandInteractionDataOption {
	if d.options == nil {
		d.options = flattenOptions(d.Data.Options)
	}
	return d.options
}

// Option returns a *discordgo.ApplicationCommandInteractionDataOption given
// by key. This is a flattened map of all options
func (d *DiscordApplicationCommand) Option(key string) (*discordgo.ApplicationCommandInteractionDataOption, bool) {
	if d.options == nil {
		d.options = flattenOptions(d.Data.Options)
	}
	val, ok := d.options[key]
	return val, ok
}

func flattenOptions(options []*discordgo.ApplicationCommandInteractionDataOption) map[string]*discordgo.ApplicationCommandInteractionDataOption {
	result := make(map[string]*discordgo.ApplicationCommandInteractionDataOption)
	flattenOptionsImpl(options, result, "")
	return result
}

func flattenOptionsImpl(options []*discordgo.ApplicationCommandInteractionDataOption, result map[string]*discordgo.ApplicationCommandInteractionDataOption, prefix string) {
	for _, option := range options {
		key := prefix + option.Name

		if option.Type == discordgo.ApplicationCommandOptionSubCommand || option.Type == discordgo.ApplicationCommandOptionSubCommandGroup {
			opt := *option
			opt.Options = nil
			result[key] = &opt

			if option.Options != nil {
				newPrefix := key + ":"
				flattenOptionsImpl(option.Options, result, newPrefix)
			}
		} else {
			result[key] = option
		}
	}
}

type DiscordMessageComponent struct {
	*DiscordInteraction
	Data discordgo.MessageComponentInteractionData
}

type DiscordModalSubmit struct {
	*DiscordInteraction
	Data discordgo.ModalSubmitInteractionData
}
