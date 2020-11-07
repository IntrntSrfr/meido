package utilitymod

import (
	"fmt"
	"github.com/andersfylling/disgord"
	"github.com/bwmarrin/discordgo"
	"github.com/dustin/go-humanize"
	"github.com/intrntsrfr/meidov2"
	"github.com/jmoiron/sqlx"
	"math"
	"runtime"
	"strconv"
	"time"

	_ "github.com/lib/pq"
)

type UtilityMod struct {
	cl        chan *meidov2.DiscordMessage
	commands  []func(msg *meidov2.DiscordMessage)
	startTime time.Time
	db        *sqlx.DB
}

func New() meidov2.Mod {
	return &UtilityMod{
		startTime: time.Now(),
		//cl: make(chan *meidov2.DiscordMessage),
	}
}

func (m *UtilityMod) Save() error {
	return nil
}

func (m *UtilityMod) Load() error {
	return nil
}

func (m *UtilityMod) Settings(msg *meidov2.DiscordMessage) {

}

func (m *UtilityMod) Help(msg *meidov2.DiscordMessage) {

}

func (m *UtilityMod) Commands() []meidov2.ModCommand {
	return nil
}

func (m *UtilityMod) Hook(b *meidov2.Bot) error {
	m.cl = b.CommandLog
	m.db = b.DB

	b.Discord.Sess.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		statusTimer := time.NewTicker(time.Second * 15)
		oldMemCount := 0
		oldSrvCount := 0
		go func() {
			for range statusTimer.C {
				memCount := 0
				srvCount := 0
				for _, sess := range b.Discord.Sessions {
					srvCount++
					for _, g := range sess.State.Guilds {
						memCount += g.MemberCount
					}
				}

				if memCount == oldMemCount && srvCount == oldSrvCount {
					continue
				}

				s.UpdateStatusComplex(discordgo.UpdateStatusData{
					Game: &discordgo.Game{
						Name: fmt.Sprintf("BEING REWORKED, %v MEMBERS, %v SERVERS", memCount, srvCount),
						Type: discordgo.GameTypeGame,
					},
				})
				oldMemCount = memCount
				oldSrvCount = srvCount
			}
		}()
	})

	m.commands = append(m.commands, m.Avatar, m.About, m.Server, m.ServerBanner, m.ServerSplash)

	return nil
}

func (m *UtilityMod) Message(msg *meidov2.DiscordMessage) {
	if msg.Type != meidov2.MessageTypeCreate {
		return
	}
	for _, c := range m.commands {
		go c(msg)
	}
}

func (m *UtilityMod) Avatar(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 1 || (msg.Args()[0] != "m?av" && msg.Args()[0] != ">av") {
		return
	}

	m.cl <- msg

	var targetUser *discordgo.User
	var err error

	if msg.LenArgs() > 1 {
		if len(msg.Message.Mentions) >= 1 {
			targetUser = msg.Message.Mentions[0]
		} else {
			targetUser, err = msg.Sess.User(msg.Args()[1])
			if err != nil {
				return
			}
		}
	} else {
		targetUser, err = msg.Discord.Sess.User(msg.Message.Author.ID)
		if err != nil {
			return
		}
	}

	if targetUser == nil {
		return
	}

	if targetUser.Avatar == "" {
		msg.ReplyEmbed(&discordgo.MessageEmbed{
			Color:       0xC80000,
			Description: fmt.Sprintf("%v has no avatar set.", targetUser.String()),
		})
	} else {
		msg.ReplyEmbed(&discordgo.MessageEmbed{
			Color: msg.HighestColor(msg.Message.GuildID, targetUser.ID),
			Title: targetUser.String(),
			Image: &discordgo.MessageEmbedImage{URL: targetUser.AvatarURL("1024")},
		})
	}
}

// only here in case of disgord
func AvatarURL(u *disgord.User, size int) string {
	a, _ := u.AvatarURL(size, true)
	return a
}

