package loggermod

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/pkg/mio"
	"github.com/intrntsrfr/meido/pkg/utils"
	"sync"
	"time"
)

type LoggerMod struct {
	sync.Mutex
	name          string
	commands      map[string]*mio.ModCommand
	passives      []*mio.ModPassive
	allowedTypes  mio.MessageType
	allowDMs      bool
	dmLogChannels []string
}

func New(logChs []string) mio.Mod {
	return &LoggerMod{
		name:          "Log",
		commands:      make(map[string]*mio.ModCommand),
		passives:      []*mio.ModPassive{},
		allowedTypes:  mio.MessageTypeCreate,
		allowDMs:      true,
		dmLogChannels: logChs,
	}
}

func (m *LoggerMod) Name() string {
	return m.name
}
func (m *LoggerMod) Passives() []*mio.ModPassive {
	return m.passives
}
func (m *LoggerMod) Commands() map[string]*mio.ModCommand {
	return m.commands
}
func (m *LoggerMod) AllowedTypes() mio.MessageType {
	return m.allowedTypes
}
func (m *LoggerMod) AllowDMs() bool {
	return m.allowDMs
}
func (m *LoggerMod) Hook() error {
	m.passives = append(m.passives, NewForwardDmsPassive(m))
	return nil
}

func (m *LoggerMod) RegisterCommand(cmd *mio.ModCommand) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.commands[cmd.Name]; ok {
		panic(fmt.Sprintf("command '%v' already exists in %v", cmd.Name, m.name))
	}
	m.commands[cmd.Name] = cmd
}

func NewForwardDmsPassive(m *LoggerMod) *mio.ModPassive {
	return &mio.ModPassive{
		Mod:          m,
		Name:         "forwarddms",
		Description:  "Forwards all dms received to a few specific channels specified by the bot owner",
		Enabled:      true,
		AllowedTypes: mio.MessageTypeCreate,
		Run:          m.forwardDmsPassive,
	}
}
func (m *LoggerMod) forwardDmsPassive(msg *mio.DiscordMessage) {
	if !msg.IsDM() {
		return
	}

	embed := &discordgo.MessageEmbed{
		Color:       utils.ColorInfo,
		Title:       fmt.Sprintf("Message from %v", msg.Message.Author.String()),
		Description: msg.Message.Content,
		Footer:      &discordgo.MessageEmbedFooter{Text: msg.Message.Author.ID},
		Timestamp:   msg.Message.Timestamp.Format(time.RFC3339),
	}
	if len(msg.Message.Attachments) > 0 {
		embed.Image = &discordgo.MessageEmbedImage{URL: msg.Message.Attachments[0].URL}
	}

	for _, id := range m.dmLogChannels {
		_, _ = msg.Sess.ChannelMessageSendEmbed(id, embed)
	}
}
