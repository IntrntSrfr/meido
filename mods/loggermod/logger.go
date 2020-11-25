package loggermod

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meidov2"
	"sync"
)

type LoggerMod struct {
	sync.Mutex
	name          string
	cl            chan *meidov2.DiscordMessage
	commands      map[string]*meidov2.ModCommand
	passives      []*meidov2.ModPassive
	dmLogChannels []string
	allowedTypes  meidov2.MessageType
	allowDMs      bool
}

func New(name string) meidov2.Mod {
	return &LoggerMod{
		name:          name,
		dmLogChannels: []string{},
		commands:      make(map[string]*meidov2.ModCommand),
		passives:      []*meidov2.ModPassive{},
		allowedTypes:  meidov2.MessageTypeCreate,
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
func (m *LoggerMod) Hook(b *meidov2.Bot) error {
	m.cl = b.CommandLog
	m.dmLogChannels = b.Config.DmLogChannels

	b.Discord.Sess.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		fmt.Println("user:", r.User.String())
		fmt.Println("servers:", len(r.Guilds))
	})

	b.Discord.Sess.AddHandler(func(s *discordgo.Session, g *discordgo.GuildCreate) {
		fmt.Println("loaded: ", g.Guild.Name)
	})

	m.passives = append(m.passives, NewForwardDmsPassive(m))
	return nil
}

func (m *LoggerMod) RegisterCommand(cmd *meidov2.ModCommand) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.commands[cmd.Name]; ok {
		panic(fmt.Sprintf("command '%v' already exists in %v", cmd.Name, m.name))
	}
	m.commands[cmd.Name] = cmd
}

func (m *LoggerMod) Commands() map[string]*meidov2.ModCommand {
	return m.commands
}
func (m *LoggerMod) Passives() []*meidov2.ModPassive {
	return m.passives
}
func (m *LoggerMod) AllowedTypes() meidov2.MessageType {
	return m.allowedTypes
}
func (m *LoggerMod) AllowDMs() bool {
	return m.allowDMs
}

func NewForwardDmsPassive(m *LoggerMod) *meidov2.ModPassive {
	return &meidov2.ModPassive{
		Mod:          m,
		Name:         "forwarddms",
		Description:  "forwards all dms sent to channels found in config",
		Enabled:      true,
		AllowedTypes: meidov2.MessageTypeCreate,
		Run:          m.forwardDmsPassive,
	}
}
func (m *LoggerMod) forwardDmsPassive(msg *meidov2.DiscordMessage) {
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
