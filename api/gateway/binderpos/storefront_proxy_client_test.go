package binderpos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"mtg-price-checker-sg/gateway"
)

func TestNewHTTPClientWithProxyURL(t *testing.T) {
	t.Run("returns error for invalid proxy URL", func(t *testing.T) {
		_, err := newHTTPClientWithProxyURL("://invalid-proxy")
		if err == nil {
			t.Fatalf("expected invalid proxy URL to return error")
		}
	})

	t.Run("builds client with configured proxy and timeout", func(t *testing.T) {
		client, err := newHTTPClientWithProxyURL("http://user:pass@10.0.0.1:8080")
		if err != nil {
			t.Fatalf("expected valid proxy URL, got error %v", err)
		}
		if client.Timeout != binderposAttemptTimeout {
			t.Fatalf("expected timeout %s, got %s", binderposAttemptTimeout, client.Timeout)
		}

		transport, ok := client.Transport.(*http.Transport)
		if !ok {
			t.Fatalf("expected http.Transport, got %T", client.Transport)
		}
		if transport.Proxy == nil {
			t.Fatalf("expected proxy function to be configured")
		}

		reqURL, err := url.Parse("https://example.com")
		if err != nil {
			t.Fatalf("failed to parse request URL: %v", err)
		}
		proxyURL, err := transport.Proxy(&http.Request{URL: reqURL})
		if err != nil {
			t.Fatalf("expected proxy function to succeed, got %v", err)
		}
		if proxyURL == nil || proxyURL.String() != "http://user:pass@10.0.0.1:8080" {
			t.Fatalf("unexpected proxy URL: %v", proxyURL)
		}
	})
}

func TestRunWithAttemptTimeout(t *testing.T) {
	t.Run("returns callback result when callback succeeds", func(t *testing.T) {
		got, err := runWithAttemptTimeout(context.Background(), func(attemptCtx context.Context) ([]gateway.Card, error) {
			if _, hasDeadline := attemptCtx.Deadline(); !hasDeadline {
				t.Fatalf("expected attempt context to have deadline")
			}
			return []gateway.Card{{Name: "ok"}}, nil
		})
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if len(got) != 1 || got[0].Name != "ok" {
			t.Fatalf("expected ok result, got %+v", got)
		}
	})

	t.Run("propagates callback error", func(t *testing.T) {
		wantErr := errors.New("boom")
		_, err := runWithAttemptTimeout(context.Background(), func(_ context.Context) ([]gateway.Card, error) {
			return nil, wantErr
		})
		if !errors.Is(err, wantErr) {
			t.Fatalf("expected %v, got %v", wantErr, err)
		}
	})
}

func TestSearchByStorefrontAPIWithClient_DetailRequestFailures(t *testing.T) {
	t.Run("returns error when all detail requests fail", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case storefrontSuggestPath:
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(storefrontSuggestResponse{
					Resources: struct {
						Results struct {
							Products []storefrontProduct `json:"products"`
						} `json:"results"`
					}{
						Results: struct {
							Products []storefrontProduct `json:"products"`
						}{
							Products: []storefrontProduct{
								{Title: "Card A", URL: "/products/card-a", Image: "//img/a.jpg"},
								{Title: "Card B", URL: "/products/card-b", Image: "//img/b.jpg"},
							},
						},
					},
				})
			case "/products/card-a.js", "/products/card-b.js":
				http.Error(w, "upstream unavailable", http.StatusServiceUnavailable)
			default:
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		cards, err := searchByStorefrontAPIWithClient(context.Background(), server.Client(), 2, "Test Store", server.URL, "Abrade")
		if err == nil {
			t.Fatal("expected error when all product detail requests fail")
		}
		if len(cards) != 0 {
			t.Fatalf("expected no cards when all product detail requests fail, got %+v", cards)
		}
		if !strings.Contains(err.Error(), "storefront detail request failed for all 2 candidate products") {
			t.Fatalf("expected aggregated detail failure error, got %v", err)
		}
	})

	t.Run("returns cards when at least one detail request succeeds", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case storefrontSuggestPath:
				w.Header().Set("Content-Type", "application/json")
				_, _ = fmt.Fprint(w, `{"resources":{"results":{"products":[{"title":"Card A","url":"/products/card-a","image":"//img/a.jpg"},{"title":"Card B","url":"/products/card-b","image":"//img/b.jpg"}]}}}`)
			case "/products/card-a.js":
				http.Error(w, "upstream unavailable", http.StatusServiceUnavailable)
			case "/products/card-b.js":
				w.Header().Set("Content-Type", "application/json")
				_, _ = fmt.Fprint(w, `{"title":"Card B","variants":[{"id":12345,"title":"NM","available":true,"price":250}]}`)
			default:
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		cards, err := searchByStorefrontAPIWithClient(context.Background(), server.Client(), 2, "Test Store", server.URL, "Abrade")
		if err != nil {
			t.Fatalf("expected nil error when at least one detail request succeeds, got %v", err)
		}
		if len(cards) != 1 {
			t.Fatalf("expected one card from successful detail response, got %+v", cards)
		}
	})
}
