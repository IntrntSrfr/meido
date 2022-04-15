package searchmod

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/base"
	"github.com/intrntsrfr/meido/internal/services"
	"github.com/intrntsrfr/meido/utils"
	"strings"
	"sync"
	"time"
)

type SearchMod struct {
	sync.Mutex
	name         string
	commands     map[string]*base.ModCommand
	allowedTypes base.MessageType
	allowDMs     bool
	bot          *base.Bot
	search       *services.SearchService
	imageCache   *services.ImageSearchCache
}

func New(b *base.Bot, s *services.SearchService) base.Mod {
	return &SearchMod{
		name:         "Search",
		commands:     make(map[string]*base.ModCommand),
		allowedTypes: base.MessageTypeCreate,
		allowDMs:     true,
		bot:          b,
		search:       s,
		imageCache:   services.NewImageSearchCache(),
	}
}

func (m *SearchMod) Name() string {
	return m.name
}
func (m *SearchMod) Passives() []*base.ModPassive {
	return []*base.ModPassive{}
}
func (m *SearchMod) Commands() map[string]*base.ModCommand {
	return m.commands
}
func (m *SearchMod) AllowedTypes() base.MessageType {
	return m.allowedTypes
}
func (m *SearchMod) AllowDMs() bool {
	return m.allowDMs
}
func (m *SearchMod) Hook() error {
	m.RegisterCommand(NewYouTubeCommand(m))
	m.RegisterCommand(NewImageCommand(m))

	m.bot.Discord.AddEventHandler(m.MessageReactionAddHandler)
	m.bot.Discord.AddEventHandler(m.MessageReactionRemoveHandler)

	return nil
}
func (m *SearchMod) RegisterCommand(cmd *base.ModCommand) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.commands[cmd.Name]; ok {
		panic(fmt.Sprintf("command '%v' already exists in %v", cmd.Name, m.Name()))
	}
	m.commands[cmd.Name] = cmd
}

func NewYouTubeCommand(m *SearchMod) *base.ModCommand {
	return &base.ModCommand{
		Mod:           m,
		Name:          "youtube",
		Description:   "Search for a YouTube video",
		Triggers:      []string{"m?youtube", "m?yt"},
		Usage:         "m?yt [query]",
		Cooldown:      2,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  base.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run:           m.youtubeCommand,
	}
}
func (m *SearchMod) youtubeCommand(msg *base.DiscordMessage) {
	if msg.LenArgs() < 2 {
		return
	}

	query := strings.Join(msg.Args()[1:], " ")
	ids, err := m.search.SearchYouTube(query)
	if err != nil {
		msg.Reply("There was an issue, please try again!")
	}

	if len(ids) < 1 {
		msg.Reply("I got no results for that :(")
		return
	}

	msg.Reply("https://youtube.com/watch?v=" + ids[0])
}

func NewImageCommand(m *SearchMod) *base.ModCommand {
	return &base.ModCommand{
		Mod:           m,
		Name:          "image",
		Description:   "Search for an image",
		Triggers:      []string{"m?image", "m?img", "m?im"},
		Usage:         "m?img [query]",
		Cooldown:      2,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  base.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run:           m.googleCommand,
	}
}

func (m *SearchMod) googleCommand(msg *base.DiscordMessage) {
	if msg.LenArgs() < 2 {
		return
	}

	query := strings.Join(msg.Args()[1:], " ")
	links, err := m.search.SearchGoogleImages(query)
	if err != nil {
		msg.Reply("There was an issue, please try again!")
		return
	}

	if len(links) < 1 {
		msg.Reply("I got no results for that :(")
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

	_ = msg.Sess.MessageReactionAdd(msg.Message.ChannelID, reply.ID, "⬅")
	_ = msg.Sess.MessageReactionAdd(msg.Message.ChannelID, reply.ID, "➡")
	_ = msg.Sess.MessageReactionAdd(msg.Message.ChannelID, reply.ID, "⏹")

	m.imageCache.Set(services.NewImageSearch(msg.Message, reply, links))

	go time.AfterFunc(time.Second*30, func() {
		msg.Sess.MessageReactionsRemoveAll(msg.Message.ChannelID, reply.ID)
		m.imageCache.Delete(reply.ID)
	})
}

func (m *SearchMod) MessageReactionAddHandler(s *discordgo.Session, msg *discordgo.MessageReactionAdd) {
	m.reactionHandler(s, msg.MessageReaction)
}

func (m *SearchMod) MessageReactionRemoveHandler(s *discordgo.Session, msg *discordgo.MessageReactionRemove) {
	m.reactionHandler(s, msg.MessageReaction)
}

func (m *SearchMod) reactionHandler(s *discordgo.Session, msg *discordgo.MessageReaction) {
	search, ok := m.imageCache.Get(msg.MessageID)
	if !ok {
		return
	}

	if msg.UserID != search.AuthorID() {
		return
	}

	switch msg.Emoji.Name {
	case "⬅":
		emb := search.UpdateEmbed(-1)
		s.ChannelMessageEditEmbed(msg.ChannelID, search.BotMsgID(), emb)
	case "➡":
		emb := search.UpdateEmbed(1)
		s.ChannelMessageEditEmbed(msg.ChannelID, search.BotMsgID(), emb)
	case "⏹":
		s.ChannelMessageDelete(msg.ChannelID, search.BotMsgID())
		s.ChannelMessageDelete(msg.ChannelID, search.AuthorMsgID())
		m.imageCache.Delete(search.BotMsgID())
	}
}
