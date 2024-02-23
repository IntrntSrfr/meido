package utility

import (
	"fmt"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/pkg/mio/bot"
	"github.com/intrntsrfr/meido/pkg/mio/discord"
	"github.com/intrntsrfr/meido/pkg/utils"
	"github.com/intrntsrfr/meido/pkg/utils/builders"
)

func newAvatarCommand(m *module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "avatar",
		Description:      "Displays a users profile picture. User can be specified. Author is default.",
		Triggers:         []string{"m?avatar", "m?av", ">av"},
		Usage:            ">av <user>",
		Cooldown:         1,
		CooldownScope:    bot.CooldownScopeChannel,
		RequiredPerms:    0,
		CheckBotPerms:    false,
		RequiresUserType: bot.UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         true,
		Enabled:          true,
		Execute: func(msg *discord.DiscordMessage) {
			targetUser := msg.Author()
			var err error
			if len(msg.Args()) > 1 {
				targetUser, err = msg.GetMemberOrUserAtArg(1)
				if err != nil {
					return
				}
			}
			if targetUser == nil {
				return
			}

			embed := builders.NewEmbedBuilder().
				WithTitle(targetUser.String()).
				WithImageUrl(targetUser.AvatarURL("1024")).
				WithColor(msg.Discord.HighestColor(msg.Message.GuildID, targetUser.ID))
			_, _ = msg.ReplyEmbed(embed.Build())
		},
	}
}

func newBannerCommand(m *module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "banner",
		Description:      "Displays a users banner. User can be specified. Author is default.",
		Triggers:         []string{"m?banner", ">banner"},
		Usage:            ">banner <user>",
		Cooldown:         1,
		CooldownScope:    bot.CooldownScopeChannel,
		RequiredPerms:    0,
		CheckBotPerms:    false,
		RequiresUserType: bot.UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         true,
		Enabled:          true,
		Execute: func(msg *discord.DiscordMessage) {
			if len(msg.Args()) < 1 {
				return
			}

			targetUserID := msg.AuthorID()

			if len(msg.Args()) > 1 {
				targetUserID = msg.Args()[1]
			}

			targetUserID = utils.TrimUserID(targetUserID)
			if !utils.IsNumber(targetUserID) {
				return
			}

			targetUser, err := msg.Sess.User(targetUserID)
			if err != nil {
				fmt.Println(err)
				return
			}

			if targetUser.Banner == "" {
				_, _ = msg.Reply(fmt.Sprintf("**%v** doesn't have a server avatar!", targetUser.String()))
				return
			}

			embed := builders.NewEmbedBuilder().
				WithTitle(targetUser.String()).
				WithImageUrl(targetUser.BannerURL("1024")).
				WithColor(msg.Discord.HighestColor(msg.Message.GuildID, targetUser.ID))
			_, _ = msg.ReplyEmbed(embed.Build())
		},
	}
}

func newMemberAvatarCommand(m *module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "memberavatar",
		Description:      "Displays a members profile picture. User can be specified. Author is default.",
		Triggers:         []string{"m?memberavatar", "m?mav", ">mav"},
		Usage:            ">av <user>",
		Cooldown:         1,
		CooldownScope:    bot.CooldownScopeChannel,
		RequiredPerms:    0,
		CheckBotPerms:    false,
		RequiresUserType: bot.UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         false,
		Enabled:          true,
		Execute: func(msg *discord.DiscordMessage) {
			if len(msg.Args()) < 1 {
				return
			}

			targetMember := msg.Member()
			var err error

			if len(msg.Args()) > 1 {
				targetMember, err = msg.GetMemberAtArg(1)
				if err != nil {
					return
				}
			}

			if targetMember == nil {
				return
			}

			if targetMember.Avatar == "" {
				_, _ = msg.Reply(fmt.Sprintf("**%v** doesn't have a server avatar!", targetMember.User.String()))
				return
			}

			embed := builders.NewEmbedBuilder().
				WithTitle(targetMember.User.String()).
				WithImageUrl(targetMember.AvatarURL("1024")).
				WithColor(msg.Discord.HighestColor(msg.Message.GuildID, targetMember.User.ID))
			_, _ = msg.ReplyEmbed(embed.Build())
		},
	}
}

