package administrationmod

import (
	"fmt"
	"github.com/intrntsrfr/meido/internal/base"
	"math/rand"
	"sync"
	"time"
)

// AdministrationMod represents the administration mod
type AdministrationMod struct {
	sync.Mutex
	name         string
	commands     map[string]*base.ModCommand
	allowedTypes base.MessageType
	allowDMs     bool
	bot          *base.Bot
}

// New returns a new AdministrationMod.
func New(n string) base.Mod {
	return &AdministrationMod{
		name:         n,
		commands:     make(map[string]*base.ModCommand),
		allowedTypes: base.MessageTypeCreate,
		allowDMs:     true,
	}
}

// Name returns the name of the mod.
func (m *AdministrationMod) Name() string {
	return m.name
}

// Save saves the mod state to a file.
func (m *AdministrationMod) Save() error {
	return nil
}

// Load loads the mod state from a file.
func (m *AdministrationMod) Load() error {
	return nil
}

// Passives returns the mod passives.
func (m *AdministrationMod) Passives() []*base.ModPassive {
	return []*base.ModPassive{}
}

// Commands returns the mod commands.
func (m *AdministrationMod) Commands() map[string]*base.ModCommand {
	return m.commands
}

// AllowedTypes returns the allowed MessageTypes.
func (m *AdministrationMod) AllowedTypes() base.MessageType {
	return m.allowedTypes
}

// AllowDMs returns whether the mod allows DMs.
func (m *AdministrationMod) AllowDMs() bool {
	return m.allowDMs
}

// Hook will hook the Mod into the Bot.
func (m *AdministrationMod) Hook(b *base.Bot) error {
	m.bot = b

	err := m.Load()
	if err != nil {
		return err
	}

	rand.Seed(time.Now().Unix())

	m.RegisterCommand(NewToggleCommandCommand(m))

	return nil
}

// RegisterCommand registers a ModCommand to the Mod
func (m *AdministrationMod) RegisterCommand(cmd *base.ModCommand) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.commands[cmd.Name]; ok {
		panic(fmt.Sprintf("command '%v' already exists in %v", cmd.Name, m.Name()))
	}
	m.commands[cmd.Name] = cmd
}

// NewToggleCommandCommand returns a new ping command.
func NewToggleCommandCommand(m *AdministrationMod) *base.ModCommand {
	return &base.ModCommand{
		Mod:           m,
		Name:          "togglecommand",
		Description:   "Enables or disables a command. Bot owner only.",
		Triggers:      []string{"m?togglecommand", "m?tc"},
		Usage:         "m?tc ping",
		Cooldown:      2,
		RequiredPerms: 0,
		RequiresOwner: true,
		AllowedTypes:  base.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run:           m.toggleCommandCommand,
	}
}

func (m *AdministrationMod) toggleCommandCommand(msg *base.DiscordMessage) {
	if msg.LenArgs() < 2 {
		return
	}

	if !m.bot.IsOwner(msg) {
		return
	}

	for _, mod := range m.bot.Mods {
		cmd, ok := base.FindCommand(mod, msg.Args())
		if !ok {
			return
		}

		if cmd.Name == "togglecommand" {
			return
		}

		cmd.Enabled = !cmd.Enabled
		if cmd.Enabled {
			msg.Reply(fmt.Sprintf("Enabled command %v", cmd.Name))
		} else {
			msg.Reply(fmt.Sprintf("Disabled command %v", cmd.Name))
		}
	}
}
