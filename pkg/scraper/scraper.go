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
	Metadata [5]string
	Title    string
	Img      []byte
}

func ExtractArticleInfo(url string) ArticleInfo {
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: transCfg}
	soup.Header("User-Agent", "")
	resp, err := soup.GetWithClient(url, client)
	if err != nil {
		fmt.Errorf("error making the request to url %s : %w", url, err)
	}
	doc := soup.HTMLParse(resp)
	title := doc.Find("h1")
	image := doc.FindAll("img")[2]
	var thumbnail_body []soup.Root = doc.FindStrict("div", "class", "thumbnail-body").FindAll("p")
	var metadata [5]string
	for i, p := range thumbnail_body {
		metadata[i] = cleanCategory(p.Text())
	}
	return ArticleInfo{
		Metadata: metadata,
		Title:    fmt.Sprintf("<a href=\"%s\">%s</a>", url, title.Text()),
		Img:      downloadImage(strings.ReplaceAll(image.Attrs()["src"], "./", "")),
	}
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
	r, _ := regexp.Compile(`[\n\r\t]`)
	var cleaned = ""
	for _, i := range x {
		cleaned += r.ReplaceAllString(string(i), " ")
	}
	return cleaned
}
