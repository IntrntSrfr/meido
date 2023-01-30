package utilitymod

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/internal/mods"
	"github.com/intrntsrfr/meido/pkg/mio"
	"github.com/intrntsrfr/meido/pkg/utils"
	"strconv"
)

func NewAvatarCommand(m *UtilityMod) *mio.ModCommand {
	return &mio.ModCommand{
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
			if msg.LenArgs() < 1 {
				return
			}

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

			_, _ = msg.ReplyEmbed(&discordgo.MessageEmbed{
				Color: msg.Discord.HighestColor(msg.Message.GuildID, targetUser.ID),
				Title: targetUser.String(),
				Image: &discordgo.MessageEmbedImage{URL: targetUser.AvatarURL("1024")},
			})
		},
	}
}

func NewBannerCommand(m *UtilityMod) *mio.ModCommand {
	return &mio.ModCommand{
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
				fmt.Println("thing isnt number")
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

			_, _ = msg.ReplyEmbed(&discordgo.MessageEmbed{
				Color: msg.Discord.HighestColor(msg.Message.GuildID, targetUser.ID),
				Title: targetUser.String(),
				Image: &discordgo.MessageEmbedImage{URL: targetUser.BannerURL("1024")},
			})
		},
	}
}

func NewMemberAvatarCommand(m *UtilityMod) *mio.ModCommand {
	return &mio.ModCommand{
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

			_, _ = msg.ReplyEmbed(&discordgo.MessageEmbed{
				Color: msg.Discord.HighestColor(msg.Message.GuildID, targetMember.User.ID),
				Title: targetMember.User.String(),
				Image: &discordgo.MessageEmbedImage{URL: targetMember.AvatarURL("1024")},
			})
		},
	}
}

func NewUserInfoCommand(m *UtilityMod) *mio.ModCommand {
	return &mio.ModCommand{
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

			ts := utils.IDToTimestamp(targetUser.ID)
			embed := &discordgo.MessageEmbed{}
			embed = mods.SetEmbedTitle(embed, fmt.Sprintf("User info | %v", targetUser.String()))
			embed = mods.SetEmbedThumbnail(embed, targetUser.AvatarURL("256"))
			embed = mods.AddEmbedField(embed, "ID | Mention", fmt.Sprintf("%v | <@!%v>", targetUser.ID, targetUser.ID), false)
			embed = mods.AddEmbedField(embed, "Creation date", fmt.Sprintf("<t:%v:R>", ts.Unix()), false)
			if targetMember == nil {
				_, _ = msg.ReplyEmbed(embed)
				return
			}

			nick := targetMember.Nick
			if nick == "" {
				nick = "None"
			}

			embed.Color = msg.Discord.HighestColor(msg.Message.GuildID, targetMember.User.ID)
			embed = mods.AddEmbedField(embed, "Join date", fmt.Sprintf("<t:%v:R>", targetMember.JoinedAt.Unix()), false)
			embed = mods.AddEmbedField(embed, "Roles", fmt.Sprint(len(targetMember.Roles)), true)
			embed = mods.AddEmbedField(embed, "Nickname", nick, true)
			_, _ = msg.ReplyEmbed(embed)
		},
	}
}
