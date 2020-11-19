package searchmod

import (
	"encoding/json"
	"fmt"
	"github.com/intrntsrfr/meidov2"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

type SearchMod struct {
	Name string
	sync.Mutex
	cl         chan *meidov2.DiscordMessage
	commands   map[string]meidov2.ModCommand // func(msg *meidov2.DiscordMessage)
	youtubeKey string
}

func New(n string) meidov2.Mod {
	return &SearchMod{
		Name:     n,
		commands: make(map[string]meidov2.ModCommand),
	}
}
func (m *SearchMod) Save() error {
	return nil
}
func (m *SearchMod) Load() error {
	return nil
}
func (m *SearchMod) Commands() map[string]meidov2.ModCommand {
	return m.commands
}
func (m *SearchMod) Hook(b *meidov2.Bot) error {
	m.cl = b.CommandLog
	m.youtubeKey = b.Config.YouTubeKey

	m.RegisterCommand(NewSearchYouTubeCommand(m))

	return nil
}
func (m *SearchMod) RegisterCommand(cmd meidov2.ModCommand) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.commands[cmd.Name()]; ok {
		panic(fmt.Sprintf("command '%v' already exists in %v", cmd.Name(), m.Name))
	}
	m.commands[cmd.Name()] = cmd
}

func (m *SearchMod) Settings(msg *meidov2.DiscordMessage) {

}
func (m *SearchMod) Help(msg *meidov2.DiscordMessage) {

}
func (m *SearchMod) Message(msg *meidov2.DiscordMessage) {
	if msg.Type != meidov2.MessageTypeCreate {
		return
	}
	for _, c := range m.commands {
		go c.Run(msg)
	}
}

type SearchYouTubeCommand struct {
	m       *SearchMod
	Enabled bool
}

func NewSearchYouTubeCommand(m *SearchMod) meidov2.ModCommand {
	return &SearchYouTubeCommand{
		m:       m,
		Enabled: true,
	}
}
func (c *SearchYouTubeCommand) Name() string {
	return "SearchYouTube"
}
func (c *SearchYouTubeCommand) Description() string {
	return "Search for a YouTube Video"
}
func (c *SearchYouTubeCommand) Triggers() []string {
	return []string{"m?youtube", "m?yt"}
}
func (c *SearchYouTubeCommand) Usage() string {
	return "m?yt deez nuts"
}
func (c *SearchYouTubeCommand) Cooldown() int {
	return 10
}
func (c *SearchYouTubeCommand) RequiredPerms() int {
	return 0
}
func (c *SearchYouTubeCommand) RequiresOwner() bool {
	return false
}
func (c *SearchYouTubeCommand) IsEnabled() bool {
	return c.Enabled
}
func (c *SearchYouTubeCommand) Run(msg *meidov2.DiscordMessage) {
	if msg.LenArgs() < 2 || (msg.Args()[0] != "m?youtube" && msg.Args()[0] != "m?yt") {
		return
	}

	c.m.cl <- msg

	query := strings.Join(msg.Args()[1:], " ")
	URI, _ := url.Parse("https://www.googleapis.com/youtube/v3/search")

	params := url.Values{}
	params.Add("key", c.m.youtubeKey)
	params.Add("q", query)
	params.Add("type", "video")
	params.Add("part", "snippet")
	URI.RawQuery = params.Encode()

	req, err := http.NewRequest("GET", URI.String(), nil)
	if err != nil {
		msg.Reply("something wrong happened")
		return
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		msg.Reply("something wrong happened")
		return
	}
	defer res.Body.Close()

	body, _ := ioutil.ReadAll(res.Body)

	result := YoutubeSearchResponse{}

	json.Unmarshal(body, &result)

	if len(result.Items) < 1 {
		msg.Reply("couldnt find anything :(")
	} else {
		id := result.Items[0].ID.VideoID
		msg.Reply("https://youtube.com/watch?v=" + id)
	}
}

type YoutubeSearchResponse struct {
	Items []struct {
		ID struct {
			VideoID string `json:"videoId"`
		} `json:"id"`
	} `json:"items"`
}