func (m *UtilityMod) Server(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 1 || msg.Args()[0] != "m?server" {
		return
	}
	if msg.IsDM() {
		return
	}
	m.cl <- msg

	g, err := msg.Discord.Sess.State.Guild(msg.Message.GuildID)
	if err != nil {
		msg.Reply("Error getting guild data")
		return
	}

	tc := 0
	vc := 0

	for _, ch := range g.Channels {
		if ch.Type == discordgo.ChannelTypeGuildText {
			tc++
		} else if ch.Type == discordgo.ChannelTypeGuildVoice {
			vc++
		}
	}

	owner, err := msg.Discord.Sess.State.Member(g.ID, g.OwnerID)
	if err != nil {
		msg.Reply("Error getting guild data")
		return
	}

	id, err := strconv.ParseInt(g.ID, 10, 64)
	if err != nil {
		return
	}

	id = ((id >> 22) + 1420070400000) / 1000

	dur := time.Since(time.Unix(id, 0))

	ts := time.Unix(id, 0)

	embed := discordgo.MessageEmbed{
		Color: 0xFFFFFF,
		Author: &discordgo.MessageEmbedAuthor{
			Name: g.Name,
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Owner",
				Value:  fmt.Sprintf("%v\n(%v)", owner.Mention(), owner.User.ID),
				Inline: true,
			},
			{
				Name:  "Creation date",
				Value: fmt.Sprintf("%v\n%v days ago", ts.Format(time.RFC1123), math.Floor(dur.Hours()/24.0)),
			},
			{
				Name:   "Members",
				Value:  fmt.Sprintf("%v members", g.MemberCount),
				Inline: true,
			},
			{
				Name:   "Channels",
				Value:  fmt.Sprintf("Total: %v\nText: %v\nVoice: %v", len(g.Channels), tc, vc),
				Inline: true,
			},
			{
				Name:   "Roles",
				Value:  fmt.Sprintf("%v roles", len(g.Roles)),
				Inline: true,
			},
		},
	}
	if g.Icon != "" {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: fmt.Sprintf("https://cdn.discordapp.com/icons/%v/%v.png", g.ID, g.Icon),
		}
		embed.Author.IconURL = fmt.Sprintf("https://cdn.discordapp.com/icons/%v/%v.png", g.ID, g.Icon)
	}

	msg.ReplyEmbed(&embed)
}

func (m *UtilityMod) About(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 1 || msg.Args()[0] != "m?about" {
		return
	}
	m.cl <- msg

	var (
		totalUsers int
		/*
			totalBots   int
			totalHumans int
		*/
		memory        runtime.MemStats
		totalCommands int
	)
	runtime.ReadMemStats(&memory)
	guilds := msg.Discord.Sess.State.Guilds

	for _, guild := range guilds {
		totalUsers += guild.MemberCount
	}

	uptime := time.Now().Sub(m.startTime)
	err := m.db.Get(&totalCommands, "SELECT COUNT(*) FROM commandlog;")
	if err != nil {
		msg.Reply("Error getting data")
		return
	}

	msg.ReplyEmbed(&discordgo.MessageEmbed{
		Title: "About",
		Color: 0xFEFEFE,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Uptime",
				Value:  uptime.String(),
				Inline: true,
			},
			{
				Name:   "Total commands ran",
				Value:  strconv.Itoa(totalCommands),
				Inline: true,
			},
			{
				Name:   "Guilds",
				Value:  strconv.Itoa(len(guilds)),
				Inline: false,
			},
			{
				Name:   "Users",
				Value:  strconv.Itoa(totalUsers),
				Inline: true,
			},
			{
				Name:   "Current memory use",
				Value:  humanize.Bytes(memory.Alloc),
				Inline: false,
			},
			{
				Name:   "Garbage collected",
				Value:  humanize.Bytes(memory.TotalAlloc - memory.Alloc),
				Inline: true,
			},
		},
	})
}

func (m *UtilityMod) ServerSplash(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() == 0 || msg.Args()[0] != "m?serversplash" {
		return
	}
	if msg.IsDM() {
		return
	}

	m.cl <- msg

	g, err := msg.Discord.Sess.State.Guild(msg.Message.GuildID)
	if err != nil {
		return
	}

	if g.Splash == "" {
		msg.Reply("this server has no splash")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title: g.Name,
		Color: 0xFFFFFF,
		Image: &discordgo.MessageEmbedImage{
			URL: fmt.Sprintf("https://cdn.discordapp.com/splashes/%v/%v.png", g.ID, g.Splash),
		},
	}
	msg.ReplyEmbed(embed)
}

func (m *UtilityMod) ServerBanner(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() == 0 || msg.Args()[0] != "m?serverbanner" {
		return
	}
	if msg.IsDM() {
		return
	}

	m.cl <- msg

	g, err := msg.Discord.Sess.State.Guild(msg.Message.GuildID)
	if err != nil {
		return
	}

	if g.Splash == "" {
		msg.Reply("this server has no banner")
		return
	}

	embed := &discordgo.MessageEmbed{
		Title: g.Name,
		Color: 0xFFFFFF,
		Image: &discordgo.MessageEmbedImage{
			URL: fmt.Sprintf("https://cdn.discordapp.com/banners/%v/%v.png", g.ID, g.Splash),
		},
	}
	msg.ReplyEmbed(embed)
}
