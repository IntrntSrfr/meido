package utility

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/pkg/mio"
	"github.com/intrntsrfr/meido/pkg/utils"
	"github.com/intrntsrfr/meido/pkg/utils/builders"
)

func newServerCommand(m *module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:              m,
		Name:             "server",
		Description:      "Displays information about the server",
		Triggers:         []string{"m?server"},
		Usage:            "m?server",
		Cooldown:         time.Second * 5,
		CooldownScope:    mio.CooldownScopeChannel,
		RequiredPerms:    0,
		CheckBotPerms:    false,
		RequiresUserType: mio.UserTypeAny,
		AllowedTypes:     mio.MessageTypeCreate,
		AllowDMs:         false,
		Enabled:          true,
		Execute: func(msg *mio.DiscordMessage) {
			if len(msg.Args()) < 1 {
				return
			}

			g, err := msg.Discord.Guild(msg.Message.GuildID)
			if err != nil {
				_, _ = msg.Reply("Error getting guild data")
				return
			}

			tc, vc := 0, 0
			for _, ch := range g.Channels {
				if ch.Type == discordgo.ChannelTypeGuildText {
					tc++
				} else if ch.Type == discordgo.ChannelTypeGuildVoice {
					vc++
				}
			}

			users, bots := 0, 0
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

			embed := builders.NewEmbedBuilder().
				WithAuthor(g.Name, "").
				WithOkColor().
				AddField("Owner", fmt.Sprintf("%v\n(%v)", owner.Mention(), owner.User.ID), false).
				AddField("Creation date", fmt.Sprintf("<t:%v:R>", utils.IDToTimestamp(g.ID).Unix()), false).
				AddField("Members", fmt.Sprintf("Total: %v\nHuman: %v\nBots: %v", g.MemberCount, users, bots), true).
				AddField("Channels", fmt.Sprintf("Total: %v\nText: %v\nVoice: %v", len(g.Channels), tc, vc), true).
				AddField("Roles", fmt.Sprintf("%v roles", len(g.Roles)), true)

			if g.Icon != "" {
				embed = embed.WithThumbnail(fmt.Sprintf("https://cdn.discordapp.com/icons/%v/%v.png", g.ID, g.Icon))
				embed = embed.WithAuthor(g.Name, fmt.Sprintf("https://cdn.discordapp.com/icons/%v/%v.png", g.ID, g.Icon))
			}

			_, _ = msg.ReplyEmbed(embed.Build())
		},
	}
}

func newServerSplashCommand(m *module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:              m,
		Name:             "serversplash",
		Description:      "Displays server splash if one exists",
		Triggers:         []string{"m?serversplash"},
		Usage:            "m?serversplash",
		Cooldown:         time.Second * 5,
		CooldownScope:    mio.CooldownScopeChannel,
		RequiredPerms:    0,
		CheckBotPerms:    false,
		RequiresUserType: mio.UserTypeAny,
		AllowedTypes:     mio.MessageTypeCreate,
		AllowDMs:         false,
		Enabled:          true,
		Execute: func(msg *mio.DiscordMessage) {
			if len(msg.Args()) < 1 {
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

			embed := builders.NewEmbedBuilder().
				WithTitle(g.Name).
				WithOkColor().
				WithImageUrl(fmt.Sprintf("https://cdn.discordapp.com/splashes/%v/%v.png?size=2048", g.ID, g.Splash))
			_, _ = msg.ReplyEmbed(embed.Build())
		},
	}
}

func newServerIconCommand(m *module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:              m,
		Name:             "servericon",
		Description:      "Displays server icon, if one exists",
		Triggers:         []string{"m?servericon", "m?si", ">si"},
		Usage:            "m?servericon",
		Cooldown:         time.Second * 5,
		CooldownScope:    mio.CooldownScopeChannel,
		RequiredPerms:    0,
		CheckBotPerms:    false,
		RequiresUserType: mio.UserTypeAny,
		AllowedTypes:     mio.MessageTypeCreate,
		AllowDMs:         false,
		Enabled:          true,
		Execute: func(msg *mio.DiscordMessage) {
			if len(msg.Args()) < 1 {
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

			embed := builders.NewEmbedBuilder().
				WithTitle(g.Name).
				WithOkColor().
				WithImageUrl(g.IconURL("1024"))
			_, _ = msg.ReplyEmbed(embed.Build())
		},
	}
}

func newServerBannerCommand(m *module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:              m,
		Name:             "serverbanner",
		Description:      "Displays server banner if one exists",
		Triggers:         []string{"m?serverbanner"},
		Usage:            "m?serverbanner",
		Cooldown:         time.Second * 5,
		CooldownScope:    mio.CooldownScopeChannel,
		RequiredPerms:    0,
		CheckBotPerms:    false,
		RequiresUserType: mio.UserTypeAny,
		AllowedTypes:     mio.MessageTypeCreate,
		AllowDMs:         false,
		Enabled:          true,
		Execute: func(msg *mio.DiscordMessage) {
			if len(msg.Args()) < 1 {
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

			embed := builders.NewEmbedBuilder().
				WithTitle(g.Name).
				WithOkColor().
				WithImageUrl(fmt.Sprintf("https://cdn.discordapp.com/banners/%v/%v.png?size=2048", g.ID, g.Banner))
			_, _ = msg.ReplyEmbed(embed.Build())
		},
	}
}
