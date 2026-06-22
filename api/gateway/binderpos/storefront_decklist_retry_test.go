package binderpos

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func newDecklistTestRequest(t *testing.T, urlStr string) func() (*http.Request, error) {
	t.Helper()
	return func() (*http.Request, error) {
		return http.NewRequest(http.MethodPost, urlStr, bytes.NewReader([]byte(`[{"card":"Abrade","quantity":1}]`)))
	}
}

func TestIsRetriableDecklistStatus(t *testing.T) {
	retriable := []int{
		http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout,
	}
	for _, status := range retriable {
		if !isRetriableDecklistStatus(status) {
			t.Fatalf("expected status %d to be retriable", status)
		}
	}

	notRetriable := []int{
		http.StatusOK,
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusNotFound,
	}
	for _, status := range notRetriable {
		if isRetriableDecklistStatus(status) {
			t.Fatalf("expected status %d to not be retriable", status)
		}
	}
}

func TestParseRetryAfter(t *testing.T) {
	t.Run("empty returns zero", func(t *testing.T) {
		if d := parseRetryAfter(""); d != 0 {
			t.Fatalf("expected 0, got %s", d)
		}
	})

	t.Run("parses seconds", func(t *testing.T) {
		if d := parseRetryAfter("2"); d != 2*time.Second {
			t.Fatalf("expected 2s, got %s", d)
		}
	})

	t.Run("non-positive seconds returns zero", func(t *testing.T) {
		if d := parseRetryAfter("0"); d != 0 {
			t.Fatalf("expected 0, got %s", d)
		}
	})

	t.Run("caps large seconds at max", func(t *testing.T) {
		if d := parseRetryAfter("3600"); d != binderposDecklistRetryMaxDelay {
			t.Fatalf("expected cap %s, got %s", binderposDecklistRetryMaxDelay, d)
		}
	})

	t.Run("parses http date in the future", func(t *testing.T) {
		when := time.Now().Add(1 * time.Second).UTC().Format(http.TimeFormat)
		d := parseRetryAfter(when)
		if d <= 0 || d > binderposDecklistRetryMaxDelay {
			t.Fatalf("expected positive capped delay, got %s", d)
		}
	})

	t.Run("past http date returns zero", func(t *testing.T) {
		when := time.Now().Add(-1 * time.Hour).UTC().Format(http.TimeFormat)
		if d := parseRetryAfter(when); d != 0 {
			t.Fatalf("expected 0 for past date, got %s", d)
		}
	})

	t.Run("garbage returns zero", func(t *testing.T) {
		if d := parseRetryAfter("soon"); d != 0 {
			t.Fatalf("expected 0 for unparseable value, got %s", d)
		}
	})
}

func TestDecklistBackoffDelayIsBoundedAndGrows(t *testing.T) {
	for attempt := 0; attempt < 6; attempt++ {
		d := decklistBackoffDelay(attempt)
		if d <= 0 {
			t.Fatalf("attempt %d: expected positive delay, got %s", attempt, d)
		}
		if d > binderposDecklistRetryMaxDelay {
			t.Fatalf("attempt %d: expected delay <= %s, got %s", attempt, binderposDecklistRetryMaxDelay, d)
		}
	}

	// Equal jitter guarantees a minimum that grows with the attempt until capped.
	if min0 := binderposDecklistRetryBaseDelay / 2; decklistBackoffDelay(0) < min0 {
		t.Fatalf("attempt 0 below equal-jitter floor %s", min0)
	}
}

func TestDoDecklistRequestWithRetry_SucceedsAfterTransient503(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) <= 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), binderposAttemptTimeout)
	defer cancel()

	res, err := doDecklistRequestWithRetry(ctx, server.Client(), newDecklistTestRequest(t, server.URL))
	if err != nil {
		t.Fatalf("expected success after retries, got %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	if got := calls.Load(); got != 3 {
		t.Fatalf("expected 3 calls (2 failures + success), got %d", got)
	}
}

func TestDoDecklistRequestWithRetry_ExhaustsRetriesOnPersistent503(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("rate limited"))
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), binderposAttemptTimeout)
	defer cancel()

	res, err := doDecklistRequestWithRetry(ctx, server.Client(), newDecklistTestRequest(t, server.URL))
	if err == nil {
		if res != nil {
			res.Body.Close()
		}
		t.Fatal("expected error after exhausting retries")
	}
	if !strings.Contains(err.Error(), "status=503") {
		t.Fatalf("expected status=503 in error, got %v", err)
	}
	if got := calls.Load(); got != int32(binderposDecklistMaxAttempts) {
		t.Fatalf("expected %d calls, got %d", binderposDecklistMaxAttempts, got)
	}
}

func TestDoDecklistRequestWithRetry_DoesNotRetryNonRetriableStatus(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), binderposAttemptTimeout)
	defer cancel()

	_, err := doDecklistRequestWithRetry(ctx, server.Client(), newDecklistTestRequest(t, server.URL))
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("expected exactly 1 call for non-retriable status, got %d", got)
	}
}

func TestDoDecklistRequestWithRetry_StopsWhenContextCancelled(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	// Budget smaller than the first backoff so the retry wait is refused.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := doDecklistRequestWithRetry(ctx, server.Client(), newDecklistTestRequest(t, server.URL))
	if err == nil {
		t.Fatal("expected error when context budget is exhausted")
	}
	// A budget smaller than the first backoff must prevent the retry loop from
	// running the full attempt count. The exact count (0 or 1) depends on shared
	// per-host pacing, so only assert it stopped well short of exhaustion.
	if got := calls.Load(); got >= int32(binderposDecklistMaxAttempts) {
		t.Fatalf("expected fewer than %d calls on tight budget, got %d", binderposDecklistMaxAttempts, got)
	}
}
