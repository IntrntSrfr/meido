package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
)

var imageReg = regexp.MustCompile(`"(http)s?://([^"])*\.(gif|png|jpg)",`)

type SearchService struct {
	youtubeToken string
	http         *http.Client
}

func NewSearchService(ytToken string) *SearchService {
	return &SearchService{
		youtubeToken: ytToken,
		http:         http.DefaultClient,
	}
}

func (s *SearchService) request(req *http.Request) ([]byte, error) {
	res, err := s.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, errors.New("bad response code")
	}
	return io.ReadAll(res.Body)
}

func (s *SearchService) SearchGoogleImages(query string) ([]string, error) {
	var links []string
	req, err := http.NewRequest("GET", "https://www.google.com/search?tbm=isch&gs_l=img&safe=yes&q="+url.QueryEscape(query), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.193 Safari/537.36")

	body, err := s.request(req)
	if err != nil {
		return nil, err
	}

	matches := imageReg.FindAll(body, -1)

	for _, m := range matches {
		ma := string(m)
		ma = strings.TrimPrefix(ma, `"`)
		ma = strings.TrimSuffix(ma, `",`)

		if strings.Contains(strings.ToLower(ma), "https://www.google.com/logos/doodles") || strings.Contains(strings.ToLower(ma), "https://www.gstatic.com") {
			continue
		}

		links = append(links, ma)
	}
	return links, nil
}

func (s *SearchService) SearchYouTube(query string) ([]string, error) {
	URI, _ := url.Parse("https://www.googleapis.com/youtube/v3/search")
	params := url.Values{}
	params.Add("key", s.youtubeToken)
	params.Add("q", query)
	params.Add("type", "video")
	params.Add("part", "snippet")
	URI.RawQuery = params.Encode()

	req, err := http.NewRequest("GET", URI.String(), nil)
	if err != nil {
		return nil, err
	}

	body, err := s.request(req)
	if err != nil {
		return nil, err
	}

	result := youtubeSearchResponse{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	var ids []string
	for _, item := range result.Items {
		ids = append(ids, item.ID.VideoID)
	}
	return ids, nil
}

type youtubeSearchResponse struct {
	Items []struct {
		ID struct {
			VideoID string `json:"videoId"`
		} `json:"id"`
	} `json:"items"`
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
}

func NewImageSearch(a, b *discordgo.Message, links []string) *ImageSearch {
	return &ImageSearch{
		AuthorMsg:  a,
		BotMsg:     b,
		ImageLinks: links,
		ImageIndex: 0,
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
