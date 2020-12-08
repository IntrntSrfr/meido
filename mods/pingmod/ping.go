package pingmod

import (
	"fmt"
	"github.com/intrntsrfr/meidov2"
	"math/rand"
	"sync"
	"time"
)

type PingMod struct {
	sync.Mutex
	name string
	//cl           chan *meidov2.ExecutedCommand
	commands     map[string]*meidov2.ModCommand // func(msg *meidov2.DiscordMessage)
	allowedTypes meidov2.MessageType
	allowDMs     bool
}

func New(n string) meidov2.Mod {
	return &PingMod{
		name:         n,
		commands:     make(map[string]*meidov2.ModCommand),
		allowedTypes: meidov2.MessageTypeCreate,
		allowDMs:     true,
	}
}

func (m *PingMod) Name() string {
	return m.name
}
func (m *PingMod) Save() error {
	return nil
}
func (m *PingMod) Load() error {
	return nil
}
func (m *PingMod) Passives() []*meidov2.ModPassive {
	return []*meidov2.ModPassive{}
}
func (m *PingMod) Commands() map[string]*meidov2.ModCommand {
	return m.commands
}
func (m *PingMod) AllowedTypes() meidov2.MessageType {
	return m.allowedTypes
}
func (m *PingMod) AllowDMs() bool {
	return m.allowDMs
}
func (m *PingMod) Hook(b *meidov2.Bot) error {
	//m.cl = b.CommandLog

	m.RegisterCommand(NewPingCommand(m))
	m.RegisterCommand(NewFishCommand(m))

	return nil
}
func (m *PingMod) RegisterCommand(cmd *meidov2.ModCommand) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.commands[cmd.Name]; ok {
		panic(fmt.Sprintf("command '%v' already exists in %v", cmd.Name, m.Name()))
	}
	m.commands[cmd.Name] = cmd
}

func NewPingCommand(m *PingMod) *meidov2.ModCommand {
	return &meidov2.ModCommand{
		Mod:           m,
		Name:          "ping",
		Description:   "Checks ping",
		Triggers:      []string{"m?ping"},
		Usage:         "m?ping",
		Cooldown:      2,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  meidov2.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run:           m.pingCommand,
	}
}

func (m *PingMod) pingCommand(msg *meidov2.DiscordMessage) {
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

func NewFishCommand(m *PingMod) *meidov2.ModCommand {
	return &meidov2.ModCommand{
		Mod:           m,
		Name:          "fish",
		Description:   "Fish",
		Triggers:      []string{"m?fish"},
		Usage:         "m?fish",
		Cooldown:      0,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  meidov2.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run:           m.fishCommand,
	}
}

var fish = []string{"🐟", "🐠", "🐡", "🦈"}

func (m *PingMod) fishCommand(msg *meidov2.DiscordMessage) {
	rand.Seed(time.Now().Unix())
	msg.Reply(fish[rand.Intn(len(fish))])
}
