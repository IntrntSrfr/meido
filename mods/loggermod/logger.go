package loggermod

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	base2 "github.com/intrntsrfr/meido/base"
	"github.com/intrntsrfr/meido/utils"
	"sync"
)

type LoggerMod struct {
	sync.Mutex
	name          string
	commands      map[string]*base2.ModCommand
	passives      []*base2.ModPassive
	allowedTypes  base2.MessageType
	allowDMs      bool
	dmLogChannels []string
}

func New(name string) base2.Mod {
	return &LoggerMod{
		name:          name,
		commands:      make(map[string]*base2.ModCommand),
		passives:      []*base2.ModPassive{},
		allowedTypes:  base2.MessageTypeCreate,
		allowDMs:      true,
		dmLogChannels: []string{},
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
func (m *LoggerMod) Passives() []*base2.ModPassive {
	return m.passives
}
func (m *LoggerMod) Commands() map[string]*base2.ModCommand {
	return m.commands
}
func (m *LoggerMod) AllowedTypes() base2.MessageType {
	return m.allowedTypes
}
func (m *LoggerMod) AllowDMs() bool {
	return m.allowDMs
}
func (m *LoggerMod) Hook(b *base2.Bot) error {
	m.dmLogChannels = b.Config.DmLogChannels

	b.Discord.Sess.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		fmt.Println("user:", r.User.String())
		fmt.Println("servers:", len(r.Guilds))
	})

	b.Discord.Sess.AddHandler(func(s *discordgo.Session, g *discordgo.GuildCreate) {
		b.Discord.Sess.RequestGuildMembers(g.ID, "", 0, false)
		fmt.Println("loading: ", g.Guild.Name, g.MemberCount, len(g.Members))
	})

	b.Discord.Sess.AddHandler(func(s *discordgo.Session, g *discordgo.GuildMembersChunk) {
		if g.ChunkIndex == g.ChunkCount-1 {
			guild, err := b.Discord.Guild(g.GuildID)
			if err != nil {
				return
			}
			fmt.Println("finished loading " + guild.Name)
		}
	})

	m.passives = append(m.passives, NewForwardDmsPassive(m))
	return nil
}

func (m *LoggerMod) RegisterCommand(cmd *base2.ModCommand) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.commands[cmd.Name]; ok {
		panic(fmt.Sprintf("command '%v' already exists in %v", cmd.Name, m.name))
	}
	m.commands[cmd.Name] = cmd
}

func NewForwardDmsPassive(m *LoggerMod) *base2.ModPassive {
	return &base2.ModPassive{
		Mod:          m,
		Name:         "forwarddms",
		Description:  "IGNORE THIS | forwards all dms sent to channels found in config",
		Enabled:      true,
		AllowedTypes: base2.MessageTypeCreate,
		Run:          m.forwardDmsPassive,
	}
}
func (m *LoggerMod) forwardDmsPassive(msg *base2.DiscordMessage) {
	if !msg.IsDM() {
		return
	}

	embed := &discordgo.MessageEmbed{
		Color:       utils.ColorInfo,
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
