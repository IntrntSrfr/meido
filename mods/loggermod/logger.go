package loggermod

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meidov2"
	"sync"
)

type LoggerMod struct {
	Name string
	sync.Mutex
	cl            chan *meidov2.DiscordMessage
	commands      map[string]meidov2.ModCommand
	passives      []func(*meidov2.DiscordMessage)
	dmLogChannels []string
}

func New(name string) meidov2.Mod {
	return &LoggerMod{
		Name:          name,
		dmLogChannels: []string{},
		commands:      make(map[string]meidov2.ModCommand),
		passives:      []func(*meidov2.DiscordMessage){},
	}
}

func (m *LoggerMod) Save() error {
	return nil
}

func (m *LoggerMod) Load() error {
	return nil
}

func (m *LoggerMod) Settings(msg *meidov2.DiscordMessage) {

}

func (m *LoggerMod) Help(msg *meidov2.DiscordMessage) {

}
func (m *LoggerMod) Commands() map[string]meidov2.ModCommand {
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

	m.passives = append(m.passives, m.ForwardDms)
	return nil
}

func (m *LoggerMod) RegisterCommand(cmd meidov2.ModCommand) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.commands[cmd.Name()]; ok {
		panic(fmt.Sprintf("command '%v' already exists in %v", cmd.Name(), m.Name))
	}
	m.commands[cmd.Name()] = cmd
}

func (m *LoggerMod) Message(msg *meidov2.DiscordMessage) {
	if msg.Type != meidov2.MessageTypeCreate {
		return
	}
	for _, c := range m.commands {
		go c.Run(msg)
	}
	for _, c := range m.passives {
		go c(msg)
	}
}
func (m *LoggerMod) ForwardDms(msg *meidov2.DiscordMessage) {
	if msg.Message.Author.Bot {
		return
	}
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