func newUserInfoCommand(m *module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "userinfo",
		Description:      "Displays information about a user",
		Triggers:         []string{"m?userinfo"},
		Usage:            "m?userinfo <user>",
		Cooldown:         1,
		CooldownScope:    bot.CooldownScopeChannel,
		RequiredPerms:    0,
		CheckBotPerms:    false,
		RequiresUserType: bot.UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         false,
		Enabled:          true,
		Execute: func(msg *discord.DiscordMessage) {
			targetUser := msg.Author()
			targetMember := msg.Member()
			if len(msg.Args()) > 1 {
				if len(msg.Message.Mentions) >= 1 {
					targetUser = msg.Message.Mentions[0]
					targetMember, _ = msg.Discord.Member(msg.Message.GuildID, msg.Message.Mentions[0].ID)
				} else {
					_, err := strconv.Atoi(msg.Args()[1])
					if err != nil {
						return
					}
					targetMember, err = msg.Discord.Member(msg.Message.GuildID, msg.Args()[1])
					if err != nil {
						targetUser, err = msg.Sess.User(msg.Args()[1])
						if err != nil {
							return
						}
					} else {
						targetUser = targetMember.User
					}
				}
			}

			embed := builders.NewEmbedBuilder().
				WithTitle(targetUser.String()).
				WithThumbnail(targetUser.AvatarURL("256")).
				AddField("ID | Mention", fmt.Sprintf("%v | <@!%v>", targetUser.ID, targetUser.ID), false).
				AddField("Creation date", fmt.Sprintf("<t:%v:R>", utils.IDToTimestamp(targetUser.ID).Unix()), false)

			if targetMember == nil {
				_, _ = msg.ReplyEmbed(embed.Build())
				return
			}

			nick := targetMember.Nick
			if nick == "" {
				nick = "None"
			}

			embed.WithColor(msg.Discord.HighestColor(msg.Message.GuildID, targetMember.User.ID)).
				AddField("Join date", fmt.Sprintf("<t:%v:R>", targetMember.JoinedAt.Unix()), false).
				AddField("Roles", fmt.Sprint(len(targetMember.Roles)), true).
				AddField("Nickname", nick, true)
			_, _ = msg.ReplyEmbed(embed.Build())
		},
	}
}

func newUserInfoUserCommand(m *module) *bot.ModuleApplicationCommand {
	bld := bot.NewModuleApplicationCommandBuilder(m, "info").
		Type(discordgo.UserApplicationCommand)

	run := func(msg *discord.DiscordApplicationCommand) {
		targetUser := msg.Data.Resolved.Users[msg.Data.TargetID]
		targetMember := msg.Data.Resolved.Members[msg.Data.TargetID]
		if targetUser == nil || targetMember == nil {
			return
		}

		embed := builders.NewEmbedBuilder().
			WithTitle(targetUser.String()).
			WithThumbnail(targetUser.AvatarURL("256")).
			WithColor(msg.Discord.HighestColor(msg.Interaction.GuildID, targetUser.ID)).
			AddField("ID | Mention", fmt.Sprintf("%v | <@!%v>", targetUser.ID, targetUser.ID), false).
			AddField("Creation date", fmt.Sprintf("<t:%v:R>", utils.IDToTimestamp(targetUser.ID).Unix()), false).
			AddField("Join date", fmt.Sprintf("<t:%v:R>", targetMember.JoinedAt.Unix()), false).
			AddField("Roles", fmt.Sprint(len(targetMember.Roles)), true)

		nick := targetMember.Nick
		if nick != "" {
			embed.AddField("Nickname", nick, true)
		}

		_ = msg.RespondEmbed(embed.Build())
	}

	return bld.Execute(run).Build()
}
