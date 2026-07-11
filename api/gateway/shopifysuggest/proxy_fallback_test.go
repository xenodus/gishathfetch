package shopifysuggest

import (
	"bytes"
	"compress/gzip"
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
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	_, err := fetchProducts(context.Background(), srv.Client(), srv.URL, suggestRequestOpts{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "429")
	require.Equal(t, int32(suggestRetryMaxAttempts), calls.Load())
}

func TestFetchProductsSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sampleSuggestBody))
	}))
	defer srv.Close()

	products, err := fetchProducts(context.Background(), srv.Client(), srv.URL, suggestRequestOpts{})
	require.NoError(t, err)
	require.Len(t, products, 1)
	require.Equal(t, "Opt", products[0].Title)
}

func TestFetchProductsSuccessGzip(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Encoding", "gzip")
		_, _ = w.Write(gzipJSON(t, sampleSuggestBody))
	}))
	defer srv.Close()

	products, err := fetchProducts(context.Background(), srv.Client(), srv.URL, suggestRequestOpts{})
	require.NoError(t, err)
	require.Len(t, products, 1)
	require.Equal(t, "Opt", products[0].Title)
}

// TestRunSearchAttemptsFallsBackOnRateLimit verifies that when one transport
// exhausts its Retry-After/backoff retries, the next transport succeeds.
func TestRunSearchAttemptsFallsBackOnRateLimit(t *testing.T) {
	var hits atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hits.Add(1) <= int32(suggestRetryMaxAttempts) {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = w.Write([]byte(sampleSuggestBody))
	}))
	defer srv.Close()

	attempts := []searchAttempt{
		{strategy: "dedicated", client: srv.Client()},
		{strategy: "direct", client: srv.Client()},
	}

	cards, err := runSearchAttempts(context.Background(), attempts, Options{Config: Config{StoreName: "Test"}, MapProduct: mapAllProducts}, srv.URL)
	require.NoError(t, err)
	require.Len(t, cards, 1)
	require.Equal(t, int32(suggestRetryMaxAttempts+1), hits.Load(), "expected dedicated retries then direct success")
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

	cards, err := runSearchAttempts(context.Background(), attempts, Options{Config: Config{StoreName: "Test"}, MapProduct: mapAllProducts}, srv.URL)
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
		{strategy: "dedicated", client: healthy.Client()},
		{strategy: "direct", client: shouldNotBeHit.Client()},
	}

	cards, err := runSearchAttempts(context.Background(), attempts, Options{Config: Config{StoreName: "Test"}, MapProduct: mapAllProducts}, healthy.URL)
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
	require.Equal(t, "dedicated", attempts[0].strategy)
	require.Equal(t, "direct", attempts[1].strategy)

	t.Setenv("DYNAMIC_PROXY", "http://5.6.7.8:9090")
	attempts = buildSearchAttempts()
	require.Len(t, attempts, 3)
	require.Equal(t, "dedicated", attempts[0].strategy)
	require.Equal(t, "direct", attempts[1].strategy)
	require.Equal(t, "dynamic", attempts[2].strategy)

	t.Setenv("USE_DYNAMIC_PROXY", "false")
	attempts = buildSearchAttempts()
	require.Len(t, attempts, 2)
	require.Equal(t, "dedicated", attempts[0].strategy)
	require.Equal(t, "direct", attempts[1].strategy)
}

func gzipJSON(t *testing.T, body string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, err := zw.Write([]byte(body))
	require.NoError(t, err)
	require.NoError(t, zw.Close())
	return buf.Bytes()
}
