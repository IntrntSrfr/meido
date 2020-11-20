package googlemod

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meidov2"
	"strings"
	"sync"
	"time"
)

type GoogleMod struct {
	Name string
	sync.Mutex
	cl                chan *meidov2.DiscordMessage
	commands          map[string]meidov2.ModCommand // func(msg *meidov2.DiscordMessage)
	activeImgSearches map[string]*ImageSearch
	deleteImgCh       chan string
}

type ImageSearch struct {
	rw           sync.RWMutex
	AuthorMsg    *discordgo.Message
	BotMsg       *discordgo.Message
	Images       []string
	CurrentImage int
}

func New(n string) meidov2.Mod {
	return &GoogleMod{
		Name:              n,
		commands:          make(map[string]meidov2.ModCommand),
		activeImgSearches: make(map[string]*ImageSearch),
		deleteImgCh:       make(chan string),
	}
}
func (m *GoogleMod) Save() error {
	return nil
}
func (m *GoogleMod) Load() error {
	return nil
}
func (m *GoogleMod) Commands() map[string]meidov2.ModCommand {
	return m.commands
}
func (m *GoogleMod) Hook(b *meidov2.Bot) error {
	m.cl = b.CommandLog

	go func() {
		for {
			select {
			case msg := <-m.deleteImgCh:
				m.Lock()
				delete(m.activeImgSearches, msg)
				m.Unlock()
			}
		}
	}()

	b.Discord.Sess.AddHandler(m.MessageReactionAddHandler)
	b.Discord.Sess.AddHandler(m.MessageReactionRemoveHandler)

	m.RegisterCommand(NewImageCommand(m))

	return nil
}
func (m *GoogleMod) RegisterCommand(cmd meidov2.ModCommand) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.commands[cmd.Name()]; ok {
		panic(fmt.Sprintf("command '%v' already exists in %v", cmd.Name(), m.Name))
	}
	m.commands[cmd.Name()] = cmd
}

func (m *GoogleMod) Settings(msg *meidov2.DiscordMessage) {

}
func (m *GoogleMod) Help(msg *meidov2.DiscordMessage) {

}
func (m *GoogleMod) Message(msg *meidov2.DiscordMessage) {
	if msg.Type != meidov2.MessageTypeCreate {
		return
	}
	for _, c := range m.commands {
		go c.Run(msg)
	}
}

type ImageCommand struct {
	m       *GoogleMod
	Enabled bool
}

func NewImageCommand(m *GoogleMod) meidov2.ModCommand {
	return &ImageCommand{
		m:       m,
		Enabled: true,
	}
}
func (c *ImageCommand) Name() string {
	return "image"
}
func (c *ImageCommand) Description() string {
	return "Search for an image"
}
func (c *ImageCommand) Triggers() []string {
	return []string{"m?image", "m?img"}
}
func (c *ImageCommand) Usage() string {
	return "m?img deez nuts"
}
func (c *ImageCommand) Cooldown() int {
	return 10
}
func (c *ImageCommand) RequiredPerms() int {
	return 0
}
func (c *ImageCommand) RequiresOwner() bool {
	return false
}
func (c *ImageCommand) IsEnabled() bool {
	return c.Enabled
}
func (c *ImageCommand) Run(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 2 || (msg.Args()[0] != "m?image" && msg.Args()[0] != "m?img" && msg.Args()[0] != "m?im") {
		return
	}

	c.m.cl <- msg

	query := strings.Join(msg.Args()[1:], "+")
	links := scrape(query)

	if len(links) < 1 {
		msg.Reply("no results")
		return
	}

	reply, err := msg.ReplyEmbed(&discordgo.MessageEmbed{
		Title: "google search",
		Color: 0xfefefe,
		Author: &discordgo.MessageEmbedAuthor{
			Name:    msg.Message.Author.String(),
			IconURL: msg.Message.Author.AvatarURL("512"),
		},
		Image: &discordgo.MessageEmbedImage{
			URL: links[0],
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("entry [ %v / %v ]", 0, len(links)-1),
		},
	})
	if err != nil {
		return
	}

	msg.Sess.MessageReactionAdd(msg.Message.ChannelID, reply.ID, "⬅")
	msg.Sess.MessageReactionAdd(msg.Message.ChannelID, reply.ID, "➡")
	msg.Sess.MessageReactionAdd(msg.Message.ChannelID, reply.ID, "⏹")

	c.m.Lock()
	c.m.activeImgSearches[reply.ID] = &ImageSearch{
		AuthorMsg:    msg.Message,
		BotMsg:       reply,
		Images:       links,
		CurrentImage: 0,
	}
	c.m.Unlock()

	go time.AfterFunc(time.Second*30, func() {
		msg.Sess.MessageReactionsRemoveAll(msg.Message.ChannelID, reply.ID)
		c.m.deleteImgCh <- reply.ID
	})

}

