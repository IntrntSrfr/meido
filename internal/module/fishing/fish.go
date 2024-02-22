package fishing

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/internal/database"
	"github.com/intrntsrfr/meido/pkg/mio/bot"
	"github.com/intrntsrfr/meido/pkg/mio/discord"
	"github.com/intrntsrfr/meido/pkg/utils"
	"github.com/intrntsrfr/meido/pkg/utils/builders"
	"go.uber.org/zap"
)

type module struct {
	*bot.ModuleBase
	db IAquariumDB
	fs *fishingService
}

func New(b *bot.Bot, db database.DB, logger *zap.Logger) bot.Module {
	logger = logger.Named("Fishing")
	return &module{
		ModuleBase: bot.NewModule(b, "Fishing", logger),
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

func newFishCommand(m *module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "fish",
		Description:      "Go fishin'",
		Triggers:         []string{"m?fish"},
		Usage:            "m?fish",
		Cooldown:         2,
		CooldownScope:    bot.CooldownScopeUser,
		RequiredPerms:    0,
		RequiresUserType: bot.UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         true,
		Enabled:          true,
		Run: func(msg *discord.DiscordMessage) {
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

func newAquariumCommand(m *module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "aquarium",
		Description:      "Displays your or someone else's aquarium",
		Triggers:         []string{"m?aquarium", "m?aq"},
		Usage:            "m?Aquarium <userID>",
		Cooldown:         3,
		CooldownScope:    bot.CooldownScopeUser,
		RequiredPerms:    0,
		RequiresUserType: bot.UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         true,
		Enabled:          true,
		Run:              m.aquariumCommand,
	}
}

func (m *module) aquariumCommand(msg *discord.DiscordMessage) {
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

func newSetFishingSettingsCommand(m *module) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "fishingsettings",
		Description:      "Fishing settings:\n- Set fishing channel [channelID]",
		Triggers:         []string{"m?settings fishing"},
		Usage:            "m?settings fishing fishingchannel [channelID]",
		Cooldown:         2,
		CooldownScope:    bot.CooldownScopeChannel,
		RequiredPerms:    discordgo.PermissionAdministrator,
		CheckBotPerms:    false,
		RequiresUserType: bot.UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         false,
		Enabled:          true,
		Run: func(msg *discord.DiscordMessage) {
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
