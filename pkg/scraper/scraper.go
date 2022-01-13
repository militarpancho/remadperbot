package scraper

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/anaskhan96/soup"
)

type ArticleInfo struct {
	Metadata []string
	Title    string
	Img      []byte
	Url      string
}

func ExtractArticleInfo(url string, download_image bool) *ArticleInfo {
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: transCfg}
	soup.Header("User-Agent", "")
	resp, err := soup.GetWithClient(url, client)
	if err != nil {
		fmt.Errorf("error making the request to url %s : %w", url, err)
	}
	if ProductFound(url) {
		doc := soup.HTMLParse(resp)
		title := doc.Find("h1")
		var thumbnail_body []soup.Root = doc.FindStrict("div", "class", "thumbnail-body").FindAll("p")
		var thumbnail_action soup.Root = doc.FindStrict("div", "class", "thumbnail-action-lg")
		if thumbnail_action.Error == nil {
			thumbnail_body = append(thumbnail_body, thumbnail_action.Find("p"))
		}

		var metadata []string
		for _, p := range thumbnail_body {
			metadata = append(metadata, cleanCategory(p.Children()[0].HTML()+" "+p.Text()))
		}

		var downloaded_image []byte
		if download_image == true {
			image := doc.FindAll("img")[2]
			downloaded_image = downloadImage(strings.ReplaceAll(image.Attrs()["src"], "./", ""))
		}

		return &ArticleInfo{
			Metadata: metadata,
			Title:    fmt.Sprintf("<a href=\"%s\">%s</a>", url, title.Text()),
			Img:      downloaded_image,
			Url:      url,
		}
	} else {
		return nil
	}
}

func ProductFound(url string) bool {
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: transCfg}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("User-Agent", "")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Errorf("Error getting product image: %w", err)
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200
}

func FindLastObject() (string, int) {
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: transCfg}
	soup.Header("User-Agent", "")
	resp, err := soup.GetWithClient("https://www.remad.es/web/catalogue", client)
	if err != nil {
		fmt.Errorf("error making the request to Remad catalogue : %w", err)
	}
	doc := soup.HTMLParse(resp)
	results := doc.FindStrict("div", "id", "results")
	href := results.Children()[1].Find("a").Attrs()["href"]
	split_href := strings.Split(href, "/")
	s_id := split_href[len(split_href)-1]
	id, err := strconv.Atoi(s_id)
	if err != nil {
		fmt.Errorf("Casting id to int error: %w", err)
	}
	return href, id
}

func downloadImage(url string) []byte {
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: transCfg}
	soup.Header("User-Agent", "")

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("User-Agent", "")
	resp, err := client.Do(req)

	if err != nil {
		fmt.Errorf("Error getting product image: %w", err)
		return nil
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Errorf("Error reading product image: %w", err)
		return nil
	}
	return body
}

func cleanCategory(x string) string {
	r, _ := regexp.Compile(`[\n\t\r]`)
	return r.ReplaceAllString(x, "")
}
