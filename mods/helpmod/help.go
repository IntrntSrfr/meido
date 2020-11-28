package helpmod

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meidov2"
	"strings"
	"sync"
)

type HelpMod struct {
	sync.Mutex
	name string
	//cl           chan *meidov2.DiscordMessage
	commands     map[string]*meidov2.ModCommand // func(msg *meidov2.DiscordMessage)
	allowedTypes meidov2.MessageType
	allowDMs     bool
	bot          *meidov2.Bot
}

func New(n string) meidov2.Mod {
	return &HelpMod{
		name:         n,
		commands:     make(map[string]*meidov2.ModCommand),
		allowedTypes: meidov2.MessageTypeCreate,
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
func (m *HelpMod) Passives() []*meidov2.ModPassive {
	return []*meidov2.ModPassive{}
}
func (m *HelpMod) Commands() map[string]*meidov2.ModCommand {
	return m.commands
}
func (m *HelpMod) AllowedTypes() meidov2.MessageType {
	return m.allowedTypes
}
func (m *HelpMod) AllowDMs() bool {
	return m.allowDMs
}
func (m *HelpMod) Hook(b *meidov2.Bot) error {
	//m.cl = b.CommandLog
	m.bot = b

	m.RegisterCommand(NewHelpCommand(m))

	return nil
}
func (m *HelpMod) RegisterCommand(cmd *meidov2.ModCommand) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.commands[cmd.Name]; ok {
		panic(fmt.Sprintf("command '%v' already exists in %v", cmd.Name, m.Name()))
	}
	m.commands[cmd.Name] = cmd
}

func NewHelpCommand(m *HelpMod) *meidov2.ModCommand {
	return &meidov2.ModCommand{
		Mod:           m,
		Name:          "help",
		Description:   "Displays helpful things",
		Triggers:      []string{"m?help", "m?h"},
		Usage:         "m?help | m?help about",
		Cooldown:      3,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  meidov2.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run:           m.helpCommand,
	}
}

func (m *HelpMod) helpCommand(msg *meidov2.DiscordMessage) {

	//m.cl <- msg

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
				info.WriteString(fmt.Sprintf("\n\nRequired permissions:\n%v", meidov2.PermMap[cmd.RequiredPerms]))
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
