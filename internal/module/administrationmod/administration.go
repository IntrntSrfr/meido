package administrationmod

import (
	"fmt"
	"github.com/intrntsrfr/meido/pkg/mio"
	"sync"
)

// AdministrationMod represents the administration mod
type AdministrationMod struct {
	sync.Mutex
	name         string
	commands     map[string]*mio.ModCommand
	allowedTypes mio.MessageType
	allowDMs     bool
	bot          *mio.Bot
}

// New returns a new AdministrationMod.
func New(b *mio.Bot) mio.Mod {
	return &AdministrationMod{
		name:         "Administration",
		commands:     make(map[string]*mio.ModCommand),
		allowedTypes: mio.MessageTypeCreate,
		allowDMs:     true,
		bot:          b,
	}
}

// Name returns the name of the mod.
func (m *AdministrationMod) Name() string {
	return m.name
}

// Passives returns the mod passives.
func (m *AdministrationMod) Passives() []*mio.ModPassive {
	return []*mio.ModPassive{}
}

// Commands returns the mod commands.
func (m *AdministrationMod) Commands() map[string]*mio.ModCommand {
	return m.commands
}

// AllowedTypes returns the allowed MessageTypes.
func (m *AdministrationMod) AllowedTypes() mio.MessageType {
	return m.allowedTypes
}

// AllowDMs returns whether the mod allows DMs.
func (m *AdministrationMod) AllowDMs() bool {
	return m.allowDMs
}

// Hook will hook the Mod into the Bot.
func (m *AdministrationMod) Hook() error {
	m.RegisterCommand(NewToggleCommandCommand(m))
	return nil
}

// RegisterCommand registers a ModCommand to the Mod
func (m *AdministrationMod) RegisterCommand(cmd *mio.ModCommand) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.commands[cmd.Name]; ok {
		panic(fmt.Sprintf("command '%v' already exists in %v", cmd.Name, m.Name()))
	}
	m.commands[cmd.Name] = cmd
}

// NewToggleCommandCommand returns a new ping command.
func NewToggleCommandCommand(m *AdministrationMod) *mio.ModCommand {
	return &mio.ModCommand{
		Mod:           m,
		Name:          "togglecommand",
		Description:   "Enables or disables a command. Bot owner only.",
		Triggers:      []string{"m?togglecommand", "m?tc"},
		Usage:         "m?tc ping",
		Cooldown:      2,
		RequiredPerms: 0,
		RequiresOwner: true,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run:           m.toggleCommandCommand,
	}
}

func (m *AdministrationMod) toggleCommandCommand(msg *mio.DiscordMessage) {
	if msg.LenArgs() < 2 {
		return
	}

	if !msg.Discord.IsBotOwner(msg) {
		return
	}

	for _, mod := range m.bot.Mods {
		cmd, ok := mio.FindCommand(mod, msg.Args())
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
