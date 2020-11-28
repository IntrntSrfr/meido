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
	sync.Mutex
	name string
	//cl                chan *meidov2.DiscordMessage
	commands          map[string]*meidov2.ModCommand // func(msg *meidov2.DiscordMessage)
	activeImgSearches map[string]*ImageSearch
	deleteImgCh       chan string
	allowedTypes      meidov2.MessageType
	allowDMs          bool
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
		name:              n,
		commands:          make(map[string]*meidov2.ModCommand),
		activeImgSearches: make(map[string]*ImageSearch),
		deleteImgCh:       make(chan string),
		allowedTypes:      meidov2.MessageTypeCreate,
		allowDMs:          true,
	}
}

func (m *GoogleMod) Name() string {
	return m.name
}
func (m *GoogleMod) Save() error {
	return nil
}
func (m *GoogleMod) Load() error {
	return nil
}
func (m *GoogleMod) Passives() []*meidov2.ModPassive {
	return []*meidov2.ModPassive{}
}
func (m *GoogleMod) Commands() map[string]*meidov2.ModCommand {
	return m.commands
}
func (m *GoogleMod) AllowedTypes() meidov2.MessageType {
	return m.allowedTypes
}
func (m *GoogleMod) AllowDMs() bool {
	return m.allowDMs
}
func (m *GoogleMod) Hook(b *meidov2.Bot) error {
	//m.cl = b.CommandLog

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
func (m *GoogleMod) RegisterCommand(cmd *meidov2.ModCommand) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.commands[cmd.Name]; ok {
		panic(fmt.Sprintf("command '%v' already exists in %v", cmd.Name, m.Name()))
	}
	m.commands[cmd.Name] = cmd
}

func (m *GoogleMod) Message(msg *meidov2.DiscordMessage) {
	if msg.Type != meidov2.MessageTypeCreate {
		return
	}
	for _, c := range m.commands {
		go c.Run(msg)
	}
}

func NewImageCommand(m *GoogleMod) *meidov2.ModCommand {
	return &meidov2.ModCommand{
		Mod:           m,
		Name:          "image",
		Description:   "Search for an image",
		Triggers:      []string{"m?image", "m?img", "m?im"},
		Usage:         "m?img jeff from 22 jump street",
		Cooldown:      3,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  meidov2.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run:           m.googleCommand,
	}
}

func (m *GoogleMod) googleCommand(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 2 {
		return
	}

	//m.cl <- msg

	query := strings.Join(msg.Args()[1:], " ")
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

	m.Lock()
	m.activeImgSearches[reply.ID] = &ImageSearch{
		AuthorMsg:    msg.Message,
		BotMsg:       reply,
		Images:       links,
		CurrentImage: 0,
	}
	m.Unlock()

	go time.AfterFunc(time.Second*30, func() {
		msg.Sess.MessageReactionsRemoveAll(msg.Message.ChannelID, reply.ID)
		m.deleteImgCh <- reply.ID
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
