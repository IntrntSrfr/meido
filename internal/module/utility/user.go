package utility

import (
	"fmt"
	"github.com/intrntsrfr/meido/internal/helpers"
	"github.com/intrntsrfr/meido/pkg/mio"
	"github.com/intrntsrfr/meido/pkg/utils"
	"strconv"
)

func newAvatarCommand(m *Module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:           m,
		Name:          "avatar",
		Description:   "Displays a users profile picture. User can be specified. Author is default.",
		Triggers:      []string{"m?avatar", "m?av", ">av"},
		Usage:         ">av <user>",
		Cooldown:      1,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run: func(msg *mio.DiscordMessage) {
			targetUser := msg.Author()
			var err error
			if msg.LenArgs() > 1 {
				targetUser, err = msg.GetMemberOrUserAtArg(1)
				if err != nil {
					return
				}
			}
			if targetUser == nil {
				return
			}

			embed := helpers.NewEmbed().
				WithTitle(targetUser.String()).
				WithImageUrl(targetUser.AvatarURL("1024")).
				WithColor(msg.Discord.HighestColor(msg.Message.GuildID, targetUser.ID))
			_, _ = msg.ReplyEmbed(embed.Build())
		},
	}
}

func newBannerCommand(m *Module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:           m,
		Name:          "banner",
		Description:   "Displays a users banner. User can be specified. Author is default.",
		Triggers:      []string{"m?banner", ">banner"},
		Usage:         ">banner <user>",
		Cooldown:      1,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run: func(msg *mio.DiscordMessage) {
			if msg.LenArgs() < 1 {
				return
			}

			targetUserID := msg.AuthorID()

			if msg.LenArgs() > 1 {
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

			embed := helpers.NewEmbed().
				WithTitle(targetUser.String()).
				WithImageUrl(targetUser.BannerURL("1024")).
				WithColor(msg.Discord.HighestColor(msg.Message.GuildID, targetUser.ID))
			_, _ = msg.ReplyEmbed(embed.Build())
		},
	}
}

func newMemberAvatarCommand(m *Module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:           m,
		Name:          "memberavatar",
		Description:   "Displays a members profile picture. User can be specified. Author is default.",
		Triggers:      []string{"m?memberavatar", "m?mav", ">mav"},
		Usage:         ">av | >av 123123123123",
		Cooldown:      1,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run: func(msg *mio.DiscordMessage) {
			if msg.LenArgs() < 1 {
				return
			}

			targetMember := msg.Member()
			var err error

			if msg.LenArgs() > 1 {
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

			embed := helpers.NewEmbed().
				WithTitle(targetMember.User.String()).
				WithImageUrl(targetMember.AvatarURL("1024")).
				WithColor(msg.Discord.HighestColor(msg.Message.GuildID, targetMember.User.ID))
			_, _ = msg.ReplyEmbed(embed.Build())
		},
	}
}

func newUserInfoCommand(m *Module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:           m,
		Name:          "userinfo",
		Description:   "Displays information about a user",
		Triggers:      []string{"m?userinfo"},
		Usage:         "m?userinfo | m?userinfo @user",
		Cooldown:      1,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run: func(msg *mio.DiscordMessage) {
			targetUser := msg.Author()
			targetMember := msg.Member()
			if msg.LenArgs() > 1 {
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

			embed := helpers.NewEmbed().
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