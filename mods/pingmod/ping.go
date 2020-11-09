package pingmod

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meidov2"
	"sync"
	"time"
)

type PingMod struct {
	Name string
	sync.Mutex
	cl       chan *meidov2.DiscordMessage
	commands map[string]meidov2.ModCommand // func(msg *meidov2.DiscordMessage)
}

func New(n string) meidov2.Mod {
	return &PingMod{
		Name:     n,
		commands: make(map[string]meidov2.ModCommand),
	}
}
func (m *PingMod) Save() error {
	return nil
}
func (m *PingMod) Load() error {
	return nil
}
func (m *PingMod) Commands() map[string]meidov2.ModCommand {
	return m.commands
}
func (m *PingMod) Hook(b *meidov2.Bot) error {
	m.cl = b.CommandLog

	b.Discord.Sess.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		fmt.Println("user:", r.User.String())
		fmt.Println("servers:", len(r.Guilds))
	})

	m.RegisterCommand(NewPingCommand(m))

	return nil
}
func (m *PingMod) RegisterCommand(cmd meidov2.ModCommand) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.commands[cmd.Name()]; ok {
		panic(fmt.Sprintf("command '%v' already exists in %v", cmd.Name(), m.Name))
	}
	m.commands[cmd.Name()] = cmd
}

func (m *PingMod) Settings(msg *meidov2.DiscordMessage) {

}
func (m *PingMod) Help(msg *meidov2.DiscordMessage) {

}
func (m *PingMod) Message(msg *meidov2.DiscordMessage) {
	if msg.Type != meidov2.MessageTypeCreate {
		return
	}
	for _, c := range m.commands {
		go c.Run(msg)
	}
}

type PingCommand struct {
	m       *PingMod
	Enabled bool
}

func NewPingCommand(m *PingMod) *PingCommand {
	return &PingCommand{
		m:       m,
		Enabled: true,
	}
}
func (c *PingCommand) Name() string {
	return "Ping"
}
func (c *PingCommand) Description() string {
	return "Checks the bots ping against Discord"
}
func (c *PingCommand) Triggers() []string {
	return []string{"m?ping"}
}
func (c *PingCommand) Usage() string {
	return "m?ping"
}
func (c *PingCommand) Cooldown() int {
	return 10
}
func (c *PingCommand) RequiredPerms() int {
	return 0
}
func (c *PingCommand) RequiresOwner() bool {
	return false
}
func (c *PingCommand) IsEnabled() bool {
	return c.Enabled
}
func (c *PingCommand) Run(msg *meidov2.DiscordMessage) {
	if msg.Message.Content != "m?ping" {
		return
	}

	c.m.cl <- msg

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
