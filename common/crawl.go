package crawlfarmCommon

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
)

var hrefRegex *regexp.Regexp = regexp.MustCompile(`(href|src)="(.*?)"`)

func Parse(site Site, referrer, content string) (urls chan Link) {
	var wg sync.WaitGroup
	wg.Add(1)
	urls = make(chan Link)

	go func() {
		defer close(urls)
		defer wg.Done()

		matches := hrefRegex.FindAllStringSubmatch(content, -1)
		for _, m := range matches {
			href := strings.Trim(m[2], " ")
			add := (href != "" && strings.HasPrefix(href, "/") && !strings.HasPrefix(href, "//")) || strings.HasPrefix(href, site.Root)
			if add && !strings.HasPrefix(href, site.Root) {
				href = site.Root + href
			}

			if add {
				urls <- Link{Referrer: referrer, Url: href}
			}
		}
	}()

	return
}

func Fetch(site Site, link Link) (reschan chan Result) {
	var wg sync.WaitGroup
	wg.Add(1)
	reschan = make(chan Result)

	go func() {
		defer close(reschan)
		defer wg.Done()

		uri, parseErr := url.Parse(link.Url)
		if parseErr != nil {
			panic(parseErr)
		}

		headers := make(map[string][]string)
		if site.Headers != nil {
			for k, v := range site.Headers {
				headers[k] = []string{v}
			}
		}

		req := &http.Request{Method: "GET", URL: uri, Header: headers}
		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if strings.Contains(site.Root, req.URL.Host) {
					return nil
				}
				return http.ErrUseLastResponse
			},
		}

		resp, err := client.Do(req)

		if err == nil {
			defer resp.Body.Close()
			if contents, err := ioutil.ReadAll(resp.Body); err == nil {
				text := string(contents)
				reschan <- Result{Link: link, Code: resp.StatusCode, Content: text}
			}
		}
	}()

	return
}
