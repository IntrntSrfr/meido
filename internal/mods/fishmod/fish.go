package fishmod

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/internal/base"
	"github.com/intrntsrfr/meido/internal/database"
	"github.com/intrntsrfr/meido/internal/utils"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

// FishMod represents the ping mod
type FishMod struct {
	sync.Mutex
	name         string
	commands     map[string]*base.ModCommand
	db           *database.DB
	allowedTypes base.MessageType
	allowDMs     bool
	bot          *base.Bot
}

type Aquarium struct {
	UserID string `db:"user_id"`
	Fish1  int    `db:"fish_1"`
	Fish2  int    `db:"fish_2"`
}

// New returns a new PingMod.
func New(n string) base.Mod {
	return &FishMod{
		name:         n,
		commands:     make(map[string]*base.ModCommand),
		allowedTypes: base.MessageTypeCreate,
		allowDMs:     true,
	}
}

// Name returns the name of the mod.
func (m *FishMod) Name() string {
	return m.name
}

// Save saves the mod state to a file.
func (m *FishMod) Save() error {
	return nil
}

// Load loads the mod state from a file.
func (m *FishMod) Load() error {
	return nil
}

// Passives returns the mod passives.
func (m *FishMod) Passives() []*base.ModPassive {
	return []*base.ModPassive{}
}

// Commands returns the mod commands.
func (m *FishMod) Commands() map[string]*base.ModCommand {
	return m.commands
}

// AllowedTypes returns the allowed MessageTypes.
func (m *FishMod) AllowedTypes() base.MessageType {
	return m.allowedTypes
}

// AllowDMs returns whether the mod allows DMs.
func (m *FishMod) AllowDMs() bool {
	return m.allowDMs
}

// Hook will hook the Mod into the Bot.
func (m *FishMod) Hook(b *base.Bot) error {
	m.bot = b
	m.db = b.DB
	err := m.Load()
	if err != nil {
		return err
	}

	rand.Seed(time.Now().Unix())

	m.RegisterCommand(NewFishCommand(m))
	m.RegisterCommand(NewAquariumCommand(m))

	return nil
}

// RegisterCommand registers a ModCommand to the Mod
func (m *FishMod) RegisterCommand(cmd *base.ModCommand) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.commands[cmd.Name]; ok {
		panic(fmt.Sprintf("command '%v' already exists in %v", cmd.Name, m.Name()))
	}
	m.commands[cmd.Name] = cmd
}

// NewFishCommand returns a new fish command.
func NewFishCommand(m *FishMod) *base.ModCommand {
	return &base.ModCommand{
		Mod:           m,
		Name:          "fish",
		Description:   "Fish",
		Triggers:      []string{"m?fish"},
		Usage:         "m?fish",
		Cooldown:      5,
		CooldownUser:  true,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  base.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run:           m.fishCommand,
	}
}

func (m *FishMod) fishCommand(msg *base.DiscordMessage) {
	fp := pickFish()

	m.updateAquarium(msg.Author().ID, fp, 1)

	caption := fp.caption
	if fp.mention {
		caption = fmt.Sprintf("%v, %v", msg.Message.Author.Mention(), fp.caption)
	}
	msg.Reply(caption)
}

// take in updated aquarium?

func (m *FishMod) updateUser(aq *Aquarium, f fish, d int) error {
	return nil
}

func (m *FishMod) updateAquarium(id string, f fish, d int) error {

	aq := &Aquarium{}
	err := m.db.Get(aq, "SELECT * FROM aquarium WHERE user_id=$1", id)
	if err != nil {
		// if no aquarium found, make one

		aq = &Aquarium{UserID: id}
		_, err2 := m.db.Exec("INSERT INTO aquarium VALUES ($1)", id)
		if err2 != nil {
			return err2
		}
	}

	//m.updateUser(aq)
	return nil
}

type fish struct {
	emoji   string
	caption string
	mention bool
}

var fishes = []fish{
	{"🐟", "You got a common - 🐟", false},
	{"🐠", "You got an uncommon - 🐠", false},
	{"🐡", "Ohhh, you got a rare! - 🐡", false},
	{"🦈", "Woah! you got a super rare! - 🦈", true},
	{"🎷🦈", "YOO YOU GOT A LEGENDARY SAXOPHONE SHARK! - 🎷🦈", true},
}

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
func NewAquariumCommand(m *FishMod) *base.ModCommand {
	return &base.ModCommand{
		Mod:           m,
		Name:          "Aquarium",
		Description:   "Shows your fish collection",
		Triggers:      []string{"m?Aquarium", "m?aq"},
		Usage:         "m?Aquarium",
		Cooldown:      5,
		CooldownUser:  true,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  base.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run:           m.aquariumCommand,
	}
}

func (m *FishMod) getAquarium(id string) (*Aquarium, error) {

	aq := &Aquarium{}
	err := m.db.Get(aq, "SELECT * FROM Aquarium WHERE user_id=$1", id)
	if err != nil {
		return nil, err
	}
	return aq, nil
}

func (m *FishMod) aquariumCommand(msg *base.DiscordMessage) {

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

	aq, err := m.getAquarium(targetUser.ID)
	if err != nil {
		msg.Reply(fmt.Sprintf("%v has no fish", targetUser.String()))
		return
	}

	// do this but for each field instead
	var w []string
	w = append(w, fmt.Sprintf("first fish: %v", aq.Fish1))
	w = append(w, fmt.Sprintf("second fish: %v", aq.Fish2))

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%v's Aquarium", targetUser.String()),
		Description: strings.Join(w, " | "),
		Color:       utils.ColorLightBlue,
		Author: &discordgo.MessageEmbedAuthor{
			Name:    msg.Author().String(),
			IconURL: msg.Author().AvatarURL("512"),
		},
	}

	msg.ReplyEmbed(embed)
}

/*
func (m *FishMod) NewFishTradeCommand() *base.ModCommand {
	return &base.ModCommand{
		Mod:           m,
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

	id := msg.Author.ID
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

	cb, err := m.bot.MakeCallback(msg.ChannelID(), msg.Author.ID)
	if err != nil {
		return
	}

}
*/
