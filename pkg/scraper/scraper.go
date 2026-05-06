package scraper

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	RemadAPIBaseURL    = "https://remad.madrid.es/REMAD_RSP/api/v1"
	RemadDetailBaseURL = "https://remad.madrid.es/REMAD_FTP/#/detalleAntique/"
	anonymousUserHash  = "null"
)

type ArticleInfo struct {
	ID       string
	Metadata []string
	Status   string
	Title    string
	Img      []byte
	Url      string
}

type Client struct {
	HTTPClient    *http.Client
	APIBaseURL    string
	DetailBaseURL string
}

type Antiquity struct {
	Category    Category `json:"category"`
	Description string   `json:"description"`
	File        File     `json:"file"`
	Hash        string   `json:"hash"`
	Location    string   `json:"location"`
	Name        string   `json:"name"`
	Reserved    bool     `json:"reserved"`
	State       string   `json:"state"`
}

type Category struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type File struct {
	Name string `json:"name"`
}

type CatalogFilter struct {
	PLF       *CatalogPLFFilter      `json:"plf"`
	Category  *CatalogCategoryFilter `json:"category"`
	Available bool                   `json:"available"`
	Text      *string                `json:"text"`
	OrderBy   string                 `json:"orderBy"`
	Asc       bool                   `json:"asc"`
}

type CatalogPLFFilter struct {
	ID int `json:"id"`
}

type CatalogCategoryFilter struct {
	Type string `json:"type"`
}

func NewClient() Client {
	return Client{
		HTTPClient: &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}},
	}
}

func DefaultCatalogFilter() CatalogFilter {
	return CatalogFilter{
		OrderBy: "createdAt",
	}
}

func ExtractArticleInfo(articleURL string, downloadImage bool) *ArticleInfo {
	articleInfo, err := NewClient().ExtractArticleInfo(articleURL, downloadImage)
	if err != nil {
		err = fmt.Errorf("error extracting article info from url %s: %w", articleURL, err)
		fmt.Println(err.Error())
		return nil
	}
	return articleInfo
}

func (c Client) ExtractArticleInfo(articleURL string, downloadImage bool) (*ArticleInfo, error) {
	hash := extractHash(articleURL)
	if hash == "" {
		return nil, fmt.Errorf("missing article hash in %q", articleURL)
	}
	antiquity, err := c.GetAntiquity(hash)
	if err != nil {
		return nil, err
	}
	return c.articleInfoFromAntiquity(antiquity, articleURL, downloadImage)
}

func ProductFound(articleURL string) bool {
	_, err := NewClient().GetAntiquity(extractHash(articleURL))
	return err == nil
}

func (c Client) GetAntiquity(hash string) (Antiquity, error) {
	var antiquity Antiquity
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/antiquities/%s?userHash=%s", c.apiBaseURL(), url.PathEscape(hash), url.QueryEscape(anonymousUserHash)), nil)
	if err != nil {
		return antiquity, err
	}
	resp, err := c.httpClient().Do(req)
	if err != nil {
		return antiquity, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return antiquity, fmt.Errorf("unexpected status %d getting antiquity %s", resp.StatusCode, hash)
	}
	if err := json.NewDecoder(resp.Body).Decode(&antiquity); err != nil {
		return antiquity, err
	}
	return antiquity, nil
}

func (c Client) CatalogPage(pageIndex int) ([]Antiquity, error) {
	return c.CatalogPageWithFilter(pageIndex, DefaultCatalogFilter())
}

func (c Client) CatalogPageWithFilter(pageIndex int, filter CatalogFilter) ([]Antiquity, error) {
	var antiques []Antiquity
	body, err := json.Marshal(filter)
	if err != nil {
		return antiques, err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/antiquities/catalog?userHash=%s&pageIndex=%d", c.apiBaseURL(), url.QueryEscape(anonymousUserHash), pageIndex), bytes.NewReader(body))
	if err != nil {
		return antiques, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient().Do(req)
	if err != nil {
		return antiques, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return antiques, fmt.Errorf("unexpected status %d getting catalog page %d", resp.StatusCode, pageIndex)
	}
	if err := json.NewDecoder(resp.Body).Decode(&antiques); err != nil {
		return antiques, err
	}
	return antiques, nil
}

func (c Client) ArticleInfosUntilKnown(knownHashes map[string]bool, downloadImage bool) ([]*ArticleInfo, error) {
	var articles []*ArticleInfo
	for pageIndex := 0; ; pageIndex++ {
		antiques, err := c.CatalogPage(pageIndex)
		if err != nil {
			return nil, err
		}
		if len(antiques) == 0 {
			break
		}
		foundKnownHash := false
		for _, antiquity := range antiques {
			if knownHashes[antiquity.Hash] {
				foundKnownHash = true
				break
			}
			article, err := c.articleInfoFromAntiquity(antiquity, "", downloadImage)
			if err != nil {
				return nil, err
			}
			articles = append(articles, article)
		}
		if foundKnownHash {
			break
		}
	}
	for left, right := 0, len(articles)-1; left < right; left, right = left+1, right-1 {
		articles[left], articles[right] = articles[right], articles[left]
	}
	return articles, nil
}

func (c Client) DownloadFile(name string) ([]byte, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/files/download/%s", c.apiBaseURL(), url.PathEscape(name)), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d downloading file %s", resp.StatusCode, name)
	}
	return io.ReadAll(resp.Body)
}

func (c Client) articleInfoFromAntiquity(antiquity Antiquity, articleURL string, downloadImage bool) (*ArticleInfo, error) {
	if articleURL == "" {
		articleURL = c.detailBaseURL() + antiquity.Hash
	}
	var image []byte
	if downloadImage && antiquity.File.Name != "" {
		downloadedImage, err := c.DownloadFile(antiquity.File.Name)
		if err != nil {
			return nil, err
		}
		image = downloadedImage
	}
	return &ArticleInfo{
		Metadata: []string{
			fmt.Sprintf("Categoría: %s / %s", antiquity.Category.Type, antiquity.Category.Name),
			fmt.Sprintf("Punto limpio: %s", antiquity.Location),
			fmt.Sprintf("Descripción: %s", antiquity.Description),
			fmt.Sprintf("Estado: %s", antiquity.State),
		},
		ID:     antiquity.Hash,
		Status: antiquity.State,
		Title:  fmt.Sprintf("<a href=\"%s\">%s</a>", articleURL, antiquity.Name),
		Img:    image,
		Url:    articleURL,
	}, nil
}

func (c Client) httpClient() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return NewClient().HTTPClient
}

func (c Client) apiBaseURL() string {
	if c.APIBaseURL == "" {
		return RemadAPIBaseURL
	}
	return strings.TrimRight(c.APIBaseURL, "/")
}

func (c Client) detailBaseURL() string {
	if c.DetailBaseURL == "" {
		return RemadDetailBaseURL
	}
	return c.DetailBaseURL
}

func extractHash(articleURL string) string {
	trimmedURL := strings.TrimRight(articleURL, "/")
	parts := strings.Split(trimmedURL, "/")
	return parts[len(parts)-1]
}
