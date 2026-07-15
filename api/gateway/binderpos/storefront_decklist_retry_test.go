package binderpos

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func newDecklistTestRequest(t *testing.T, urlStr string) func() (*http.Request, error) {
	t.Helper()
	return func() (*http.Request, error) {
		return http.NewRequest(http.MethodPost, urlStr, bytes.NewReader([]byte(`[{"card":"Abrade","quantity":1}]`)))
	}
}

func TestDoDecklistRequestWithRetry_SucceedsOnOK(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	res, err := doDecklistRequestWithRetry(context.Background(), server.Client(), newDecklistTestRequest(t, server.URL))
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
}

func TestDoDecklistRequestWithRetry_ReturnsErrorOnNonOK(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("rate limited"))
	}))
	defer server.Close()

	_, err := doDecklistRequestWithRetry(context.Background(), server.Client(), newDecklistTestRequest(t, server.URL))
	if err == nil {
		t.Fatal("expected error for 503 response")
	}
	if !strings.Contains(err.Error(), "status=503") {
		t.Fatalf("expected status=503 in error, got %v", err)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("expected exactly 1 call, got %d", got)
	}
}
