package utilitymod

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/base"
	"github.com/intrntsrfr/meido/utils"
	"math"
	"strconv"
	"time"
)

func NewAvatarCommand(m *UtilityMod) *base.ModCommand {
	return &base.ModCommand{
		Mod:           m,
		Name:          "avatar",
		Description:   "Displays a users profile picture. User can be specified. Author is default.",
		Triggers:      []string{"m?avatar", "m?av", ">av"},
		Usage:         ">av <user>",
		Cooldown:      1,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  base.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run:           m.avatarCommand,
	}
}

func (m *UtilityMod) avatarCommand(msg *base.DiscordMessage) {
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

	msg.ReplyEmbed(&discordgo.MessageEmbed{
		Color: msg.Discord.HighestColor(msg.Message.GuildID, targetUser.ID),
		Title: targetUser.String(),
		Image: &discordgo.MessageEmbedImage{URL: targetUser.AvatarURL("1024")},
	})
}

func NewMemberAvatarCommand(m *UtilityMod) *base.ModCommand {
	return &base.ModCommand{
		Mod:           m,
		Name:          "memberavatar",
		Description:   "Displays a members profile picture. User can be specified. Author is default.",
		Triggers:      []string{"m?memberavatar", "m?mav", ">mav"},
		Usage:         ">av | >av 123123123123",
		Cooldown:      1,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  base.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           m.memberAvatarCommand,
	}
}

func (m *UtilityMod) memberAvatarCommand(msg *base.DiscordMessage) {
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
		msg.Reply(fmt.Sprintf("**%v** doesn't have a server avatar!", targetMember.User.String()))
		return
	}

	msg.ReplyEmbed(&discordgo.MessageEmbed{
		Color: msg.Discord.HighestColor(msg.Message.GuildID, targetMember.User.ID),
		Title: targetMember.User.String(),
		Image: &discordgo.MessageEmbedImage{URL: targetMember.AvatarURL("1024")},
	})
}

func NewUserInfoCommand(m *UtilityMod) *base.ModCommand {
	return &base.ModCommand{
		Mod:           m,
		Name:          "userinfo",
		Description:   "Displays information about a user",
		Triggers:      []string{"m?userinfo"},
		Usage:         "m?userinfo | m?userinfo @user",
		Cooldown:      1,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  base.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           m.userinfoCommand,
	}
}
func (m *UtilityMod) userinfoCommand(msg *base.DiscordMessage) {

	var (
		targetUser   *discordgo.User
		targetMember *discordgo.Member
	)

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
	} else {
		targetMember = msg.Member()
		targetUser = msg.Author()
	}

	createTs := utils.IDToTimestamp(targetUser.ID)
	createDur := time.Since(createTs)

	emb := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("User info | %v", targetUser.String()),
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: targetUser.AvatarURL("512"),
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "ID | Mention",
				Value:  fmt.Sprintf("%v | <@!%v>", targetUser.ID, targetUser.ID),
				Inline: false,
			},
			{
				Name:   "Creation date",
				Value:  fmt.Sprintf("%v | %v day(s) ago", createTs.Format(time.RFC1123), math.Floor(createDur.Hours()/24.0)),
				Inline: false,
			},
		},
	}

	if targetMember != nil {
		joinTs := targetMember.JoinedAt
		joinDur := time.Since(joinTs)

		nick := targetMember.Nick
		if nick == "" {
			nick = "None"
		}

		emb.Color = msg.Discord.HighestColor(msg.Message.GuildID, targetMember.User.ID)
		emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
			Name:   "Join date",
			Value:  fmt.Sprintf("%v | %v day(s) ago", joinTs.Format(time.RFC1123), math.Floor(joinDur.Hours()/24.0)),
			Inline: false,
		})
		emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
			Name:   "Roles",
			Value:  strconv.Itoa(len(targetMember.Roles)),
			Inline: true,
		})
		emb.Fields = append(emb.Fields, &discordgo.MessageEmbedField{
			Name:   "Nickname",
			Value:  nick,
			Inline: true,
		})

	}
	msg.ReplyEmbed(emb)
}
