package searchmod

import (
	"encoding/json"
	"fmt"
	base2 "github.com/intrntsrfr/meido/base"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

type SearchMod struct {
	sync.Mutex
	name         string
	commands     map[string]*base2.ModCommand
	youtubeKey   string
	allowedTypes base2.MessageType
	allowDMs     bool
}

func New(n string) base2.Mod {
	return &SearchMod{
		name:         n,
		commands:     make(map[string]*base2.ModCommand),
		allowedTypes: base2.MessageTypeCreate,
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
func (m *SearchMod) Passives() []*base2.ModPassive {
	return []*base2.ModPassive{}
}
func (m *SearchMod) Commands() map[string]*base2.ModCommand {
	return m.commands
}
func (m *SearchMod) AllowedTypes() base2.MessageType {
	return m.allowedTypes
}
func (m *SearchMod) AllowDMs() bool {
	return m.allowDMs
}
func (m *SearchMod) Hook(b *base2.Bot) error {
	m.youtubeKey = b.Config.YouTubeKey

	m.RegisterCommand(NewYouTubeCommand(m))

	return nil
}
func (m *SearchMod) RegisterCommand(cmd *base2.ModCommand) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.commands[cmd.Name]; ok {
		panic(fmt.Sprintf("command '%v' already exists in %v", cmd.Name, m.Name()))
	}
	m.commands[cmd.Name] = cmd
}

func NewYouTubeCommand(m *SearchMod) *base2.ModCommand {
	return &base2.ModCommand{
		Mod:           m,
		Name:          "youtube",
		Description:   "Search for a YouTube video",
		Triggers:      []string{"m?youtube", "m?yt"},
		Usage:         "m?yt deez nuts",
		Cooldown:      2,
		RequiredPerms: 0,
		RequiresOwner: false,
		AllowedTypes:  base2.MessageTypeCreate,
		AllowDMs:      true,
		Enabled:       true,
		Run:           m.youtubeCommand,
	}
}
func (m *SearchMod) youtubeCommand(msg *base2.DiscordMessage) {
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
		return
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}

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