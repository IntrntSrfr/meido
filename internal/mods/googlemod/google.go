package googlemod

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/base"
	"github.com/intrntsrfr/meido/utils"
	"strings"
	"sync"
	"time"
)

type GoogleMod struct {
	sync.Mutex
	name              string
	commands          map[string]*base.ModCommand
	allowedTypes      base.MessageType
	allowDMs          bool
	bot               *base.Bot
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

func New(b *base.Bot) base.Mod {
	return &GoogleMod{
		name:              "Google",
		commands:          make(map[string]*base.ModCommand),
		allowedTypes:      base.MessageTypeCreate,
		allowDMs:          true,
		bot:               b,
		activeImgSearches: make(map[string]*ImageSearch),
		deleteImgCh:       make(chan string),
	}
}

func (m *GoogleMod) Name() string {
	return m.name
}
func (m *GoogleMod) Passives() []*base.ModPassive {
	return []*base.ModPassive{}
}
func (m *GoogleMod) Commands() map[string]*base.ModCommand {
	return m.commands
}
func (m *GoogleMod) AllowedTypes() base.MessageType {
	return m.allowedTypes
}
func (m *GoogleMod) AllowDMs() bool {
	return m.allowDMs
}
func (m *GoogleMod) Hook() error {

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

	m.bot.Discord.Sess.AddHandler(m.MessageReactionAddHandler)
	m.bot.Discord.Sess.AddHandler(m.MessageReactionRemoveHandler)

	m.RegisterCommand(NewImageCommand(m))
	return nil
}
func (m *GoogleMod) RegisterCommand(cmd *base.ModCommand) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.commands[cmd.Name]; ok {
		panic(fmt.Sprintf("command '%v' already exists in %v", cmd.Name, m.Name()))
	}
	m.commands[cmd.Name] = cmd
}

func (m *GoogleMod) Message(msg *base.DiscordMessage) {
	if msg.Type() != base.MessageTypeCreate {
		return
	}
	for _, c := range m.commands {
		go c.Run(msg)
	}
}

func NewImageCommand(m *GoogleMod) *base.ModCommand {
	return &base.ModCommand{
		Mod:           m,
		Name:          "image",
		Description:   "Search for an image",
		Triggers:      []string{"m?image", "m?img", "m?im"},
		Usage:         "m?img jeff from 22 jump street",
		Cooldown:      3,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  base.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run:           m.googleCommand,
	}
}

func (m *GoogleMod) googleCommand(msg *base.DiscordMessage) {
	if msg.LenArgs() < 2 {
		return
	}

	query := strings.Join(msg.Args()[1:], " ")
	links := scrape(query)

	if len(links) < 1 {
		msg.Reply("no results")
		return
	}

	reply, err := msg.ReplyEmbed(&discordgo.MessageEmbed{
		Title: "google search",
		Color: utils.ColorInfo,
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
	m.reactionHandler(s, msg.MessageReaction)
}

func (m *GoogleMod) MessageReactionRemoveHandler(s *discordgo.Session, msg *discordgo.MessageReactionRemove) {
	m.reactionHandler(s, msg.MessageReaction)
}

func (m *GoogleMod) reactionHandler(s *discordgo.Session, msg *discordgo.MessageReaction) {
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
	}
}
