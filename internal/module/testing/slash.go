package testing

import "github.com/bwmarrin/discordgo"

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "pingo",
		Description: "pongo",
		Type:        discordgo.ChatApplicationCommand,
	},
}
