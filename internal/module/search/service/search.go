package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sync"

	"github.com/bwmarrin/discordgo"
)

type Service struct {
	youtubeToken      string
	openWeatherApiKey string
}

func NewService(ytToken, weatherKey string) *Service {
	return &Service{
		youtubeToken:      ytToken,
		openWeatherApiKey: weatherKey,
	}
}

func (s *Service) request(req *http.Request) ([]byte, error) {
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, errors.New("bad response code")
	}
	return io.ReadAll(res.Body)
}

func parse[V Response](d []byte) (*V, error) {
	var resp V
	err := json.Unmarshal(d, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

const OpenWeatherApiURL = "https://api.openweathermap.org/data/2.5/weather"

func (s *Service) GetWeatherData(query string) (*WeatherResponse, error) {
	params := url.Values{}
	params.Set("q", query)
	params.Set("appid", s.openWeatherApiKey)
	params.Set("units", "metric")

	// this will always work
	req, _ := http.NewRequest("GET", OpenWeatherApiURL, nil)
	req.URL.RawQuery = params.Encode()

	b, err := s.request(req)
	if err != nil {
		return nil, err
	}

	resp, err := parse[WeatherResponse](b)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

var ddgVQDReg = regexp.MustCompile(`vqd='([^']+)'`)
var ddgVQDAltReg = regexp.MustCompile(`vqd=([a-zA-Z0-9-]+)`)

func (s *Service) SearchGoogleImages(query string) ([]string, error) {
	return s.SearchDuckDuckGoImages(query)
}

type ddgImageResponse struct {
	Results []struct {
		Image string `json:"image"`
	} `json:"results"`
}

func (s *Service) SearchDuckDuckGoImages(query string) ([]string, error) {
	homeURL := "https://duckduckgo.com/?q=" + url.QueryEscape(query) + "&iax=images&ia=images"
	req, err := http.NewRequest("GET", homeURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")
	req.Header.Add("Accept-Language", "en-US,en;q=0.9")
	body, err := s.request(req)
	if err != nil {
		return nil, err
	}

	vqd := ""
	if m := ddgVQDReg.FindSubmatch(body); len(m) == 2 {
		vqd = string(m[1])
	} else if m := ddgVQDAltReg.FindSubmatch(body); len(m) == 2 {
		vqd = string(m[1])
	}
	if vqd == "" {
		return nil, errors.New("failed to find ddg vqd token")
	}

	apiURL := "https://duckduckgo.com/i.js?l=us-en&o=json&q=" + url.QueryEscape(query) + "&vqd=" + url.QueryEscape(vqd)
	req2, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req2.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")
	req2.Header.Add("Accept-Language", "en-US,en;q=0.9")
	req2.Header.Add("Referer", "https://duckduckgo.com/")
	req2.Header.Add("Origin", "https://duckduckgo.com")
	req2.Header.Add("Accept", "application/json,text/javascript,*/*;q=0.1")
	req2.Header.Add("Sec-Fetch-Site", "same-origin")
	req2.Header.Add("Sec-Fetch-Mode", "cors")
	req2.Header.Add("Sec-Fetch-Dest", "empty")
	req2.Header.Add("X-Requested-With", "XMLHttpRequest")
	b, err := s.request(req2)
	if err != nil {
		return nil, err
	}

	var resp ddgImageResponse
	if err := json.Unmarshal(b, &resp); err != nil {
		return nil, err
	}

	links := make([]string, 0, len(resp.Results))
	seen := make(map[string]struct{}, len(resp.Results))
	for _, r := range resp.Results {
		if r.Image == "" {
			continue
		}
		if _, ok := seen[r.Image]; ok {
			continue
		}
		seen[r.Image] = struct{}{}
		links = append(links, r.Image)
	}
	return links, nil
}

const YoutubeURL = "https://www.googleapis.com/youtube/v3/search"

func (s *Service) SearchYoutube(query string) ([]string, error) {
	params := url.Values{}
	params.Add("key", s.youtubeToken)
	params.Add("q", query)
	params.Add("type", "video")
	params.Add("part", "snippet")

	req, _ := http.NewRequest("GET", YoutubeURL, nil)
	req.URL.RawQuery = params.Encode()

	b, err := s.request(req)
	if err != nil {
		return nil, err
	}

	resp, err := parse[YoutubeSearchResponse](b)
	if err != nil {
		return nil, err
	}

	var ids []string
	for _, item := range resp.Items {
		ids = append(ids, item.ID.VideoID)
	}
	return ids, nil
}

type ImageSearchCache struct {
	sync.Mutex
	images map[string]*ImageSearch
}

func NewImageSearchCache() *ImageSearchCache {
	return &ImageSearchCache{images: make(map[string]*ImageSearch)}
}

func (c *ImageSearchCache) Set(i *ImageSearch) {
	c.Lock()
	defer c.Unlock()
	c.images[i.BotMsgID()] = i
}
func (c *ImageSearchCache) Get(id string) (*ImageSearch, bool) {
	c.Lock()
	defer c.Unlock()
	if img, ok := c.images[id]; ok {
		return img, true
	}
	return nil, false
}
func (c *ImageSearchCache) Delete(id string) {
	c.Lock()
	defer c.Unlock()
	delete(c.images, id)
}

type ImageSearch struct {
	sync.Mutex
	AuthorMsg  *discordgo.Message
	BotMsg     *discordgo.Message
	ImageLinks []string
	ImageIndex int
	NextID     string
	PrevID     string
	StopID     string
	UpdateCh   chan string
}

func NewImageSearch(a, b *discordgo.Message, links []string, nID, pID, sID string) *ImageSearch {
	return &ImageSearch{
		AuthorMsg:  a,
		BotMsg:     b,
		ImageLinks: links,
		ImageIndex: 0,
		NextID:     nID,
		PrevID:     pID,
		StopID:     sID,
		UpdateCh:   make(chan string),
	}
}

func (i *ImageSearch) AuthorID() string {
	if i.AuthorMsg == nil || i.AuthorMsg.Author == nil {
		return ""
	}
	return i.AuthorMsg.Author.ID
}
func (i *ImageSearch) AuthorMsgID() string {
	if i.AuthorMsg == nil {
		return ""
	}
	return i.AuthorMsg.ID
}
func (i *ImageSearch) BotMsgID() string {
	if i.BotMsg == nil {
		return ""
	}
	return i.BotMsg.ID
}

func (i *ImageSearch) UpdateEmbed(delta int) *discordgo.MessageEmbed {
	newIndex := i.ImageIndex + delta
	if newIndex > len(i.ImageLinks)-1 {
		newIndex = 0
	} else if newIndex < 0 {
		newIndex = len(i.ImageLinks) - 1
	}
	emb := i.BotMsg.Embeds[0]
	emb.Image.URL = i.ImageLinks[newIndex]
	emb.Footer.Text = fmt.Sprintf("Image [ %v / %v ]", newIndex+1, len(i.ImageLinks))
	i.ImageIndex = newIndex
	return emb
}