func (m *GoogleMod) MessageReactionAddHandler(s *discordgo.Session, msg *discordgo.MessageReactionAdd) {
	m.Lock()

	search, ok := m.activeImgSearches[msg.MessageID]
	if !ok {
		m.Unlock()
		return
	}
	m.Unlock()

	if msg.UserID != search.AuthorMsg.Author.ID {
		return
	}
	switch msg.Emoji.Name {
	case "⬅":
		search.rw.Lock()

		emb := search.BotMsg.Embeds[0]

		var index int

		if search.CurrentImage-1 < 0 {
			index = len(search.Images) - 1
		} else {
			index = search.CurrentImage - 1
		}
		emb.Image.URL = search.Images[index]
		emb.Footer.Text = fmt.Sprintf("entry [ %v / %v ]", index, len(search.Images)-1)

		s.ChannelMessageEditEmbed(msg.ChannelID, search.BotMsg.ID, emb)

		search.CurrentImage = index
		search.rw.Unlock()
	case "➡":
		search.rw.Lock()

		emb := search.BotMsg.Embeds[0]

		var index int

		if search.CurrentImage+1 > len(search.Images)-1 {
			index = 0
		} else {
			index = search.CurrentImage + 1
		}

		emb.Image.URL = search.Images[index]
		emb.Footer.Text = fmt.Sprintf("entry [ %v / %v ]", index, len(search.Images)-1)

		s.ChannelMessageEditEmbed(msg.ChannelID, search.BotMsg.ID, emb)

		search.CurrentImage = index
		search.rw.Unlock()

	case "⏹":
		s.ChannelMessageDelete(msg.ChannelID, search.BotMsg.ID)
		s.ChannelMessageDelete(msg.ChannelID, search.AuthorMsg.ID)
		m.deleteImgCh <- search.BotMsg.ID

	default:
	}
}

func (m *GoogleMod) MessageReactionRemoveHandler(s *discordgo.Session, msg *discordgo.MessageReactionRemove) {
	m.Lock()
	search, ok := m.activeImgSearches[msg.MessageID]
	if !ok {
		m.Unlock()
		return
	}
	m.Unlock()

	if msg.UserID != search.AuthorMsg.Author.ID {
		return
	}
	switch msg.Emoji.Name {
	case "⬅":
		search.rw.Lock()

		emb := search.BotMsg.Embeds[0]

		var index int

		if search.CurrentImage-1 < 0 {
			index = len(search.Images) - 1
		} else {
			index = search.CurrentImage - 1
		}
		emb.Image.URL = search.Images[index]
		emb.Footer.Text = fmt.Sprintf("entry [ %v / %v ]", index, len(search.Images)-1)

		s.ChannelMessageEditEmbed(msg.ChannelID, search.BotMsg.ID, emb)

		search.CurrentImage = index
		search.rw.Unlock()
	case "➡":
		search.rw.Lock()

		emb := search.BotMsg.Embeds[0]

		var index int

		if search.CurrentImage+1 > len(search.Images)-1 {
			index = 0
		} else {
			index = search.CurrentImage + 1
		}

		emb.Image.URL = search.Images[index]
		emb.Footer.Text = fmt.Sprintf("entry [ %v / %v ]", index, len(search.Images)-1)

		s.ChannelMessageEditEmbed(msg.ChannelID, search.BotMsg.ID, emb)

		search.CurrentImage = index
		search.rw.Unlock()

	case "⏹":
		s.ChannelMessageDelete(msg.ChannelID, search.BotMsg.ID)
		s.ChannelMessageDelete(msg.ChannelID, search.AuthorMsg.ID)
		m.deleteImgCh <- search.BotMsg.ID

	default:
	}
}
