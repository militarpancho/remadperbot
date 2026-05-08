package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"remadperbot/pkg/scraper"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func TestSeedSeenProductsMarksCurrentCatalogHashes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/antiquities/catalog" {
			t.Fatalf("path = %q, want /api/v1/antiquities/catalog", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[
			{"hash":"current-2","name":"Actual 2","state":"Publicado"},
			{"hash":"current-1","name":"Actual 1","state":"Publicado"}
		]`))
	}))
	defer server.Close()
	client := scraper.Client{
		HTTPClient: server.Client(),
		APIBaseURL: server.URL + "/api/v1",
	}
	seen := map[string]bool{}

	if err := seedSeenProducts(seen, client); err != nil {
		t.Fatalf("seedSeenProducts returned error: %v", err)
	}

	if !seen["current-1"] || !seen["current-2"] {
		t.Fatalf("seen = %#v, want both current hashes marked", seen)
	}
}

func TestPostStartupDebugArticlePublishesLatestCatalogArticle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/antiquities/catalog":
			w.Write([]byte(`[
				{"category":{"name":"Menaje","type":"Hogar"},"description":"Reciente","file":{"name":"latest.jpeg"},"hash":"latest","location":"Villaverde","name":"Ultimo","state":"Publicado"},
				{"category":{"name":"Otros","type":"Hogar"},"description":"Anterior","file":{"name":"older.jpeg"},"hash":"older","location":"Moncloa","name":"Anterior","state":"Publicado"}
			]`))
		case "/api/v1/files/download/latest.jpeg":
			w.Write([]byte("image bytes"))
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer server.Close()
	client := scraper.Client{
		HTTPClient:    server.Client(),
		APIBaseURL:    server.URL + "/api/v1",
		DetailBaseURL: "https://remad.example/detalle/",
	}
	poster := &startupArticlePoster{}

	postStartupDebugArticle(poster, client)

	if poster.article == nil {
		t.Fatalf("no article was posted")
	}
	if poster.article.ID != "latest" {
		t.Fatalf("posted article ID = %q, want latest", poster.article.ID)
	}
	if string(poster.article.Img) != "image bytes" {
		t.Fatalf("posted image = %q, want image bytes", string(poster.article.Img))
	}
}

type startupArticlePoster struct {
	article *scraper.ArticleInfo
	err     error
}

func (p *startupArticlePoster) PostNewArticle(articleInfo *scraper.ArticleInfo) (tgbotapi.Message, error) {
	p.article = articleInfo
	if p.err != nil {
		return tgbotapi.Message{}, p.err
	}
	return tgbotapi.Message{}, nil
}

func TestPostStartupDebugArticleContinuesWhenPostingFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/antiquities/catalog":
			w.Write([]byte(`[{"hash":"latest","name":"Ultimo","state":"Publicado"}]`))
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer server.Close()
	client := scraper.Client{
		HTTPClient:    server.Client(),
		APIBaseURL:    server.URL + "/api/v1",
		DetailBaseURL: "https://remad.example/detalle/",
	}
	poster := &startupArticlePoster{err: errors.New("telegram failed")}

	postStartupDebugArticle(poster, client)

	if poster.article == nil {
		t.Fatalf("no article was attempted")
	}
}
