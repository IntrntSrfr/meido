package mediatransform

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"path/filepath"

	"github.com/bwmarrin/discordgo"
	"github.com/intrntsrfr/meido/pkg/mio/bot"
	"github.com/intrntsrfr/meido/pkg/mio/discord"
	"go.uber.org/zap"
)

type MediaTransformMod struct {
	*bot.ModuleBase
}

func New(b *bot.Bot, logger *zap.Logger) bot.Module {
	logger = logger.Named("MediaTransform")
	return &MediaTransformMod{
		ModuleBase: bot.NewModule(b, "MediaTransform", logger),
	}
}

func (m *MediaTransformMod) Hook() error {
	return m.RegisterPassives(newJpgLargeConvertPassive(m))
}

func newJpgLargeConvertPassive(m *MediaTransformMod) *bot.ModulePassive {
	return &bot.ModulePassive{
		Mod:          m,
		Name:         "jpglargeconvert",
		Description:  "Automatically converts jpglarge files to jpg",
		AllowedTypes: discord.MessageTypeCreate,
		Enabled:      true,
		Run:          m.jpgLargeConvertPassive,
	}
}

func (m *MediaTransformMod) jpgLargeConvertPassive(msg *discord.DiscordMessage) {
	if len(msg.Message.Attachments) < 1 {
		return
	}
	var files []*discordgo.File
	for _, att := range msg.Message.Attachments {
		func() {
			if filepath.Ext(att.URL) != ".jpglarge" {
				return
			}
			res, err := http.Get(att.URL)
			if err != nil {
				return
			}
			if res.StatusCode != 200 {
				return
			}
			defer res.Body.Close()
			files = append(files, &discordgo.File{
				Name:   "converted.jpg",
				Reader: res.Body,
			})
		}()
	}
	if len(files) < 1 {
		return
	}
	_, _ = msg.ReplyComplex(&discordgo.MessageSend{
		Content: "I converted that to .jpg for you",
		Files:   files,
		Reference: &discordgo.MessageReference{
			MessageID: msg.ID(),
			ChannelID: msg.ChannelID(),
			GuildID:   msg.GuildID(),
		},
	})
}

func newMediaConvertCommand(m *MediaTransformMod) *bot.ModuleCommand {
	return &bot.ModuleCommand{
		Mod:              m,
		Name:             "mediaconvert",
		Description:      "Converts some media files from one format to another",
		Triggers:         []string{"m?mediaconvert"},
		Usage:            "m?mediaconvert [url] [target format]",
		Cooldown:         30,
		CooldownScope:    bot.CooldownScopeChannel,
		RequiredPerms:    0,
		CheckBotPerms:    false,
		RequiresUserType: bot.UserTypeAny,
		AllowedTypes:     discord.MessageTypeCreate,
		AllowDMs:         true,
		Enabled:          true,
		Run:              m.mediaConvertCommand,
	}
}

func (m *MediaTransformMod) mediaConvertCommand(msg *discord.DiscordMessage) {
	if len(msg.Args()) < 3 {
		return
	}
	// m?mediaconvert link format

	format := msg.Args()[2]
	if format != "mp4" {
		msg.Reply("invalid format")
		return
	}

	src, err := http.Get(msg.RawArgs()[1])
	if err != nil {
		fmt.Println(err)
		msg.Reply("invalid link")
		return
	}
	defer src.Body.Close()

	fmt.Println(src.StatusCode)
	if src.StatusCode != 200 {
		msg.Reply("invalid link")
		return
	}

	p, _ := io.ReadAll(src.Body)
	if http.DetectContentType(p) != "video/webm" {
		msg.Reply("link is not webm")
		return
	}

	if len(p) > 1024*(1024*6) {
		msg.Reply("file too large")
		return
	}

	inp := &bytes.Buffer{}
	inp.Write(p)

	msg.Reply("this might take a while")

	cmd := exec.Command("ffmpeg", "-i", "pipe:0", "-f", "mp4", "-movflags", "frag_keyframe", "pipe:1")
	//cmd := exec.Command("ffmpeg", "-i", "pipe:0", "-f", "mp4", "pipe:1")
	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	cmd.Start()

	func() {
		defer stdin.Close()
		io.Copy(stdin, inp)
	}()

	var out bytes.Buffer
	io.Copy(&out, stdout)

	cmd.Wait()

	fmt.Println(out.Len())

	if out.Len() > 1024*(1024*8) {
		msg.Reply("file too large")
	}

	msg.Discord.Sess.ChannelFileSend(msg.Message.ChannelID, "video.mp4", &out)
}
