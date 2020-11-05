package loggermod

import (
	"fmt"
	"github.com/andersfylling/disgord"
	"github.com/intrntsrfr/meidov2"
	"github.com/jmoiron/sqlx"
)

type LoggerMod struct {
	cl            chan *meidov2.DiscordMessage
	commands      []func(msg *meidov2.DiscordMessage)
	dmLogChannels []int
}

func New() meidov2.Mod {
	return &LoggerMod{
		//cl:            make(chan *meidov2.DiscordMessage),
		dmLogChannels: []int{},
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
func (m *LoggerMod) Commands() []meidov2.ModCommand {
	return nil
}

func (m *LoggerMod) Hook(b *meidov2.Bot, _ *sqlx.DB, cl chan *meidov2.DiscordMessage) error {
	m.cl = cl

	m.dmLogChannels = b.Config.DmLogChannels

	b.Discord.Client.Gateway().GuildCreate(func(s disgord.Session, g *disgord.GuildCreate) {
		fmt.Println("loaded: ", g.Guild.Name)
	})

	m.commands = append(m.commands, m.ForwardDms)
	return nil
}

func (m *LoggerMod) Message(msg *meidov2.DiscordMessage) {
	if msg.Type != meidov2.MessageTypeCreate {
		return
	}
	for _, c := range m.commands {
		go c(msg)
	}
}
func (m *LoggerMod) ForwardDms(msg *meidov2.DiscordMessage) {

	if msg.Message.Author.Bot {
		return
	}

	if !msg.Message.IsDirectMessage() {
		return
	}

	embed := &disgord.Embed{
		Color:       0xffffff,
		Title:       fmt.Sprintf("Message from %v", msg.Message.Author.Tag()),
		Description: msg.Message.Content,
		Footer:      &disgord.EmbedFooter{Text: msg.Message.Author.ID.String()},
		Timestamp:   msg.Message.Timestamp,
	}
	if len(msg.Message.Attachments) > 0 {
		embed.Image = &disgord.EmbedImage{URL: msg.Message.Attachments[0].URL}
	}

	for _, id := range m.dmLogChannels {
		msg.Sess.SendMsg(disgord.Snowflake(id), embed)
	}
}
