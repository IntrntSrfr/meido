package utilitymod

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/pkg/mio"
	"github.com/intrntsrfr/meido/pkg/utils"
	"math"
	"time"
)

func NewServerCommand(m *UtilityMod) *mio.ModCommand {
	return &mio.ModCommand{
		Mod:           m,
		Name:          "server",
		Description:   "Displays information about the server",
		Triggers:      []string{"m?server"},
		Usage:         "m?server",
		Cooldown:      5,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run: func(msg *mio.DiscordMessage) {

			if msg.LenArgs() < 1 {
				return
			}

			g, err := msg.Discord.Guild(msg.Message.GuildID)
			if err != nil {
				_, _ = msg.Reply("Error getting guild data")
				return
			}

			tc := 0
			vc := 0

			for _, ch := range g.Channels {
				if ch.Type == discordgo.ChannelTypeGuildText {
					tc++
				} else if ch.Type == discordgo.ChannelTypeGuildVoice {
					vc++
				}
			}

			users := 0
			bots := 0

			for _, mem := range g.Members {
				if mem.User.Bot {
					bots++
				} else {
					users++
				}
			}

			owner, err := msg.Discord.Member(g.ID, g.OwnerID)
			if err != nil {
				_, _ = msg.Reply("Error getting guild data")
				return
			}

			ts := utils.IDToTimestamp(g.ID)
			dur := time.Since(ts)

			embed := &discordgo.MessageEmbed{
				Color: utils.ColorInfo,
				Author: &discordgo.MessageEmbedAuthor{
					Name: g.Name,
				},
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "Owner",
						Value:  fmt.Sprintf("%v\n(%v)", owner.Mention(), owner.User.ID),
						Inline: true,
					},
					{
						Name:  "Creation date",
						Value: fmt.Sprintf("%v | %v day(s) ago", ts.Format(time.RFC1123), math.Floor(dur.Hours()/24.0)),
					},
					{
						Name:   "Members",
						Value:  fmt.Sprintf("%v members\n%v users\n%v bots", g.MemberCount, users, bots),
						Inline: true,
					},
					{
						Name:   "Channels",
						Value:  fmt.Sprintf("Total: %v\nText: %v\nVoice: %v", len(g.Channels), tc, vc),
						Inline: true,
					},
					{
						Name:   "Roles",
						Value:  fmt.Sprintf("%v roles", len(g.Roles)),
						Inline: true,
					},
				},
			}
			if g.Icon != "" {
				embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
					URL: fmt.Sprintf("https://cdn.discordapp.com/icons/%v/%v.png", g.ID, g.Icon),
				}
				embed.Author.IconURL = fmt.Sprintf("https://cdn.discordapp.com/icons/%v/%v.png", g.ID, g.Icon)
			}

			_, _ = msg.ReplyEmbed(embed)
		},
	}
}

func NewServerSplashCommand(m *UtilityMod) *mio.ModCommand {
	return &mio.ModCommand{
		Mod:           m,
		Name:          "serversplash",
		Description:   "Displays server splash if one exists",
		Triggers:      []string{"m?serversplash"},
		Usage:         "m?serversplash",
		Cooldown:      5,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run: func(msg *mio.DiscordMessage) {
			if msg.LenArgs() < 1 {
				return
			}

			g, err := msg.Discord.Guild(msg.Message.GuildID)
			if err != nil {
				return
			}

			if g.Splash == "" {
				_, _ = msg.Reply("This server doesn't have a splash!")
				return
			}

			embed := &discordgo.MessageEmbed{
				Title: g.Name,
				Color: utils.ColorInfo,
				Image: &discordgo.MessageEmbedImage{
					URL: fmt.Sprintf("https://cdn.discordapp.com/splashes/%v/%v.png?size=2048", g.ID, g.Splash),
				},
			}
			_, _ = msg.ReplyEmbed(embed)
		},
	}
}

func NewServerIconCommand(m *UtilityMod) *mio.ModCommand {
	return &mio.ModCommand{
		Mod:           m,
		Name:          "servericon",
		Description:   "Displays server icon, if one exists",
		Triggers:      []string{"m?servericon", "m?si", ">si"},
		Usage:         "m?servericon",
		Cooldown:      5,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run: func(msg *mio.DiscordMessage) {
			if msg.LenArgs() < 1 {
				return
			}
			g, err := msg.Discord.Guild(msg.Message.GuildID)
			if err != nil {
				return
			}

			if g.Icon == "" {
				_, _ = msg.Reply("This server doesn't have an icon!")
				return
			}

			embed := &discordgo.MessageEmbed{
				Title: g.Name,
				Color: utils.ColorInfo,
				Image: &discordgo.MessageEmbedImage{
					URL: fmt.Sprintf("%v?size=2048", g.IconURL()),
				},
			}
			_, _ = msg.ReplyEmbed(embed)
		},
	}
}

func NewServerBannerCommand(m *UtilityMod) *mio.ModCommand {
	return &mio.ModCommand{
		Mod:           m,
		Name:          "serverbanner",
		Description:   "Displays server banner if one exists",
		Triggers:      []string{"m?serverbanner"},
		Usage:         "m?serverbanner",
		Cooldown:      5,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run: func(msg *mio.DiscordMessage) {
			if msg.LenArgs() < 1 {
				return
			}

			g, err := msg.Discord.Guild(msg.Message.GuildID)
			if err != nil {
				return
			}

			if g.Banner == "" {
				_, _ = msg.Reply("This server doesn't have a banner!")
				return
			}

			embed := &discordgo.MessageEmbed{
				Title: g.Name,
				Color: utils.ColorInfo,
				Image: &discordgo.MessageEmbedImage{
					URL: fmt.Sprintf("https://cdn.discordapp.com/banners/%v/%v.png?size=2048", g.ID, g.Banner),
				},
			}
			_, _ = msg.ReplyEmbed(embed)
		},
	}
}
