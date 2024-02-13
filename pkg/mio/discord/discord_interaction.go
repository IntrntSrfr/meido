package discord

import (
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

func (it *DiscordInteraction) IsDM() bool {
	return it.Interaction.GuildID == ""
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

type DiscordApplicationCommand struct {
	*DiscordInteraction
	Data discordgo.ApplicationCommandInteractionData
}

type DiscordMessageComponent struct {
	*DiscordInteraction
	Data discordgo.MessageComponentInteractionData
}

type DiscordModalSubmit struct {
	*DiscordInteraction
	Data discordgo.ModalSubmitInteractionData
}
