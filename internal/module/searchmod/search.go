package searchmod

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/internal/helpers"
	"github.com/intrntsrfr/meido/internal/service/search"
	"github.com/intrntsrfr/meido/pkg/mio"
	"github.com/intrntsrfr/meido/pkg/utils"
	"strings"
	"sync"
	"time"
)

type SearchMod struct {
	sync.Mutex
	name         string
	commands     map[string]*mio.ModCommand
	allowedTypes mio.MessageType
	allowDMs     bool
	bot          *mio.Bot
	search       *search.Service
	imageCache   *search.ImageSearchCache
}

func New(b *mio.Bot, s *search.Service) mio.Mod {
	return &SearchMod{
		name:         "Search",
		commands:     make(map[string]*mio.ModCommand),
		allowedTypes: mio.MessageTypeCreate,
		allowDMs:     true,
		bot:          b,
		search:       s,
		imageCache:   search.NewImageSearchCache(),
	}
}

func (m *SearchMod) Name() string {
	return m.name
}
func (m *SearchMod) Passives() []*mio.ModPassive {
	return []*mio.ModPassive{}
}
func (m *SearchMod) Commands() map[string]*mio.ModCommand {
	return m.commands
}
func (m *SearchMod) AllowedTypes() mio.MessageType {
	return m.allowedTypes
}
func (m *SearchMod) AllowDMs() bool {
	return m.allowDMs
}
func (m *SearchMod) Hook() error {
	m.RegisterCommand(NewWeatherCommand(m))
	m.RegisterCommand(NewYouTubeCommand(m))
	m.RegisterCommand(NewImageCommand(m))

	m.bot.Discord.AddEventHandler(m.MessageReactionAddHandler)
	m.bot.Discord.AddEventHandler(m.MessageReactionRemoveHandler)

	return nil
}
func (m *SearchMod) RegisterCommand(cmd *mio.ModCommand) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.commands[cmd.Name]; ok {
		panic(fmt.Sprintf("command '%v' already exists in %v", cmd.Name, m.Name()))
	}
	m.commands[cmd.Name] = cmd
}

func NewWeatherCommand(m *SearchMod) *mio.ModCommand {
	return &mio.ModCommand{
		Mod:           m,
		Name:          "weather",
		Description:   "Finds the weather at a provided location",
		Triggers:      []string{"m?weather"},
		Usage:         "m?weather Oslo",
		Cooldown:      0,
		CooldownUser:  false,
		RequiredPerms: 0,
		RequiresOwner: false,
		CheckBotPerms: false,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run: func(msg *mio.DiscordMessage) {
			// utilize open weather api?
			if msg.LenArgs() < 2 {
				return
			}

			query := strings.TrimSpace(strings.Join(msg.RawArgs()[1:], " "))
			if query == "" {
				return
			}

			d, err := m.search.GetWeatherData(query)
			if err != nil {
				_, _ = msg.Reply("I could not find that city :(")
				return
			}

			f := helpers.CelsiusToFahrenheit
			embed := helpers.NewEmbed().
				WithDescription(fmt.Sprintf("[%v, %v](https://openweathermap.org/city/%v)", d.Name, d.Sys.Country, d.ID)).
				WithOkColor()

			if len(d.Weather) > 0 {
				embed.AddField("Weather", d.Weather[0].Main, true)
			}
			embed.AddField("Temperature", fmt.Sprintf("%.1f°C / %.1f°F", d.Main.Temp, f(d.Main.Temp)), true).
				AddField("Min | Max", fmt.Sprintf("%.1f°C | %.1f°C\n%.1f°F | %.1f°F",
					d.Main.TempMin, d.Main.TempMax, f(d.Main.TempMin), f(d.Main.TempMax)), true).
				AddField("Wind", fmt.Sprintf("%.1f m/s", d.Wind.Speed), true).
				AddField("Sunrise", fmt.Sprintf("<t:%v:R>", d.Sys.Sunrise), true).
				AddField("Sunset", fmt.Sprintf("<t:%v:R>", d.Sys.Sunset), true).
				WithFooter("Powered by openweathermap.org", "")
			_, _ = msg.ReplyEmbed(embed.Build())
		},
	}
}

func NewYouTubeCommand(m *SearchMod) *mio.ModCommand {
	return &mio.ModCommand{
		Mod:           m,
		Name:          "youtube",
		Description:   "Search for a YouTube video",
		Triggers:      []string{"m?youtube", "m?yt"},
		Usage:         "m?yt [query]",
		Cooldown:      2,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run: func(msg *mio.DiscordMessage) {
			if msg.LenArgs() < 2 {
				return
			}

			query := strings.TrimSpace(strings.Join(msg.RawArgs()[1:], " "))
			if query == "" {
				return
			}

			ids, err := m.search.SearchYoutube(query)
			if err != nil {
				_, _ = msg.Reply("There was an issue, please try again!")
				return
			}

			if len(ids) < 1 {
				_, _ = msg.Reply("I got no results for that :(")
				return
			}

			_, _ = msg.Reply("https://youtube.com/watch?v=" + ids[0])
		},
	}
}

func NewImageCommand(m *SearchMod) *mio.ModCommand {
	return &mio.ModCommand{
		Mod:           m,
		Name:          "image",
		Description:   "Search for an image",
		Triggers:      []string{"m?image", "m?img", "m?im"},
		Usage:         "m?img [query]",
		Cooldown:      2,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  mio.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run: func(msg *mio.DiscordMessage) {
			if msg.LenArgs() < 2 {
				return
			}

			query := strings.Join(msg.Args()[1:], " ")
			links, err := m.search.SearchGoogleImages(query)
			if err != nil {
				_, _ = msg.Reply("There was an issue, please try again!")
				return
			}

			if len(links) < 1 {
				_, _ = msg.Reply("I got no results for that :(")
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

			m.imageCache.Set(search.NewImageSearch(msg.Message, reply, links))

			go time.AfterFunc(time.Second*30, func() {
				_ = msg.Sess.MessageReactionsRemoveAll(msg.Message.ChannelID, reply.ID)
				m.imageCache.Delete(reply.ID)
			})
		},
	}
}

func (m *SearchMod) MessageReactionAddHandler(s *discordgo.Session, msg *discordgo.MessageReactionAdd) {
	m.reactionHandler(s, msg.MessageReaction)
}

func (m *SearchMod) MessageReactionRemoveHandler(s *discordgo.Session, msg *discordgo.MessageReactionRemove) {
	m.reactionHandler(s, msg.MessageReaction)
}

func (m *SearchMod) reactionHandler(s *discordgo.Session, msg *discordgo.MessageReaction) {
	srch, ok := m.imageCache.Get(msg.MessageID)
	if !ok {
		return
	}

	if msg.UserID != srch.AuthorID() {
		return
	}

	switch msg.Emoji.Name {
	case "⬅":
		emb := srch.UpdateEmbed(-1)
		_, _ = s.ChannelMessageEditEmbed(msg.ChannelID, srch.BotMsgID(), emb)
	case "➡":
		emb := srch.UpdateEmbed(1)
		_, _ = s.ChannelMessageEditEmbed(msg.ChannelID, srch.BotMsgID(), emb)
	case "⏹":
		_ = s.ChannelMessageDelete(msg.ChannelID, srch.BotMsgID())
		_ = s.ChannelMessageDelete(msg.ChannelID, srch.AuthorMsgID())
		m.imageCache.Delete(srch.BotMsgID())
	}
}
