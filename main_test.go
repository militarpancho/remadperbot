package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"remadperbot/pkg/scraper"
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
