package binderpos

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUseDecklistForRoll(t *testing.T) {
	tests := []struct {
		name string
		roll int
		want bool
	}{
		{name: "0 routes to decklist", roll: 0, want: true},
		{name: "69 routes to decklist", roll: 69, want: true},
		{name: "70 routes to decklist", roll: 70, want: true},
		{name: "99 routes to decklist", roll: 99, want: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := useDecklistForRoll(test.roll)
			if got != test.want {
				t.Fatalf("expected useDecklistForRoll(%d)=%t, got %t", test.roll, test.want, got)
			}
		})
	}
}

func TestSearchByStorefrontAPIWithClient_UsesProductDetailPathWhenDecklistNotSelected(t *testing.T) {
	server := newStorefrontProductDetailFixtureServer()
	defer server.Close()

	previousSelector := shouldUseDecklistEndpoint
	shouldUseDecklistEndpoint = func() bool { return false }
	t.Cleanup(func() { shouldUseDecklistEndpoint = previousSelector })

	cards, err := searchByStorefrontAPIWithClient(
		context.Background(),
		server.Client(),
		2,
		"Test Store",
		server.URL,
		"Abrade",
	)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(cards) != 1 {
		t.Fatalf("expected 1 card from product details path, got %d", len(cards))
	}
}

func TestSearchByStorefrontAPIWithClient_ReturnsDecklistErrorWhenSelected(t *testing.T) {
	server := newStorefrontProductDetailFixtureServer()
	defer server.Close()

	previousSelector := shouldUseDecklistEndpoint
	shouldUseDecklistEndpoint = func() bool { return true }
	t.Cleanup(func() { shouldUseDecklistEndpoint = previousSelector })

	cards, err := searchByStorefrontAPIWithClient(
		context.Background(),
		server.Client(),
		2,
		"Test Store",
		server.URL,
		"Abrade",
	)
	if err == nil {
		t.Fatalf("expected decklist request error, got nil")
	}
	if len(cards) != 0 {
		t.Fatalf("expected 0 cards when decklist request fails, got %d", len(cards))
	}
}

func newStorefrontProductDetailFixtureServer() *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case storefrontSuggestPath:
			_ = json.NewEncoder(w).Encode(map[string]any{
				"resources": map[string]any{
					"results": map[string]any{
						"products": []map[string]any{
							{
								"title": "Abrade [Foundations]",
								"url":   "/products/abrade",
								"image": "https://images.example/abrade.png",
							},
						},
					},
				},
			})
		case "/products/abrade.js":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"title": "Abrade [Foundations]",
				"variants": []map[string]any{
					{
						"id":        int64(12345),
						"title":     "Near Mint",
						"available": true,
						"price":     199,
					},
				},
			})
		default:
			http.NotFound(w, r)
		}
	})

	return httptest.NewServer(handler)
}
