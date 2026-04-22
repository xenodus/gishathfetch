package binderpos

import (
	"context"
	"errors"
	"fmt"
	"net/http/httptest"
	"net/http"
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

func TestFetchProductDetail_AcceptsAny2xxStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/products/test-card.js" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(server.Close)

	detail, err := fetchProductDetail(context.Background(), server.Client(), server.URL, "/products/test-card")
	if err != nil {
		t.Fatalf("expected nil error for 2xx response, got %v", err)
	}
	if detail == nil {
		t.Fatalf("expected detail struct for 2xx response")
	}
}

func TestFetchProductDetail_ReturnsErrorFor4xxAnd5xx(t *testing.T) {
	tests := []int{http.StatusNotFound, http.StatusServiceUnavailable}

	for _, status := range tests {
		t.Run(fmt.Sprintf("status-%d", status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/products/test-card.js" {
					t.Fatalf("unexpected path: %s", r.URL.Path)
				}
				w.WriteHeader(status)
				_, _ = w.Write([]byte("detail endpoint failed"))
			}))
			t.Cleanup(server.Close)

			_, err := fetchProductDetail(context.Background(), server.Client(), server.URL, "/products/test-card")
			if err == nil {
				t.Fatalf("expected error for %d response", status)
			}
			var statusErr *httpStatusError
			if !errors.As(err, &statusErr) {
				t.Fatalf("expected httpStatusError, got %T (%v)", err, err)
			}
			if statusErr.status != status {
				t.Fatalf("expected status %d, got %d", status, statusErr.status)
			}
		})
	}
}

func TestSearchByStorefrontAPIWithClient_PropagatesDetail4xx5xxErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/search/suggest.json":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"resources":{"results":{"products":[{"title":"Test Card","url":"/products/test-card","image":"//example.com/image.png"}]}}}`))
		case "/products/test-card.js":
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("temporarily unavailable"))
		default:
			t.Fatalf("unexpected request path: %s", r.URL.Path)
		}
	}))
	t.Cleanup(server.Close)

	cards, err := searchByStorefrontAPIWithClient(context.Background(), server.Client(), 2, "Test Shop", server.URL, "test card")
	if err == nil {
		t.Fatalf("expected detail 5xx error to be propagated")
	}
	if len(cards) != 0 {
		t.Fatalf("expected no cards when detail endpoint fails, got %+v", cards)
	}
	if !strings.Contains(err.Error(), "status=503") {
		t.Fatalf("expected status code in error, got %v", err)
	}
}
