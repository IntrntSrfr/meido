package googlemod

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var reg = regexp.MustCompile(`"(http)s?://([^"])*\.(gif|png|jpg)"`)

func scrape(query string) (links []string) {

	req, err := http.NewRequest("GET", "https://www.google.com/search?tbm=isch&gs_l=img&safe=yes&q="+url.QueryEscape(query), nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.193 Safari/537.36")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}

	matches := reg.FindAll(body, -1)

	for _, m := range matches {
		ma := string(m)
		ma = strings.TrimPrefix(ma, `"`)
		ma = strings.TrimSuffix(ma, `"`)

		if strings.Contains(strings.ToLower(ma), "https://www.google.com/logos/doodles") {
			continue
		}

		links = append(links, ma)
	}
	return
}
