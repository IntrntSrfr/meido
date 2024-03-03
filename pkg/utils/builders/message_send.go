package builders

import "github.com/bwmarrin/discordgo"

type MessageSendBuilder struct {
	message *discordgo.MessageSend
}

func NewMessageSendBuilder() *MessageSendBuilder {
	return &MessageSendBuilder{message: &discordgo.MessageSend{}}
}

func (b *MessageSendBuilder) Content(content string) *MessageSendBuilder {
	b.message.Content = content
	return b
}

func (b *MessageSendBuilder) Embed(embed *discordgo.MessageEmbed) *MessageSendBuilder {
	b.message.Embed = embed
	return b
}

func (b *MessageSendBuilder) AddActionRow(actionRow *discordgo.ActionsRow) *MessageSendBuilder {
	if b.message.Components == nil {
		b.message.Components = make([]discordgo.MessageComponent, 0)
	}
	b.message.Components = append(b.message.Components, actionRow)
	return b
}

func (b *MessageSendBuilder) Build() *discordgo.MessageSend {
	return b.message
}
