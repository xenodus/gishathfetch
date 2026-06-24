package shopifysuggest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestIsRetriableSuggestStatus(t *testing.T) {
	retriable := []int{
		http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout,
	}
	for _, status := range retriable {
		if !isRetriableSuggestStatus(status) {
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
		if isRetriableSuggestStatus(status) {
			t.Fatalf("expected status %d to not be retriable", status)
		}
	}
}

func TestParseSuggestRetryAfter(t *testing.T) {
	t.Run("empty returns zero", func(t *testing.T) {
		if d := parseSuggestRetryAfter(""); d != 0 {
			t.Fatalf("expected 0, got %s", d)
		}
	})

	t.Run("parses seconds", func(t *testing.T) {
		if d := parseSuggestRetryAfter("2"); d != 2*time.Second {
			t.Fatalf("expected 2s, got %s", d)
		}
	})

	t.Run("non-positive seconds returns zero", func(t *testing.T) {
		if d := parseSuggestRetryAfter("0"); d != 0 {
			t.Fatalf("expected 0, got %s", d)
		}
	})

	t.Run("caps large seconds at max", func(t *testing.T) {
		if d := parseSuggestRetryAfter("3600"); d != suggestRetryMaxDelay {
			t.Fatalf("expected cap %s, got %s", suggestRetryMaxDelay, d)
		}
	})

	t.Run("parses http date in the future", func(t *testing.T) {
		when := time.Now().Add(1 * time.Second).UTC().Format(http.TimeFormat)
		d := parseSuggestRetryAfter(when)
		if d <= 0 || d > suggestRetryMaxDelay {
			t.Fatalf("expected positive capped delay, got %s", d)
		}
	})

	t.Run("past http date returns zero", func(t *testing.T) {
		when := time.Now().Add(-1 * time.Hour).UTC().Format(http.TimeFormat)
		if d := parseSuggestRetryAfter(when); d != 0 {
			t.Fatalf("expected 0 for past date, got %s", d)
		}
	})

	t.Run("garbage returns zero", func(t *testing.T) {
		if d := parseSuggestRetryAfter("soon"); d != 0 {
			t.Fatalf("expected 0 for unparseable value, got %s", d)
		}
	})
}

func TestDoSuggestGETWithRetry_SucceedsAfter429WithRetryAfter(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(sampleSuggestBody))
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), suggestAttemptTimeout)
	defer cancel()

	body, err := doSuggestGETWithRetry(ctx, server.Client(), server.URL)
	if err != nil {
		t.Fatalf("expected success after Retry-After retry, got %v", err)
	}
	if !strings.Contains(string(body), "Opt") {
		t.Fatalf("expected suggest body, got %q", string(body))
	}
	if got := calls.Load(); got != 2 {
		t.Fatalf("expected 2 calls (429 + success), got %d", got)
	}
}

func TestDoSuggestGETWithRetry_SucceedsAfterTransient503(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) <= 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(sampleSuggestBody))
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), suggestAttemptTimeout)
	defer cancel()

	_, err := doSuggestGETWithRetry(ctx, server.Client(), server.URL)
	if err != nil {
		t.Fatalf("expected success after retries, got %v", err)
	}
	if got := calls.Load(); got != 3 {
		t.Fatalf("expected 3 calls (2 failures + success), got %d", got)
	}
}

func TestDoSuggestGETWithRetry_ExhaustsRetriesOnPersistent429(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), suggestAttemptTimeout)
	defer cancel()

	_, err := doSuggestGETWithRetry(ctx, server.Client(), server.URL)
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	if !strings.Contains(err.Error(), "429") {
		t.Fatalf("expected status=429 in error, got %v", err)
	}
	if got := calls.Load(); got != int32(suggestRetryMaxAttempts) {
		t.Fatalf("expected %d calls, got %d", suggestRetryMaxAttempts, got)
	}
}

func TestDoSuggestGETWithRetry_DoesNotRetryNonRetriableStatus(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), suggestAttemptTimeout)
	defer cancel()

	_, err := doSuggestGETWithRetry(ctx, server.Client(), server.URL)
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("expected exactly 1 call for non-retriable status, got %d", got)
	}
}
