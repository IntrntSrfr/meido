package searchmod

import (
	"encoding/json"
	"fmt"
	"github.com/intrntsrfr/meido"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

type SearchMod struct {
	sync.Mutex
	name         string
	commands     map[string]*meido.ModCommand
	youtubeKey   string
	allowedTypes meido.MessageType
	allowDMs     bool
}

func New(n string) meido.Mod {
	return &SearchMod{
		name:         n,
		commands:     make(map[string]*meido.ModCommand),
		allowedTypes: meido.MessageTypeCreate,
		allowDMs:     true,
	}
}

func (m *SearchMod) Name() string {
	return m.name
}

func (m *SearchMod) Save() error {
	return nil
}
func (m *SearchMod) Load() error {
	return nil
}
func (m *SearchMod) Passives() []*meido.ModPassive {
	return []*meido.ModPassive{}
}
func (m *SearchMod) Commands() map[string]*meido.ModCommand {
	return m.commands
}
func (m *SearchMod) AllowedTypes() meido.MessageType {
	return m.allowedTypes
}
func (m *SearchMod) AllowDMs() bool {
	return m.allowDMs
}
func (m *SearchMod) Hook(b *meido.Bot) error {
	m.youtubeKey = b.Config.YouTubeKey

	m.RegisterCommand(NewYouTubeCommand(m))

	return nil
}
func (m *SearchMod) RegisterCommand(cmd *meido.ModCommand) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.commands[cmd.Name]; ok {
		panic(fmt.Sprintf("command '%v' already exists in %v", cmd.Name, m.Name()))
	}
	m.commands[cmd.Name] = cmd
}

func NewYouTubeCommand(m *SearchMod) *meido.ModCommand {
	return &meido.ModCommand{
		Mod:           m,
		Name:          "youtube",
		Description:   "Search for a YouTube video",
		Triggers:      []string{"m?youtube", "m?yt"},
		Usage:         "m?yt deez nuts",
		Cooldown:      2,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  meido.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run:           m.youtubeCommand,
	}
}
func (m *SearchMod) youtubeCommand(msg *meido.DiscordMessage) {
	if msg.LenArgs() < 2 {
		return
	}

	query := strings.Join(msg.Args()[1:], " ")
	URI, _ := url.Parse("https://www.googleapis.com/youtube/v3/search")

	params := url.Values{}
	params.Add("key", m.youtubeKey)
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
