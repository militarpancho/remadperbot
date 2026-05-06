package scraper

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestExtractArticleInfoUsesRemadDetailAPI(t *testing.T) {
	var requestedPath string
	var requestedUserHash string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.Path
		requestedUserHash = r.URL.Query().Get("userHash")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"category":{"name":"Menaje","type":"Hogar"},
			"description":"Tazas de cafe",
			"file":{"name":"abc123.jpeg"},
			"hash":"abc123",
			"location":"Villaverde",
			"name":"4 tazas de cafe",
			"state":"Publicado"
		}`))
	}))
	defer server.Close()

	client := Client{
		HTTPClient:    server.Client(),
		APIBaseURL:    server.URL + "/api/v1",
		DetailBaseURL: "https://remad.example/detalle/",
	}

	article, err := client.ExtractArticleInfo("https://remad.example/detalle/abc123", false)

	if err != nil {
		t.Fatalf("ExtractArticleInfo returned error: %v", err)
	}
	if requestedPath != "/api/v1/antiquities/abc123" {
		t.Fatalf("requested path = %q, want /api/v1/antiquities/abc123", requestedPath)
	}
	if requestedUserHash != "null" {
		t.Fatalf("requested userHash = %q, want null", requestedUserHash)
	}
	if article.Title != `<a href="https://remad.example/detalle/abc123">4 tazas de cafe</a>` {
		t.Fatalf("title = %q", article.Title)
	}
	if article.ID != "abc123" {
		t.Fatalf("id = %q, want abc123", article.ID)
	}
	if article.Status != "Publicado" {
		t.Fatalf("status = %q, want Publicado", article.Status)
	}
	wantMetadata := []string{
		"Categoría: Hogar / Menaje",
		"Punto limpio: Villaverde",
		"Descripción: Tazas de cafe",
		"Estado: Publicado",
	}
	if !reflect.DeepEqual(article.Metadata, wantMetadata) {
		t.Fatalf("metadata = %#v, want %#v", article.Metadata, wantMetadata)
	}
	if article.Img != nil {
		t.Fatalf("image was downloaded when downloadImage=false")
	}
}

func TestCatalogPagePostsPublicCreatedAtFilter(t *testing.T) {
	var requestedMethod string
	var requestedPath string
	var requestedUserHash string
	var requestedPageIndex string
	var requestedFilter map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedMethod = r.Method
		requestedPath = r.URL.Path
		requestedUserHash = r.URL.Query().Get("userHash")
		requestedPageIndex = r.URL.Query().Get("pageIndex")
		if err := json.NewDecoder(r.Body).Decode(&requestedFilter); err != nil {
			t.Fatalf("could not decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{
			"category":{"name":"Juguetes","type":"Niños y bebes"},
			"file":{"name":"thumbnail_hash-1.jpeg"},
			"hash":"hash-1",
			"location":"Villaverde",
			"name":"Maxicosi con muñecas",
			"state":"Publicado"
		}]`))
	}))
	defer server.Close()
	client := Client{
		HTTPClient: server.Client(),
		APIBaseURL: server.URL + "/api/v1",
	}

	antiques, err := client.CatalogPage(0)

	if err != nil {
		t.Fatalf("CatalogPage returned error: %v", err)
	}
	if requestedMethod != http.MethodPost {
		t.Fatalf("requested method = %q, want POST", requestedMethod)
	}
	if requestedPath != "/api/v1/antiquities/catalog" {
		t.Fatalf("requested path = %q, want /api/v1/antiquities/catalog", requestedPath)
	}
	if requestedUserHash != "null" {
		t.Fatalf("requested userHash = %q, want null", requestedUserHash)
	}
	if requestedPageIndex != "0" {
		t.Fatalf("requested pageIndex = %q, want 0", requestedPageIndex)
	}
	wantFilter := map[string]interface{}{
		"plf":       nil,
		"category":  nil,
		"available": false,
		"text":      nil,
		"orderBy":   "createdAt",
		"asc":       false,
	}
	if !reflect.DeepEqual(requestedFilter, wantFilter) {
		t.Fatalf("filter = %#v, want %#v", requestedFilter, wantFilter)
	}
	if len(antiques) != 1 {
		t.Fatalf("len(antiques) = %d, want 1", len(antiques))
	}
	if antiques[0].Hash != "hash-1" {
		t.Fatalf("hash = %q, want hash-1", antiques[0].Hash)
	}
}

func TestArticleInfosUntilKnownStopsAtFirstKnownHash(t *testing.T) {
	var requestedPages []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPages = append(requestedPages, r.URL.Query().Get("pageIndex"))
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("pageIndex") {
		case "0":
			w.Write([]byte(`[
				{"category":{"name":"Menaje","type":"Hogar"},"file":{"name":"new-2.jpeg"},"hash":"new-2","location":"Villaverde","name":"Nuevo 2","state":"Publicado"},
				{"category":{"name":"Juguetes","type":"Niños y bebes"},"file":{"name":"new-1.jpeg"},"hash":"new-1","location":"Barajas","name":"Nuevo 1","state":"Publicado"}
			]`))
		case "1":
			w.Write([]byte(`[
				{"category":{"name":"Otros","type":"Hogar"},"file":{"name":"known.jpeg"},"hash":"known","location":"Moncloa","name":"Conocido","state":"Publicado"},
				{"category":{"name":"Otros","type":"Hogar"},"file":{"name":"older.jpeg"},"hash":"older","location":"Moncloa","name":"Antiguo","state":"Publicado"}
			]`))
		default:
			t.Fatalf("unexpected pageIndex %q", r.URL.Query().Get("pageIndex"))
		}
	}))
	defer server.Close()
	client := Client{
		HTTPClient:    server.Client(),
		APIBaseURL:    server.URL + "/api/v1",
		DetailBaseURL: "https://remad.example/detalle/",
	}

	articles, err := client.ArticleInfosUntilKnown(map[string]bool{"known": true}, false)

	if err != nil {
		t.Fatalf("ArticleInfosUntilKnown returned error: %v", err)
	}
	wantPages := []string{"0", "1"}
	if !reflect.DeepEqual(requestedPages, wantPages) {
		t.Fatalf("requested pages = %#v, want %#v", requestedPages, wantPages)
	}
	if len(articles) != 2 {
		t.Fatalf("len(articles) = %d, want 2", len(articles))
	}
	if articles[0].Url != "https://remad.example/detalle/new-1" {
		t.Fatalf("first article URL = %q, want new-1 URL", articles[0].Url)
	}
	if articles[1].Url != "https://remad.example/detalle/new-2" {
		t.Fatalf("second article URL = %q, want new-2 URL", articles[1].Url)
	}
}
