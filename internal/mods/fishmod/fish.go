package fishmod

import (
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/internal/base"
	"github.com/intrntsrfr/meido/internal/utils"
	"io/ioutil"
	"math/rand"
	"os"
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
	allowedTypes base.MessageType
	allowDMs     bool
	bot          *base.Bot
	Aquariums    map[string]*aquarium
}
type aquarium struct {
	sync.Mutex
	UserID string
	Fish   map[string]int
}

// New returns a new PingMod.
func New(n string) base.Mod {
	return &FishMod{
		name:         n,
		commands:     make(map[string]*base.ModCommand),
		allowedTypes: base.MessageTypeCreate,
		allowDMs:     true,
		Aquariums:    make(map[string]*aquarium),
	}
}

// Name returns the name of the mod.
func (m *FishMod) Name() string {
	return m.name
}

// Save saves the mod state to a file.
func (m *FishMod) Save() error {
	data, err := json.Marshal(m.Aquariums)
	if err != nil {
		return err
	}
	return ioutil.WriteFile("./data/fish", data, os.ModePerm)
}

// Load loads the mod state from a file.
func (m *FishMod) Load() error {
	if _, err := os.Stat("./data/fish"); err != nil {
		return nil
	}
	data, err := ioutil.ReadFile("./data/fish")
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &m.Aquariums)
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

func (a *aquarium) updateUser(f fish, d int) {
	defer a.Unlock()
	a.Lock()

	a.Fish[f.emoji] += d
}

func (m *FishMod) updateAquarium(id string, f fish, d int) {
	defer m.Unlock()
	m.Lock()

	aq, ok := m.Aquariums[id]
	if !ok {
		aq = &aquarium{UserID: id, Fish: make(map[string]int)}
		m.Aquariums[id] = aq
	}
	aq.updateUser(f, d)
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

// NewAquariumCommand returns a new aquarium command.
func NewAquariumCommand(m *FishMod) *base.ModCommand {
	return &base.ModCommand{
		Mod:           m,
		Name:          "aquarium",
		Description:   "Shows your fish collection",
		Triggers:      []string{"m?aquarium", "m?aq"},
		Usage:         "m?aquarium",
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

func (m *FishMod) getAquarium(id string) *aquarium {
	m.Lock()
	defer m.Unlock()
	return m.Aquariums[id]
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

	aq := m.getAquarium(targetUser.ID)
	if aq == nil {
		msg.Reply(fmt.Sprintf("%v has no fish", targetUser.String()))
		return
	}

	var w []string
	aq.Lock()
	for _, f := range fishes {
		w = append(w, fmt.Sprintf("%v: %v", f.emoji, aq.Fish[f.emoji]))
	}
	aq.Unlock()

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%v's aquarium", targetUser.String()),
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
