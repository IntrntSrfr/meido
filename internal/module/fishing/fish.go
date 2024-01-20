package fishing

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/internal/database"
	iutils "github.com/intrntsrfr/meido/internal/utils"
	"github.com/intrntsrfr/meido/pkg/mio"
	"github.com/intrntsrfr/meido/pkg/utils"
	"go.uber.org/zap"
)

// FishMod represents the ping mod
type FishMod struct {
	*mio.ModuleBase
	db IAquariumDB
	fs *fishingService
}

// New returns a new FishMod.
func New(bot *mio.Bot, db database.DB, logger *zap.Logger) mio.Module {
	return &FishMod{
		ModuleBase: mio.NewModule(bot, "Fishing", logger.Named("fishing")),
		db:         &AquariumDB{db},
	}
}

// Hook will hook the Module into the Bot.
func (m *FishMod) Hook() error {
	var err error
	if m.fs, err = newFishingService(m.db, m.Log); err != nil {
		return err
	}

	return m.RegisterCommands([]*mio.ModuleCommand{
		newFishCommand(m),
		newAquariumCommand(m),
		newSetFishingSettingsCommand(m),
	})
}

// newFishCommand returns a new fish command.
func newFishCommand(m *FishMod) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:              m,
		Name:             "fish",
		Description:      "Go fishin'",
		Triggers:         []string{"m?fish"},
		Usage:            "m?fish",
		Cooldown:         2,
		CooldownScope:    mio.User,
		RequiredPerms:    0,
		RequiresUserType: mio.UserTypeAny,
		AllowedTypes:     mio.MessageTypeCreate,
		AllowDMs:         true,
		IsEnabled:        true,
		Run: func(msg *mio.DiscordMessage) {
			if gc, err := m.db.GetGuild(msg.GuildID()); err != nil || msg.ChannelID() != gc.FishingChannelID {
				return
			}
			creature, err := m.fs.goFishing(msg.AuthorID())
			if err != nil {
				_, _ = msg.Reply("There was an issue, please try again!")
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
		Mod:              m,
		Name:             "aquarium",
		Description:      "Displays your or someone else's aquarium",
		Triggers:         []string{"m?aquarium", "m?aq"},
		Usage:            "m?Aquarium <userID>",
		Cooldown:         3,
		CooldownScope:    mio.User,
		RequiredPerms:    0,
		RequiresUserType: mio.UserTypeAny,
		AllowedTypes:     mio.MessageTypeCreate,
		AllowDMs:         true,
		IsEnabled:        true,
		Run:              m.aquariumCommand,
	}
}

func (m *FishMod) aquariumCommand(msg *mio.DiscordMessage) {
	if gc, err := m.db.GetGuild(msg.GuildID()); err != nil || msg.ChannelID() != gc.FishingChannelID {
		return
	}
	targetUser := msg.Author()
	if len(msg.Args()) > 1 {
		targetMember, err := msg.GetMemberAtArg(1)
		if err != nil {
			return
		}
		targetUser = targetMember.User
	}

	aq, err := m.fs.getOrCreateAquarium(targetUser.ID)
	if err != nil {
		_, _ = msg.Reply("There was an issue, please try again!")
		m.Log.Error("could not get aquarium", zap.Error(err))
		return
	}
	embed := iutils.NewEmbed().
		WithOkColor().
		WithTitle(fmt.Sprintf("%v's aquarium", targetUser.String())).
		WithThumbnail(targetUser.AvatarURL("256"))

	// this is terrible
	embed.AddField("Common", fmt.Sprintf("🐟: %v", aq.Common), true)
	embed.AddField("Uncommon", fmt.Sprintf("🐠: %v", aq.Uncommon), true)
	embed.AddField("Rare", fmt.Sprintf("🐠: %v", aq.Rare), true)
	embed.AddField("Super rare", fmt.Sprintf("🦈: %v", aq.SuperRare), true)
	embed.AddField("Legendary", fmt.Sprintf("🎷🦈: %v", aq.Legendary), true)
	_, _ = msg.ReplyEmbed(embed.Build())
}

// newSetFishingSettingsCommand returns a new fish command.
func newSetFishingSettingsCommand(m *FishMod) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:              m,
		Name:             "fishingsettings",
		Description:      "Fishing settings:\n- Set fishing channel [channelID]",
		Triggers:         []string{"m?settings fishing"},
		Usage:            "m?settings fishing fishingchannel [channelID]",
		Cooldown:         2,
		CooldownScope:    mio.Channel,
		RequiredPerms:    discordgo.PermissionAdministrator,
		CheckBotPerms:    false,
		RequiresUserType: mio.UserTypeAny,
		AllowedTypes:     mio.MessageTypeCreate,
		AllowDMs:         false,
		IsEnabled:        true,
		Run: func(msg *mio.DiscordMessage) {
			if len(msg.Args()) < 2 {
				return
			}
			gc, err := m.db.GetGuild(msg.GuildID())
			if err != nil {
				_, _ = msg.Reply("There was an issue, please try again!")
				return
			}

			switch len(msg.Args()) {
			case 2:
				embed := iutils.NewEmbed().
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
						_, _ = msg.Reply("There was an issue, please try again!")
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
