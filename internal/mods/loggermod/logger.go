package loggermod

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/base"
	"github.com/intrntsrfr/meido/utils"
	"sync"
)

type LoggerMod struct {
	sync.Mutex
	name          string
	commands      map[string]*base.ModCommand
	passives      []*base.ModPassive
	allowedTypes  base.MessageType
	allowDMs      bool
	dmLogChannels []string
}

func New(logChs []string) base.Mod {
	return &LoggerMod{
		name:          "Log",
		commands:      make(map[string]*base.ModCommand),
		passives:      []*base.ModPassive{},
		allowedTypes:  base.MessageTypeCreate,
		allowDMs:      true,
		dmLogChannels: logChs,
	}
}

func (m *LoggerMod) Name() string {
	return m.name
}
func (m *LoggerMod) Passives() []*base.ModPassive {
	return m.passives
}
func (m *LoggerMod) Commands() map[string]*base.ModCommand {
	return m.commands
}
func (m *LoggerMod) AllowedTypes() base.MessageType {
	return m.allowedTypes
}
func (m *LoggerMod) AllowDMs() bool {
	return m.allowDMs
}
func (m *LoggerMod) Hook() error {
	m.passives = append(m.passives, NewForwardDmsPassive(m))
	return nil
}

func (m *LoggerMod) RegisterCommand(cmd *base.ModCommand) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.commands[cmd.Name]; ok {
		panic(fmt.Sprintf("command '%v' already exists in %v", cmd.Name, m.name))
	}
	m.commands[cmd.Name] = cmd
}

func NewForwardDmsPassive(m *LoggerMod) *base.ModPassive {
	return &base.ModPassive{
		Mod:          m,
		Name:         "forwarddms",
		Description:  "Forwards all dms received to a few specific channels specified by the bot owner",
		Enabled:      true,
		AllowedTypes: base.MessageTypeCreate,
		Run:          m.forwardDmsPassive,
	}
}
func (m *LoggerMod) forwardDmsPassive(msg *base.DiscordMessage) {
	if !msg.IsDM() {
		return
	}

	embed := &discordgo.MessageEmbed{
		Color:       utils.ColorInfo,
		Title:       fmt.Sprintf("Message from %v", msg.Message.Author.String()),
		Description: msg.Message.Content,
		Footer:      &discordgo.MessageEmbedFooter{Text: msg.Message.Author.ID},
		Timestamp:   msg.Message.Timestamp.String(),
	}
	if len(msg.Message.Attachments) > 0 {
		embed.Image = &discordgo.MessageEmbedImage{URL: msg.Message.Attachments[0].URL}
	}

	for _, id := range m.dmLogChannels {
		msg.Sess.ChannelMessageSendEmbed(id, embed)
	}
}
