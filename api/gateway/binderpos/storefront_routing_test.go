package binderpos

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUseDecklistForRoute(t *testing.T) {
	tests := []struct {
		name string
		seq  uint32
		want bool
	}{
		{name: "even seq 0 routes to decklist", seq: 0, want: true},
		{name: "odd seq 1 routes to product details", seq: 1, want: false},
		{name: "even seq 2 routes to decklist", seq: 2, want: true},
		{name: "odd seq 3 routes to product details", seq: 3, want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := useDecklistForRoute(test.seq)
			if got != test.want {
				t.Fatalf("expected useDecklistForRoute(%d)=%t, got %t", test.seq, test.want, got)
			}
		})
	}
}

// TestShouldUseDecklistEndpoint_SplitsConsecutiveCallsEvenly verifies that the
// round-robin selector hands half of the first attempts to the decklist portal
// and half to the per-store product-details path, which is what reduces the
// concurrent load on the shared portal host.
func TestShouldUseDecklistEndpoint_SplitsConsecutiveCallsEvenly(t *testing.T) {
	previousSeq := binderposDecklistRouteSeq.Load()
	binderposDecklistRouteSeq.Store(0)
	t.Cleanup(func() { binderposDecklistRouteSeq.Store(previousSeq) })

	const calls = 10
	decklistCount := 0
	for i := 0; i < calls; i++ {
		if shouldUseDecklistEndpoint() {
			decklistCount++
		}
	}

	if decklistCount != calls/2 {
		t.Fatalf("expected %d of %d first attempts routed to decklist, got %d", calls/2, calls, decklistCount)
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
		"",
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
		"",
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
