package services

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
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
	return ioutil.ReadAll(res.Body)
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

func (s *SearchService) SearchYoutube(query string) ([]string, error) {
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
