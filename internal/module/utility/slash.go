package utility

import "github.com/bwmarrin/discordgo"

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "color",
		Description: "Displays a color from a given hex string",
		Type:        discordgo.ChatApplicationCommand,
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "hex",
				Description: "The hex string of the desired color",
				Required:    true,
				Type:        discordgo.ApplicationCommandOptionString,
			},
		},
	},
}
