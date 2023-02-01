package fishingmod

import (
	"database/sql"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/internal/database"
	"github.com/intrntsrfr/meido/pkg/mio"
	"github.com/intrntsrfr/meido/pkg/utils"
	"math/rand"
	"strconv"
	"strings"
)

// FishMod represents the ping mod
type FishMod struct {
	*mio.ModuleBase
	bot *mio.Bot
	db  *database.PsqlDB
}

// New returns a new PingMod.
func New(b *mio.Bot, db *database.PsqlDB) mio.Module {
	return &FishMod{
		ModuleBase: mio.NewModule("Fishing"),
		bot:        b,
		db:         db,
	}
}

// Hook will hook the Module into the Bot.
func (m *FishMod) Hook() error {
	return m.RegisterCommands([]*mio.ModuleCommand{
		NewFishCommand(m),
		NewAquariumCommand(m),
	})
}

// NewFishCommand returns a new fish command.
func NewFishCommand(m *FishMod) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:           m,
		Name:          "fish",
		Description:   "Fish",
		Triggers:      []string{"m?fish"},
		Usage:         "m?fish",
		Cooldown:      5,
		CooldownUser:  true,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run: func(msg *mio.DiscordMessage) {
			if gc, err := m.db.GetGuild(msg.GuildID()); err != nil || msg.ChannelID() != gc.FishingChannelID {
				return
			}
			fp := pickFish()
			err := m.updateAquarium(msg.Author().ID, fp, 1)
			if err != nil {
				fmt.Println(err)
				fmt.Println("second")
				return
			}

			caption := fp.caption
			if fp.mention {
				caption = msg.Message.Author.Mention() + ", " + fp.caption
			}
			_, _ = msg.Reply(caption)
		},
	}
}

func (m *FishMod) updateAquarium(userID string, f fish, d int) error {
	aq, err := m.db.GetAquarium(userID)
	if err != nil && err == sql.ErrNoRows {
		// if no aquarium found, make one
		err = m.db.CreateAquarium(userID)
		if err != nil {
			return err
		}
	} else if err != nil {
		// everything else we just return
		return err
	}

	switch f.level {
	case common:
		aq.Common += d
	case uncommon:
		aq.Uncommon += d
	case rare:
		aq.Rare += d
	case superRare:
		aq.SuperRare += d
	case legendary:
		aq.Legendary += d
	}

	err = m.db.UpdateAquarium(aq)
	if err != nil {
		return err
	}

	return nil
}

type fish struct {
	level   fishLevel
	caption string
	mention bool
}

var fishes = []fish{
	{common, "You got a common - 🐟", false},
	{uncommon, "You got an uncommon - 🐠", false},
	{rare, "Ohhh, you got a rare! - 🐡", false},
	{superRare, "Woah! you got a super rare! - 🦈", true},
	{legendary, "No way, you got a LEGENDARY!! - 🎷🦈", true},
}

type fishLevel int

const (
	common fishLevel = iota + 1
	uncommon
	rare
	superRare
	legendary
)

func pickFish() fish {
	pick := rand.Intn(1000) + 1
	var fp fish
	if pick <= 800 {
		fp = fishes[0]
	} else if pick <= 940 {
		fp = fishes[1]
	} else if pick <= 990 {
		fp = fishes[2]
	} else if pick <= 999 {
		fp = fishes[3]
	} else {
		fp = fishes[4]
	}
	return fp
}

// NewAquariumCommand returns a new Aquarium command.
func NewAquariumCommand(m *FishMod) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:           m,
		Name:          "Aquarium",
		Description:   "Displays your aquarium",
		Triggers:      []string{"m?Aquarium", "m?aq"},
		Usage:         "m?Aquarium",
		Cooldown:      5,
		CooldownUser:  true,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run:           m.aquariumCommand,
	}
}

func (m *FishMod) aquariumCommand(msg *mio.DiscordMessage) {
	targetUser := msg.Author()
	id := msg.Author().ID
	if msg.LenArgs() > 1 {
		id = utils.TrimUserID(msg.Args()[1])
		_, err := strconv.Atoi(id)
		if err != nil {
			return
		}

		if targetMember, err := msg.Discord.Member(msg.Message.GuildID, id); err == nil {
			targetUser = targetMember.User
		} else if targetUser, err = msg.Discord.Sess.User(id); err != nil {
			return
		}
	}

	aq, err := m.db.GetAquarium(targetUser.ID)
	if err != nil && err == sql.ErrNoRows {
		msg.Reply(fmt.Sprintf("%v has no fish", targetUser.String()))
		return
	} else if err != nil {
		return
	}

	// do this but for each field instead
	var w []string
	w = append(w, fmt.Sprintf("🐟: %v", aq.Common))
	w = append(w, fmt.Sprintf("🐠: %v", aq.Uncommon))
	w = append(w, fmt.Sprintf("🐡: %v", aq.Rare))
	w = append(w, fmt.Sprintf("🦈: %v", aq.SuperRare))
	w = append(w, fmt.Sprintf("🎷🦈: %v", aq.Legendary))

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%v's Aquarium", targetUser.String()),
		Description: strings.Join(w, " | "),
		Color:       utils.ColorLightBlue,
		Author: &discordgo.MessageEmbedAuthor{
			Name:    msg.Author().String(),
			IconURL: msg.Author().AvatarURL("512"),
		},
	}

	_, _ = msg.ReplyEmbed(embed)
}

/*
func (m *FishMod) NewFishTradeCommand() *base.ModuleCommand {
	return &base.ModuleCommand{
		Module:           m,
		Name:          "fishgive",
		Description:   "Give someone a fish",
		Triggers:      []string{"m?fishgive"},
		Usage:         "m?fishgive [user]",
		Cooldown:      5,
		CooldownUser:  true,
		RequiredPerms: 0,
		RequiresOwner: false,
		CheckBotPerms: false,
		AllowedTypes:  base.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run:           nil,
	}
}

func (m *FishMod) newFishGiveCommand(msg *base.DiscordMessage) {
	if msg.LenArgs() < 2 {
		return
	}

	id := msg.Author.UID
	if msg.LenArgs() > 1 {
		id = msg.Args()[1]
		_, err := strconv.Atoi(id)
		if err != nil {
			return
		}
		targetUser, err = msg.Discord.Sess.User(id)
		if err != nil {
			return
		}
	}

	cb, err := m.bot.MakeCallback(msg.ChannelID(), msg.Author.UID)
	if err != nil {
		return
	}

}
*/
