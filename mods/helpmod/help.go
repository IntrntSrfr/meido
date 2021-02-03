package helpmod

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido"
	"strings"
	"sync"
)

type HelpMod struct {
	sync.Mutex
	name         string
	commands     map[string]*meido.ModCommand
	allowedTypes meido.MessageType
	allowDMs     bool
	bot          *meido.Bot
}

func New(n string) meido.Mod {
	return &HelpMod{
		name:         n,
		commands:     make(map[string]*meido.ModCommand),
		allowedTypes: meido.MessageTypeCreate,
		allowDMs:     true,
	}
}

func (m *HelpMod) Name() string {
	return m.name
}
func (m *HelpMod) Save() error {
	return nil
}
func (m *HelpMod) Load() error {
	return nil
}
func (m *HelpMod) Passives() []*meido.ModPassive {
	return []*meido.ModPassive{}
}
func (m *HelpMod) Commands() map[string]*meido.ModCommand {
	return m.commands
}
func (m *HelpMod) AllowedTypes() meido.MessageType {
	return m.allowedTypes
}
func (m *HelpMod) AllowDMs() bool {
	return m.allowDMs
}
func (m *HelpMod) Hook(b *meido.Bot) error {
	m.bot = b

	m.RegisterCommand(NewHelpCommand(m))

	return nil
}
func (m *HelpMod) RegisterCommand(cmd *meido.ModCommand) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.commands[cmd.Name]; ok {
		panic(fmt.Sprintf("command '%v' already exists in %v", cmd.Name, m.Name()))
	}
	m.commands[cmd.Name] = cmd
}

func NewHelpCommand(m *HelpMod) *meido.ModCommand {
	return &meido.ModCommand{
		Mod:           m,
		Name:          "help",
		Description:   "Displays helpful things",
		Triggers:      []string{"m?help", "m?h"},
		Usage:         "m?help | m?help about",
		Cooldown:      3,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  meido.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run:           m.helpCommand,
	}
}

func (m *HelpMod) helpCommand(msg *meido.DiscordMessage) {

	emb := &discordgo.MessageEmbed{
		Color: 0xFEFEFE,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "use m?help followed by folder name to see commands for that folder\nuse m?help followed by command name to see specific command help",
		},
	}
	switch msg.LenArgs() {
	case 1:
		desc := strings.Builder{}
		for _, mod := range m.bot.Mods {
			desc.WriteString(fmt.Sprintf("- %v\n", mod.Name()))
		}
		emb.Title = "Meido folders"
		emb.Description = desc.String()
		msg.ReplyEmbed(emb)
	case 2:

		inp := strings.ToLower(msg.Args()[1])

		for _, mod := range m.bot.Mods {
			if mod.Name() == inp {
				// this can maybe be replaced by making a helptext method for every mod so they have more control
				// over what they want to display, if they even want to display anything.

				list := strings.Builder{}

				list.WriteString("\nPassives:\n")
				for _, pas := range mod.Passives() {
					list.WriteString(fmt.Sprintf("- %v\n", pas.Name))
				}

				list.WriteString("\nCommands:\n")
				for _, cmd := range mod.Commands() {
					list.WriteString(fmt.Sprintf("- %v\n", cmd.Name))
				}
				list.WriteString(fmt.Sprintf("\n\nWorks in DMs?: %v", mod.AllowDMs()))

				emb.Title = fmt.Sprintf("commands in %v folder", mod.Name())
				emb.Description = list.String()

				msg.ReplyEmbed(emb)
				return
			}

			for _, pas := range mod.Passives() {
				if pas.Name == inp {

					emb.Title = fmt.Sprintf("Passive - %v", pas.Name)
					emb.Description = "Description:\n" + pas.Description + "\n"
					msg.ReplyEmbed(emb)

					return
				}
			}
			for _, cmd := range mod.Commands() {
				isCmd := false
				if cmd.Name == inp {
					isCmd = true
				}

				for _, trig := range cmd.Triggers {
					if trig == inp {
						isCmd = true
					}
				}

				if !isCmd {
					continue
				}

				emb.Title = fmt.Sprintf("Command - %v", cmd.Name)

				info := strings.Builder{}
				info.WriteString(fmt.Sprintf("\n\nDescription:\n%v", cmd.Description))
				info.WriteString(fmt.Sprintf("\n\nTriggers:\n%v", strings.Join(cmd.Triggers, ", ")))
				info.WriteString(fmt.Sprintf("\n\nUsage:\n%v", cmd.Usage))
				info.WriteString(fmt.Sprintf("\n\nCooldown:\n%v seconds", cmd.Cooldown))
				info.WriteString(fmt.Sprintf("\n\nRequired permissions:\n%v", meido.PermMap[cmd.RequiredPerms]))
				info.WriteString(fmt.Sprintf("\n\nWorks in DMs?:\n%v", cmd.AllowDMs))
				emb.Description = info.String()

				msg.ReplyEmbed(emb)

				return
			}
		}

	default:
		return
	}
}
