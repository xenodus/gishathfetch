package shopifysuggest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"mtg-price-checker-sg/gateway"
)

func mapAllProducts(cfg Config, product Product) (gateway.Card, bool) {
	return gateway.Card{Name: product.Title, Source: cfg.StoreName, InStock: true}, true
}

const sampleSuggestBody = `{
  "resources": {
    "results": {
      "products": [
        {"title": "Opt", "available": true, "type": "MTG Single Cards", "price": "0.50"}
      ]
    }
  }
}`

func TestFetchProductsRateLimited(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	_, err := fetchProducts(context.Background(), srv.Client(), srv.URL)
	require.Error(t, err)
	require.Contains(t, err.Error(), "429")
}

func TestFetchProductsSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sampleSuggestBody))
	}))
	defer srv.Close()

	products, err := fetchProducts(context.Background(), srv.Client(), srv.URL)
	require.NoError(t, err)
	require.Len(t, products, 1)
	require.Equal(t, "Opt", products[0].Title)
}

// TestRunSearchAttemptsFallsBackOnRateLimit verifies that a rate-limited first
// attempt falls through to the next transport, which succeeds. All attempts hit
// the same endpoint (as they do in production), so the server returns 429 on
// the first request and a valid body afterwards.
func TestRunSearchAttemptsFallsBackOnRateLimit(t *testing.T) {
	var hits atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hits.Add(1) == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = w.Write([]byte(sampleSuggestBody))
	}))
	defer srv.Close()

	attempts := []searchAttempt{
		{strategy: "direct", client: srv.Client()},
		{strategy: "dedicated", client: srv.Client()},
	}

	cards, err := runSearchAttempts(context.Background(), attempts, Config{StoreName: "Test"}, srv.URL, mapAllProducts)
	require.NoError(t, err)
	require.Len(t, cards, 1)
	require.Equal(t, int32(2), hits.Load(), "expected fallback to retry after the first 429")
}

// TestRunSearchAttemptsReturnsLastError verifies that when every transport
// fails, the final (last attempt) error is surfaced.
func TestRunSearchAttemptsReturnsLastError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	attempts := []searchAttempt{
		{strategy: "direct", client: srv.Client()},
		{strategy: "dedicated", client: srv.Client()},
		{strategy: "dynamic", client: srv.Client()},
	}

	cards, err := runSearchAttempts(context.Background(), attempts, Config{StoreName: "Test"}, srv.URL, mapAllProducts)
	require.Error(t, err)
	require.Empty(t, cards)
	require.Contains(t, err.Error(), "attempt 3 (dynamic)")
}

// TestRunSearchAttemptsStopsAtFirstSuccess verifies that a successful attempt
// short-circuits the chain so later transports are never used.
func TestRunSearchAttemptsStopsAtFirstSuccess(t *testing.T) {
	var secondHits atomic.Int32

	healthy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(sampleSuggestBody))
	}))
	defer healthy.Close()

	shouldNotBeHit := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secondHits.Add(1)
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer shouldNotBeHit.Close()

	attempts := []searchAttempt{
		{strategy: "direct", client: healthy.Client()},
		{strategy: "dedicated", client: shouldNotBeHit.Client()},
	}

	cards, err := runSearchAttempts(context.Background(), attempts, Config{StoreName: "Test"}, healthy.URL, mapAllProducts)
	require.NoError(t, err)
	require.Len(t, cards, 1)
	require.Zero(t, secondHits.Load(), "second transport should not be tried after first success")
}

func TestBuildSearchAttemptsProxyConfiguration(t *testing.T) {
	t.Setenv("DEDICATED_PROXY_1", "")
	t.Setenv("DEDICATED_PROXY_2", "")
	t.Setenv("DEDICATED_PROXY_3", "")
	t.Setenv("DEDICATED_PROXY_4", "")
	t.Setenv("DEDICATED_PROXY_5", "")
	t.Setenv("DEDICATED_PROXY_6", "")
	t.Setenv("DEDICATED_PROXY_7", "")
	t.Setenv("DYNAMIC_PROXY", "")

	attempts := buildSearchAttempts()
	require.Len(t, attempts, 1)
	require.Equal(t, "direct", attempts[0].strategy)

	t.Setenv("DEDICATED_PROXY_1", "1.2.3.4|8080|user|pass")
	attempts = buildSearchAttempts()
	require.Len(t, attempts, 2)
	require.Equal(t, "direct", attempts[0].strategy)
	require.Equal(t, "dedicated", attempts[1].strategy)

	t.Setenv("DYNAMIC_PROXY", "http://5.6.7.8:9090")
	attempts = buildSearchAttempts()
	require.Len(t, attempts, 3)
	require.Equal(t, "dynamic", attempts[2].strategy)
}
