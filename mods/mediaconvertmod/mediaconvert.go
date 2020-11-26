package mediaconvertmod

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meidov2"
	"net/http"
	"path/filepath"
	"sync"
)

type MediaConvertMod struct {
	sync.Mutex
	name         string
	cl           chan *meidov2.DiscordMessage
	commands     map[string]*meidov2.ModCommand // func(msg *meidov2.DiscordMessage)
	passives     []*meidov2.ModPassive
	allowedTypes meidov2.MessageType
	allowDMs     bool
}

func New(n string) meidov2.Mod {
	return &MediaConvertMod{
		name:         n,
		commands:     make(map[string]*meidov2.ModCommand),
		passives:     []*meidov2.ModPassive{},
		allowedTypes: meidov2.MessageTypeCreate,
		allowDMs:     true,
	}
}

func (m *MediaConvertMod) Name() string {
	return m.name
}
func (m *MediaConvertMod) Save() error {
	return nil
}
func (m *MediaConvertMod) Load() error {
	return nil
}
func (m *MediaConvertMod) Passives() []*meidov2.ModPassive {
	return m.passives
}
func (m *MediaConvertMod) Commands() map[string]*meidov2.ModCommand {
	return m.commands
}
func (m *MediaConvertMod) AllowedTypes() meidov2.MessageType {
	return m.allowedTypes
}
func (m *MediaConvertMod) AllowDMs() bool {
	return m.allowDMs
}
func (m *MediaConvertMod) Hook(b *meidov2.Bot) error {
	m.cl = b.CommandLog

	m.passives = append(m.passives, NewJpgLargeConvertPassive(m))

	return nil
}
func (m *MediaConvertMod) RegisterCommand(cmd *meidov2.ModCommand) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.commands[cmd.Name]; ok {
		panic(fmt.Sprintf("command '%v' already exists in %v", cmd.Name, m.Name()))
	}
	m.commands[cmd.Name] = cmd
}

func NewJpgLargeConvertPassive(m *MediaConvertMod) *meidov2.ModPassive {
	return &meidov2.ModPassive{
		Mod:          m,
		Name:         "jpglargeconvert",
		Description:  "Automatically converts jpglarge files to jpg",
		AllowedTypes: meidov2.MessageTypeCreate,
		Enabled:      true,
		Run:          m.jpglargeconvertPassive,
	}
}

func (m *MediaConvertMod) jpglargeconvertPassive(msg *meidov2.DiscordMessage) {
	if len(msg.Message.Attachments) < 1 {
		return
	}

	var files []*discordgo.File

	for _, att := range msg.Message.Attachments {
		if filepath.Ext(att.URL) != ".jpglarge" {
			continue
		}

		res, err := http.Get(att.URL)
		if err != nil {
			continue
		}
		defer res.Body.Close()

		files = append(files, &discordgo.File{
			Name:   "converted.jpg",
			Reader: res.Body,
		})
	}

	msg.Sess.ChannelMessageSendComplex(msg.Message.ChannelID, &discordgo.MessageSend{
		Content: fmt.Sprintf("%v, I converted that to JPG for you", msg.Author.Mention()),
		Files:   files,
	})
}
