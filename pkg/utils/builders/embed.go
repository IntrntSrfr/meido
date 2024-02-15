package builders

import (
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/pkg/utils"
)

type EmbedBuilder struct {
	*discordgo.MessageEmbed
}

func NewEmbedBuilder() *EmbedBuilder {
	return &EmbedBuilder{&discordgo.MessageEmbed{}}
}

func (e *EmbedBuilder) AddField(name, value string, inline bool) *EmbedBuilder {
	e.Fields = append(e.Fields, &discordgo.MessageEmbedField{
		Name:   name,
		Value:  value,
		Inline: inline,
	})
	return e
}

func (e *EmbedBuilder) WithThumbnail(url string) *EmbedBuilder {
	e.Thumbnail = &discordgo.MessageEmbedThumbnail{
		URL: url,
	}
	return e
}

func (e *EmbedBuilder) WithFooter(text, url string) *EmbedBuilder {
	e.Footer = &discordgo.MessageEmbedFooter{
		Text:    text,
		IconURL: url,
	}
	return e
}

func (e *EmbedBuilder) WithAuthor(name, url string) *EmbedBuilder {
	e.Author = &discordgo.MessageEmbedAuthor{
		Name: name,
		URL:  url,
	}
	return e
}

func (e *EmbedBuilder) WithUrl(url string) *EmbedBuilder {
	e.URL = url
	return e
}

func (e *EmbedBuilder) WithImageUrl(url string) *EmbedBuilder {
	e.Image = &discordgo.MessageEmbedImage{
		URL: url,
	}
	return e
}

func (e *EmbedBuilder) WithTitle(title string) *EmbedBuilder {
	e.Title = title
	return e
}

func (e *EmbedBuilder) WithDescription(description string) *EmbedBuilder {
	e.Description = description
	return e
}

func (e *EmbedBuilder) WithTimestamp(timestamp string) *EmbedBuilder {
	e.Timestamp = timestamp
	return e
}

func (e *EmbedBuilder) WithOkColor() *EmbedBuilder {
	e.Color = utils.ColorInfo
	return e
}

func (e *EmbedBuilder) WithErrorColor() *EmbedBuilder {
	e.Color = utils.ColorCritical
	return e
}

func (e *EmbedBuilder) WithColor(color int) *EmbedBuilder {
	e.Color = color
	return e
}

func (e *EmbedBuilder) Build() *discordgo.MessageEmbed {
	return e.MessageEmbed
}
