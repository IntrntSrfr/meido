package fishing

import (
	"database/sql"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/internal/database"
	"github.com/intrntsrfr/meido/internal/helpers"
	"github.com/intrntsrfr/meido/pkg/mio"
	"github.com/intrntsrfr/meido/pkg/utils"
	"go.uber.org/zap"
	"strings"
)

// FishMod represents the ping mod
type FishMod struct {
	*mio.ModuleBase
	db database.DB
	fs *fishingService
}

// New returns a new FishMod.
func New(bot *mio.Bot, db database.DB, logger *zap.Logger) mio.Module {
	return &FishMod{
		ModuleBase: mio.NewModule(bot, "Fishing", logger.Named("fishing")),
		db:         db,
		fs:         newFishingService(db, logger),
	}
}

// Hook will hook the Module into the Bot.
func (m *FishMod) Hook() error {
	return m.RegisterCommands([]*mio.ModuleCommand{
		newFishCommand(m),
		newAquariumCommand(m),
		newSetFishingSettingsCommand(m),
	})
}

// newFishCommand returns a new fish command.
func newFishCommand(m *FishMod) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:           m,
		Name:          "fish",
		Description:   "Go fishin'",
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
			creature, err := m.fs.goFishing(msg.AuthorID())
			if err != nil {
				m.Log.Error("could not fish", zap.Error(err))
				return
			}
			_, _ = msg.Reply(creature.caption)
		},
	}
}

// newAquariumCommand returns a new Aquarium command.
func newAquariumCommand(m *FishMod) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:           m,
		Name:          "aquarium",
		Description:   "Displays your or someone else's aquarium",
		Triggers:      []string{"m?Aquarium", "m?aq"},
		Usage:         "m?Aquarium <userID>",
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
	if gc, err := m.db.GetGuild(msg.GuildID()); err != nil || msg.ChannelID() != gc.FishingChannelID {
		return
	}
	targetUser := msg.Author()
	if msg.LenArgs() > 1 {
		targetMember, err := msg.GetMemberAtArg(1)
		if err != nil {
			return
		}
		targetUser = targetMember.User
	}

	aq, err := m.db.GetAquarium(targetUser.ID)
	if err != nil && err == sql.ErrNoRows {
		reply := "You have no fish!"
		if targetUser.ID != msg.AuthorID() {
			reply = fmt.Sprintf("%s has no fish!", targetUser.String())
		}
		_, _ = msg.Reply(reply)
		return
	} else if err != nil {
		m.Log.Error("could not get aquarium", zap.Error(err))
		return
	}

	// do this but for each field instead
	var w []string
	w = append(w, fmt.Sprintf("Common 🐟: %v", aq.Common))
	w = append(w, fmt.Sprintf("Uncommon 🐠: %v", aq.Uncommon))
	w = append(w, fmt.Sprintf("Rare 🐡: %v", aq.Rare))
	w = append(w, fmt.Sprintf("Super rare 🦈: %v", aq.SuperRare))
	w = append(w, fmt.Sprintf("Legendary 🎷🦈: %v", aq.Legendary))

	embed := helpers.NewEmbed().
		WithOkColor().
		WithTitle(fmt.Sprintf("%v's aquarium", targetUser.String())).
		WithThumbnail(targetUser.AvatarURL("256")).
		WithDescription(strings.Join(w, "\n"))
	_, _ = msg.ReplyEmbed(embed.Build())
}

// newSetFishingSettingsCommand returns a new fish command.
func newSetFishingSettingsCommand(m *FishMod) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:           m,
		Name:          "fishingsettings",
		Description:   "Fishing settings:\n- Set fishing channel [channelID]",
		Triggers:      []string{"m?settings fishing"},
		Usage:         "m?settings fishing fishingchannel [channelID]",
		Cooldown:      2,
		CooldownUser:  false,
		RequiredPerms: discordgo.PermissionAdministrator,
		RequiresOwner: false,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      false,
		Enabled:       true,
		Run: func(msg *mio.DiscordMessage) {
			if msg.LenArgs() < 2 {
				return
			}
			gc, err := m.db.GetGuild(msg.GuildID())
			if err != nil {
				_, _ = msg.Reply("There was an issue, please try again")
				return
			}

			switch msg.LenArgs() {
			case 2:
				embed := helpers.NewEmbed().
					WithTitle("Fishing settings").
					WithOkColor().
					AddField("Fishing channel", fmt.Sprintf("<#%v>", gc.FishingChannelID), true)
				_, _ = msg.ReplyEmbed(embed.Build())
			case 4:
				switch msg.Args()[2] {
				case "fishingchannel":
					chID := utils.TrimChannelID(msg.Args()[3])
					if !utils.IsNumber(chID) {
						_, _ = msg.Reply("Please provide a valid channel ID")
						return
					}
					if gc.FishingChannelID == chID {
						return
					}
					before := gc.FishingChannelID
					if before == "" {
						before = "Unset"
					}
					gc.FishingChannelID = chID
					if err = m.db.UpdateGuild(gc); err != nil {
						_, _ = msg.Reply("There was an issue, please try again")
						return
					}
					if before == "" {
						_, _ = msg.Reply(fmt.Sprintf("Fishing channel: %v -> <#%v>", before, gc.FishingChannelID))
						return
					}
					_, _ = msg.Reply(fmt.Sprintf("Fishing channel: <#%v> -> <#%v>", before, gc.FishingChannelID))
				}
			}
		},
	}
}
