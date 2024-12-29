package builders

import (
	"bytes"

	"github.com/bwmarrin/discordgo"
)

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

func (b *MessageSendBuilder) WithTTS(tts bool) *MessageSendBuilder {
	b.message.TTS = tts
	return b
}

func (b *MessageSendBuilder) WithFile(file *discordgo.File) *MessageSendBuilder {
	b.message.File = file
	return b
}

func (b *MessageSendBuilder) WithFiles(files []*discordgo.File) *MessageSendBuilder {
	b.message.Files = files
	return b
}

func (b *MessageSendBuilder) AddTextFile(name, content string) *MessageSendBuilder {
	if b.message.Files == nil {
		b.message.Files = make([]*discordgo.File, 0)
	}

	b.message.Files = append(b.message.Files, &discordgo.File{
		Name:   name,
		Reader: bytes.NewBufferString(content),
	})

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
