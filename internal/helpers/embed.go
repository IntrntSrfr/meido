package helpers

import (
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/pkg/utils"
)

type Embed struct {
	*discordgo.MessageEmbed
}

func NewEmbed() *Embed {
	return &Embed{&discordgo.MessageEmbed{}}
}

func (e *Embed) AddField(name, value string, inline bool) *Embed {
	e.Fields = append(e.Fields, &discordgo.MessageEmbedField{
		Name:   name,
		Value:  value,
		Inline: inline,
	})
	return e
}

func (e *Embed) WithThumbnail(url string) *Embed {
	e.Thumbnail = &discordgo.MessageEmbedThumbnail{
		URL: url,
	}
	return e
}

func (e *Embed) WithFooter(text, url string) *Embed {
	e.Footer = &discordgo.MessageEmbedFooter{
		Text:    text,
		IconURL: url,
	}
	return e
}

func (e *Embed) WithAuthor(name, url string) *Embed {
	e.Author = &discordgo.MessageEmbedAuthor{
		Name: name,
		URL:  url,
	}
	return e
}

func (e *Embed) WithUrl(url string) *Embed {
	e.URL = url
	return e
}

func (e *Embed) WithImageUrl(url string) *Embed {
	e.Image = &discordgo.MessageEmbedImage{
		URL: url,
	}
	return e
}

func (e *Embed) WithTitle(title string) *Embed {
	e.Title = title
	return e
}

func (e *Embed) WithDescription(description string) *Embed {
	e.Description = description
	return e
}

func (e *Embed) WithTimestamp(timestamp string) *Embed {
	e.Timestamp = timestamp
	return e
}

func (e *Embed) WithOkColor() *Embed {
	e.Color = utils.ColorInfo
	return e
}

func (e *Embed) WithErrorColor() *Embed {
	e.Color = utils.ColorCritical
	return e
}

func (e *Embed) WithColor(color int) *Embed {
	e.Color = color
	return e
}

func AddEmbedField(e *discordgo.MessageEmbed, name, value string, inline bool) *discordgo.MessageEmbed {
	e.Fields = append(e.Fields, &discordgo.MessageEmbedField{
		Name:   name,
		Value:  value,
		Inline: inline,
	})
	return e
}

func SetEmbedThumbnail(e *discordgo.MessageEmbed, url string) *discordgo.MessageEmbed {
	e.Thumbnail = &discordgo.MessageEmbedThumbnail{
		URL: url,
	}
	return e
}

func SetEmbedTitle(e *discordgo.MessageEmbed, title string) *discordgo.MessageEmbed {
	e.Title = title
	return e
}

func (e *Embed) Build() *discordgo.MessageEmbed {
	return e.MessageEmbed
}
