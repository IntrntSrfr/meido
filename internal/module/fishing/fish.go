package fishing

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/internal/database"
	"github.com/intrntsrfr/meido/pkg/mio"
	"github.com/intrntsrfr/meido/pkg/utils"
	"github.com/intrntsrfr/meido/pkg/utils/builders"
	"go.uber.org/zap"
)

type module struct {
	*mio.ModuleBase
	db IAquariumDB
	fs *fishingService
}

func New(b *mio.Bot, db database.DB, logger mio.Logger) mio.Module {
	logger = logger.Named("Fishing")
	return &module{
		ModuleBase: mio.NewModule(b, "Fishing", logger),
		db:         &AquariumDB{db},
	}
}

func (m *module) Hook() error {
	var err error
	if m.fs, err = newFishingService(m.db, m.Logger); err != nil {
		return err
	}

	return m.RegisterCommands(
		newFishCommand(m),
		newAquariumCommand(m),
		newSetFishingSettingsCommand(m),
	)
}

func newFishCommand(m *module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:              m,
		Name:             "fish",
		Description:      "Go fishin'",
		Triggers:         []string{"m?fish"},
		Usage:            "m?fish",
		Cooldown:         time.Second * 2,
		CooldownScope:    mio.CooldownScopeUser,
		RequiredPerms:    0,
		RequiresUserType: mio.UserTypeAny,
		AllowedTypes:     mio.MessageTypeCreate,
		AllowDMs:         true,
		Enabled:          true,
		Execute: func(msg *mio.DiscordMessage) {
			if gc, err := m.db.GetGuild(msg.GuildID()); err != nil || msg.ChannelID() != gc.FishingChannelID {
				return
			}
			creature, err := m.fs.goFishing(msg.AuthorID())
			if err != nil {
				_, _ = msg.Reply("There was an issue, please try again!")
				m.Logger.Error("Going fishing failed", zap.Error(err))
				return
			}
			_, _ = msg.Reply(creature.caption)
		},
	}
}

func newAquariumCommand(m *module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:              m,
		Name:             "aquarium",
		Description:      "Displays your or someone else's aquarium",
		Triggers:         []string{"m?aquarium", "m?aq"},
		Usage:            "m?Aquarium <userID>",
		Cooldown:         time.Second * 3,
		CooldownScope:    mio.CooldownScopeUser,
		RequiredPerms:    0,
		RequiresUserType: mio.UserTypeAny,
		AllowedTypes:     mio.MessageTypeCreate,
		AllowDMs:         true,
		Enabled:          true,
		Execute:          m.aquariumCommand,
	}
}

func (m *module) aquariumCommand(msg *mio.DiscordMessage) {
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
		m.Logger.Error("Getting aquarium failed", zap.Error(err))
		return
	}
	embed := builders.NewEmbedBuilder().
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

func newSetFishingSettingsCommand(m *module) *mio.ModuleCommand {
	return &mio.ModuleCommand{
		Mod:              m,
		Name:             "fishingsettings",
		Description:      "Fishing settings:\n- Set fishing channel [channelID]",
		Triggers:         []string{"m?settings fishing"},
		Usage:            "m?settings fishing fishingchannel [channelID]",
		Cooldown:         time.Second * 2,
		CooldownScope:    mio.CooldownScopeChannel,
		RequiredPerms:    discordgo.PermissionAdministrator,
		CheckBotPerms:    false,
		RequiresUserType: mio.UserTypeAny,
		AllowedTypes:     mio.MessageTypeCreate,
		AllowDMs:         false,
		Enabled:          true,
		Execute: func(msg *mio.DiscordMessage) {
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
				embed := builders.NewEmbedBuilder().
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
