package loggermod

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/internal/base"
	"sync"
)

type LoggerMod struct {
	sync.Mutex
	name          string
	commands      map[string]*base.ModCommand
	passives      []*base.ModPassive
	dmLogChannels []string
	allowedTypes  base.MessageType
	allowDMs      bool
}

func New(name string) base.Mod {
	return &LoggerMod{
		name:          name,
		dmLogChannels: []string{},
		commands:      make(map[string]*base.ModCommand),
		passives:      []*base.ModPassive{},
		allowedTypes:  base.MessageTypeCreate,
		allowDMs:      true,
	}
}

func (m *LoggerMod) Name() string {
	return m.name
}
func (m *LoggerMod) Save() error {
	return nil
}
func (m *LoggerMod) Load() error {
	return nil
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
func (m *LoggerMod) Hook(b *base.Bot) error {
	m.dmLogChannels = b.Config.DmLogChannels

	b.Discord.Sess.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		fmt.Println("user:", r.User.String())
		fmt.Println("servers:", len(r.Guilds))
	})

	b.Discord.Sess.AddHandler(func(s *discordgo.Session, g *discordgo.GuildCreate) {
		b.Discord.Sess.RequestGuildMembers(g.ID, "", 0, false)
		fmt.Println("loaded: ", g.Guild.Name, g.MemberCount, len(g.Members))
	})

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
		Description:  "IGNORE THIS | forwards all dms sent to channels found in config",
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
		Color:       0xFEFEFE,
		Title:       fmt.Sprintf("Message from %v", msg.Message.Author.String()),
		Description: msg.Message.Content,
		Footer:      &discordgo.MessageEmbedFooter{Text: msg.Message.Author.ID},
		Timestamp:   string(msg.Message.Timestamp),
	}
	if len(msg.Message.Attachments) > 0 {
		embed.Image = &discordgo.MessageEmbedImage{URL: msg.Message.Attachments[0].URL}
	}

	for _, id := range m.dmLogChannels {
		msg.Sess.ChannelMessageSendEmbed(id, embed)
	}
}
