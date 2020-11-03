package utilitymod

import (
	"context"
	"fmt"
	"github.com/andersfylling/disgord"
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

func (m *UtilityMod) Hook(b *meidov2.Bot, db *sqlx.DB, cl chan *meidov2.DiscordMessage) error {
	m.cl = cl
	m.db = db

	b.Discord.Client.On(disgord.EvtReady, func(s disgord.Session, r *disgord.Ready) {
		s.UpdateStatus(&disgord.UpdateStatusPayload{
			Game: &disgord.Activity{
				Type: disgord.ActivityTypeGame,
				Name: "BEING REWORKED, WILL WORK AGAIN SOON",
			},
		})
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

	var targetUser *disgord.User
	var err error

	if msg.LenArgs() > 1 {
		if len(msg.Message.Mentions) >= 1 {
			targetUser = msg.Message.Mentions[0]
		} else {
			id, err := strconv.Atoi(msg.Args()[1])
			if err != nil {
				return
			}
			targetUser, err = msg.Discord.Client.GetUser(context.Background(), disgord.Snowflake(id))
			if err != nil {
				return
			}
		}
	} else {
		targetUser, err = msg.Discord.Client.GetUser(context.Background(), msg.Message.Author.ID)
		if err != nil {
			return
		}
	}

	if targetUser == nil {
		return
	}

	if targetUser.Avatar == "" {
		msg.Reply(&disgord.Embed{
			Color:       0xC80000,
			Description: fmt.Sprintf("%v has no avatar set.", targetUser.Tag()),
		})
	} else {
		msg.Reply(&disgord.Embed{
			Color: msg.HighestColor(msg.Message.GuildID, targetUser.ID),
			Title: targetUser.Tag(),
			Image: &disgord.EmbedImage{URL: AvatarURL(targetUser, 1024)},
		})
	}
}

func AvatarURL(u *disgord.User, size int) string {
	a, _ := u.AvatarURL(size, true)
	return a
}

func (m *UtilityMod) Server(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 1 || msg.Args()[0] != "m?server" {
		return
	}
	if msg.Message.IsDirectMessage() {
		return
	}
	m.cl <- msg

	g, err := msg.Discord.Client.GetGuild(context.Background(), msg.Message.GuildID)
	if err != nil {
		msg.Reply("Error getting guild data")
		return
	}

	tc := 0
	vc := 0

	for _, ch := range g.Channels {
		if ch.Type == disgord.ChannelTypeGuildText {
			tc++
		} else if ch.Type == disgord.ChannelTypeGuildVoice {
			vc++
		}
	}

	owner, err := msg.Discord.Client.GetMember(context.Background(), g.ID, g.OwnerID)
	if err != nil {
		msg.Reply("Error getting guild data")
		return
	}

	c := g.ID.Date()
	dur := time.Since(c)

	embed := disgord.Embed{
		Color: 0xFFFFFF,
		Author: &disgord.EmbedAuthor{
			Name: g.Name,
		},
		Fields: []*disgord.EmbedField{
			{
				Name:   "Owner",
				Value:  fmt.Sprintf("%v\n(%v)", owner.Mention(), owner.UserID),
				Inline: true,
			},
			{
				Name:  "Creation date",
				Value: fmt.Sprintf("%v\n%v days ago", c.Format(time.RFC1123), math.Floor(dur.Hours()/24.0)),
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
		embed.Thumbnail = &disgord.EmbedThumbnail{
			URL: fmt.Sprintf("https://cdn.discordapp.com/icons/%v/%v.png", g.ID, g.Icon),
		}
		embed.Author.IconURL = fmt.Sprintf("https://cdn.discordapp.com/icons/%v/%v.png", g.ID, g.Icon)
	}

	msg.Reply(&embed)
}

func (m *UtilityMod) About(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 1 || msg.Args()[0] != "m?about" {
		return
	}
	m.cl <- msg

	var (
		/*
			totalUsers  int
			totalBots   int
			totalHumans int
		*/
		memory        runtime.MemStats
		totalCommands int
	)
	runtime.ReadMemStats(&memory)
	guilds := len(msg.Discord.Client.GetConnectedGuilds())

	uptime := time.Now().Sub(m.startTime)
	err := m.db.Get(&totalCommands, "SELECT COUNT(*) FROM commandlog;")
	if err != nil {
		msg.Reply("Error getting data")
		return
	}

	msg.Reply(&disgord.Embed{
		Title: "About",
		Color: 0xFFFFFF,
		Fields: []*disgord.EmbedField{
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
				Value:  strconv.Itoa(guilds),
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
				Inline: false,
			},
		},
	})
}
func (m *UtilityMod) ServerSplash(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() == 0 || msg.Args()[0] != "m?serversplash" {
		return
	}
	if msg.Message.IsDirectMessage() {
		return
	}

	m.cl <- msg

	g, err := msg.Discord.Client.GetGuild(context.Background(), msg.Message.GuildID)
	if err != nil {
		return
	}

	if g.Splash == "" {
		msg.Reply("this server has no splash")
		return
	}

	embed := &disgord.Embed{
		Title: g.Name,
		Color: 0xFFFFFF,
		Image: &disgord.EmbedImage{
			URL: fmt.Sprintf("https://cdn.discordapp.com/splashes/%v/%v.png", g.ID, g.Splash),
		},
	}
	msg.Reply(embed)
}

func (m *UtilityMod) ServerBanner(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() == 0 || msg.Args()[0] != "m?serverbanner" {
		return
	}
	if msg.Message.IsDirectMessage() {
		return
	}

	m.cl <- msg

	g, err := msg.Discord.Client.GetGuild(context.Background(), msg.Message.GuildID)
	if err != nil {
		return
	}

	if g.Splash == "" {
		msg.Reply("this server has no banner")
		return
	}

	embed := &disgord.Embed{
		Title: g.Name,
		Color: 0xFFFFFF,
		Image: &disgord.EmbedImage{
			URL: fmt.Sprintf("https://cdn.discordapp.com/banners/%v/%v.png", g.ID, g.Splash),
		},
	}
	msg.Reply(embed)
}
