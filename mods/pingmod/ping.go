package pingmod

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido"
)

// PingMod represents the ping mod
type PingMod struct {
	sync.Mutex
	name         string
	commands     map[string]*meido.ModCommand
	allowedTypes meido.MessageType
	allowDMs     bool
	Aquariums    map[string]*aquarium
}
type aquarium struct {
	sync.Mutex
	UserID string
	Fish   map[string]int
}

// New returns a new PingMod.
func New(n string) meido.Mod {
	return &PingMod{
		name:         n,
		commands:     make(map[string]*meido.ModCommand),
		allowedTypes: meido.MessageTypeCreate,
		allowDMs:     true,
		Aquariums:    make(map[string]*aquarium),
	}
}

// Name returns the name of the mod.
func (m *PingMod) Name() string {
	return m.name
}

// Save saves the mod state to a file.
func (m *PingMod) Save() error {
	data, err := json.Marshal(m.Aquariums)
	if err != nil {
		return err
	}
	return ioutil.WriteFile("./data/fish", data, os.ModePerm)
}

// Load loads the mod state from a file.
func (m *PingMod) Load() error {
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
func (m *PingMod) Passives() []*meido.ModPassive {
	return []*meido.ModPassive{}
}

// Commands returns the mod commands.
func (m *PingMod) Commands() map[string]*meido.ModCommand {
	return m.commands
}

// AllowedTypes returns the allowed MessageTypes.
func (m *PingMod) AllowedTypes() meido.MessageType {
	return m.allowedTypes
}

// AllowDMs returns whether the mod allows DMs.
func (m *PingMod) AllowDMs() bool {
	return m.allowDMs
}

// Hook will hook the Mod into the Bot.
func (m *PingMod) Hook(b *meido.Bot) error {
	err := m.Load()
	if err != nil {
		return err
	}

	rand.Seed(time.Now().Unix())

	m.RegisterCommand(NewPingCommand(m))
	m.RegisterCommand(NewFishCommand(m))
	m.RegisterCommand(NewAquariumCommand(m))
	m.RegisterCommand(NewMonkeyCommand(m))

	return nil
}

// RegisterCommand registers a ModCommand to the Mod
func (m *PingMod) RegisterCommand(cmd *meido.ModCommand) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.commands[cmd.Name]; ok {
		panic(fmt.Sprintf("command '%v' already exists in %v", cmd.Name, m.Name()))
	}
	m.commands[cmd.Name] = cmd
}

// NewPingCommand returns a new ping command.
func NewPingCommand(m *PingMod) *meido.ModCommand {
	return &meido.ModCommand{
		Mod:           m,
		Name:          "ping",
		Description:   "Checks ping",
		Triggers:      []string{"m?ping"},
		Usage:         "m?ping",
		Cooldown:      2,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  meido.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run:           m.pingCommand,
	}
}

func (m *PingMod) pingCommand(msg *meido.DiscordMessage) {
	if msg.LenArgs() < 1 {
		return
	}

	startTime := time.Now()

	first, err := msg.Reply("Ping")
	if err != nil {
		return
	}

	now := time.Now()
	discordLatency := now.Sub(startTime)
	botLatency := now.Sub(msg.TimeReceived)

	msg.Sess.ChannelMessageEdit(msg.Message.ChannelID, first.ID,
		fmt.Sprintf("Pong!\nDiscord delay: %s\nBot delay: %s", discordLatency, botLatency))
}

// NewFishCommand returns a new fish command.
func NewFishCommand(m *PingMod) *meido.ModCommand {
	return &meido.ModCommand{
		Mod:           m,
		Name:          "fish",
		Description:   "Fish",
		Triggers:      []string{"m?fish"},
		Usage:         "m?fish",
		Cooldown:      2,
		CooldownUser:  true,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  meido.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run:           m.fishCommand,
	}
}

func (m *PingMod) fishCommand(msg *meido.DiscordMessage) {
	fp := pickFish()

	m.updateAquarium(msg.Author.ID, fp)

	caption := fp.caption
	if fp.mention {
		caption = fmt.Sprintf("%v, %v", msg.Message.Author.Mention(), fp.caption)
	}
	msg.Reply(caption)
}

func (a *aquarium) updateUser(f fish) {
	defer a.Unlock()
	a.Lock()

	a.Fish[f.emoji]++
}
func (m *PingMod) updateAquarium(id string, f fish) {
	defer m.Unlock()
	m.Lock()

	aq, ok := m.Aquariums[id]
	if !ok {
		aq = &aquarium{UserID: id, Fish: make(map[string]int)}
		m.Aquariums[id] = aq
	}
	aq.updateUser(f)
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

//

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
func NewAquariumCommand(m *PingMod) *meido.ModCommand {
	return &meido.ModCommand{
		Mod:           m,
		Name:          "aquarium",
		Description:   "Shows your fish collection",
		Triggers:      []string{"m?aquarium", "m?aq"},
		Usage:         "m?aquarium",
		Cooldown:      5,
		CooldownUser:  true,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  meido.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run:           m.aquariumCommand,
	}
}

func (m *PingMod) aquariumCommand(msg *meido.DiscordMessage) {

	aq := m.Aquariums[msg.Author.ID]
	if aq == nil {
		msg.Reply("You have no fish.")
		return
	}

	var w []string
	aq.Lock()
	for _, f := range fishes {
		w = append(w, fmt.Sprintf("%v: %v", f.emoji, aq.Fish[f.emoji]))
	}
	aq.Unlock()

	embed := &discordgo.MessageEmbed{
		Title:       "Your aquarium",
		Description: strings.Join(w, " | "),
		Color:       0x00bbe0,
		Author: &discordgo.MessageEmbedAuthor{
			Name:    msg.Author.String(),
			IconURL: msg.Author.AvatarURL("512"),
		},
	}

	msg.ReplyEmbed(embed)
}

// NewMonkeyCommand returns a new monkey command.
func NewMonkeyCommand(m *PingMod) *meido.ModCommand {
	return &meido.ModCommand{
		Mod:           m,
		Name:          "monkey",
		Description:   "Monkey",
		Triggers:      []string{"m?monkey", "m?monke"},
		Usage:         "m?monkey",
		Cooldown:      0,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  meido.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run:           m.monkeyCommand,
	}
}

func (m *PingMod) monkeyCommand(msg *meido.DiscordMessage) {
	rand.Seed(time.Now().Unix())
	msg.Reply(monkeys[rand.Intn(len(monkeys))])
}

var monkeys = []string{"🐒", "🐒💨", "🔫🐒", "🎷🐒", "\U0001F9FB🖊️🐒", "🐒🚿", "🐒\n🚽"}
