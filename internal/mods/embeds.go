package mods

import "github.com/bwmarrin/discordgo"

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
