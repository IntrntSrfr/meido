package moderationmod

import (
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/base"
	"github.com/intrntsrfr/meido/utils"
	"time"
)

func NewMuteCommand(m *ModerationMod) *base.ModCommand {
	return &base.ModCommand{
		Mod:           m,
		Name:          "mute",
		Description:   "Mutes a member, making them unable to chat or speak. Duration will be 1 day unless something else is specified.",
		Triggers:      []string{"m?mute"},
		Usage:         "m?mute <user> [duration] | m?mute 163454407999094786 1h30m",
		Cooldown:      1,
		CooldownUser:  false,
		RequiredPerms: discordgo.PermissionModerateMembers,
		RequiresOwner: false,
		CheckBotPerms: true,
		AllowedTypes:  base.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           m.muteCommand,
	}
}

func (m *ModerationMod) muteCommand(msg *base.DiscordMessage) {
	if msg.LenArgs() < 2 {
		return
	}

	duration := time.Hour * 24
	if msg.LenArgs() > 2 {
		pDur, err := time.ParseDuration(msg.Args()[2])
		if err != nil {
			_, _ = msg.Reply("invalid time format - I allow hours and minutes! Example: 1h30m")
			return
		}
		if pDur < time.Minute || pDur > time.Hour*24*28 {
			_, _ = msg.Reply("duration is either too short or too long - Minimum 1 minute, max 28 days")
			return
		}
		duration = pDur
	}
	until := time.Now().Add(duration)

	// get the target member
	//targetMember, err := msg.GetMember(1)
	targetMember, err := getMemberAtArg(msg, 1)
	if err != nil {
		return
	}

	if msg.AuthorID() == targetMember.User.ID {
		_, _ = msg.Reply("you cannot mute yourself")
		return
	}

	// check if command hierarchy is valid

	topUserRole := msg.Discord.HighestRolePosition(msg.Message.GuildID, msg.Message.Author.ID)
	topTargetRole := msg.Discord.HighestRolePosition(msg.Message.GuildID, targetMember.User.ID)
	topBotRole := msg.Discord.HighestRolePosition(msg.Message.GuildID, msg.Sess.State.User.ID)

	if topUserRole <= topTargetRole || topBotRole <= topTargetRole {
		_, _ = msg.Reply("no (you can only mute users who are below you and me in the role hierarchy)")
		return
	}

	// just unmute 4head
	err = msg.Discord.Sess.GuildMemberTimeout(msg.GuildID(), msg.AuthorID(), &until)
	if err != nil {
		_, _ = msg.Reply("I was unable to mute that member")
	}
	_, _ = msg.Reply(fmt.Sprintf("muted %v for %v", targetMember.User, duration))
}

func NewUnmuteCommand(m *ModerationMod) *base.ModCommand {
	return &base.ModCommand{
		Mod:           m,
		Name:          "unmute",
		Description:   "Unmutes a member",
		Triggers:      []string{"m?unmute"},
		Usage:         "m?unmute <user>",
		Cooldown:      1,
		CooldownUser:  false,
		RequiredPerms: discordgo.PermissionModerateMembers,
		RequiresOwner: false,
		CheckBotPerms: true,
		AllowedTypes:  base.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           m.unmuteCommand,
	}
}

func (m *ModerationMod) unmuteCommand(msg *base.DiscordMessage) {
	if msg.LenArgs() < 2 {
		return
	}

	// get the target member
	//targetMember, err := msg.GetMember(1)
	targetMember, err := getMemberAtArg(msg, 1)
	if err != nil {
		return
	}

	// user is not timed out
	if targetMember.CommunicationDisabledUntil == nil {
		return
	}

	if msg.AuthorID() == targetMember.User.ID {
		_, _ = msg.Reply("you cannot unmute yourself")
		return
	}

	// check if command hierarchy is valid
	topUserRole := msg.Discord.HighestRolePosition(msg.Message.GuildID, msg.Message.Author.ID)
	topTargetRole := msg.Discord.HighestRolePosition(msg.Message.GuildID, targetMember.User.ID)
	topBotRole := msg.Discord.HighestRolePosition(msg.Message.GuildID, msg.Sess.State.User.ID)

	if topUserRole <= topTargetRole || topBotRole <= topTargetRole {
		_, _ = msg.Reply("no (you can only unmute users who are below you and me in the role hierarchy)")
		return
	}

	// just unmute 4head
	err = msg.Discord.Sess.GuildMemberTimeout(msg.GuildID(), msg.AuthorID(), nil)
	if err != nil {
		_, _ = msg.Reply("I was unable to unmute that member")
	}
	_, _ = msg.Reply(fmt.Sprintf("unmuted %v", targetMember.User))
}

func getMemberAtArg(msg *base.DiscordMessage, index int) (*discordgo.Member, error) {
	if len(msg.Args()) >= index {
		return nil, errors.New("index out of range")
	}
	str := msg.Args()[index]
	userID := utils.TrimUserID(str)
	if !utils.IsNumber(userID) {
		return nil, errors.New(fmt.Sprintf("%s could not be parsed as a number", userID))
	}
	return msg.Discord.Member(msg.GuildID(), str)
}

func getUserAtArg(msg *base.DiscordMessage, index int) (*discordgo.User, error) {
	if len(msg.Args()) >= index {
		return nil, errors.New("index out of range")
	}
	str := msg.Args()[index]
	userID := utils.TrimUserID(str)
	if !utils.IsNumber(userID) {
		return nil, errors.New(fmt.Sprintf("%s could not be parsed as a number", userID))
	}
	return msg.Sess.User(str)
}

func getMemberOrUserAtArg(msg *base.DiscordMessage, index int) (*discordgo.User, error) {
	member, err := getMemberAtArg(msg, index)
	if err != nil {
		return getUserAtArg(msg, index)
	}
	return member.User, nil
}
