package steam

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStoreHeroBuildsTheHashedAssetURL(t *testing.T) {
	var gotInput string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotInput = r.URL.Query().Get("input_json")
		_, _ = w.Write([]byte(`{"response":{"store_items":[{"assets":{"asset_url_format":"steam/apps/42/${FILENAME}?t=1","header":"h.jpg"}}]}}`))
	}))
	defer srv.Close()
	if got, want := storeHero(t.Context(), nil, srv.URL, "https://cdn/", "42"), "https://cdn/steam/apps/42/h.jpg?t=1"; got != want {
		t.Fatalf("header fallback = %q, want %q", got, want)
	}
	if !strings.Contains(gotInput, `"appid":42`) {
		t.Fatalf("request input = %q", gotInput)
	}
}

func TestStoreHeroIsEmptyWhenTheStoreNamesNone(t *testing.T) {
	for name, body := range map[string]string{
		"no items":  `{"response":{"store_items":[]}}`,
		"no format": `{"response":{"store_items":[{"assets":{"library_hero":"x.jpg"}}]}}`,
		"no files":  `{"response":{"store_items":[{"assets":{"asset_url_format":"a/${FILENAME}"}}]}}`,
		"not json":  `<html>`,
	} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(body)) }))
		if got := storeHero(t.Context(), nil, srv.URL, "https://cdn/", "42"); got != "" {
			t.Errorf("%s: %q", name, got)
		}
		srv.Close()
	}
	down := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { http.Error(w, "x", http.StatusInternalServerError) }))
	defer down.Close()
	if got := storeHero(t.Context(), nil, down.URL, "https://cdn/", "42"); got != "" {
		t.Errorf("500: %q", got)
	}
	if got := StoreHero(t.Context(), nil, "../x"); got != "" {
		t.Errorf("non-numeric id: %q", got)
	}
}
