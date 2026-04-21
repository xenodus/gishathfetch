package binderpos

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
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

func TestDoStorefrontGETWithRetry_RetriesTransientStatusThenSucceeds(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		curr := attempts.Add(1)
		if curr < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = io.WriteString(w, "temporary unavailable")
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"ok":true}`)
	}))
	defer server.Close()

	client := &http.Client{Timeout: binderposAttemptTimeout}
	profile := newStorefrontRequestProfile(server.URL)

	resp, err := doStorefrontGETWithRetry(context.Background(), client, server.URL, profile)
	if err != nil {
		t.Fatalf("expected retry to eventually succeed, got error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 after retries, got %d", resp.StatusCode)
	}
	if got := attempts.Load(); got != 3 {
		t.Fatalf("expected 3 attempts, got %d", got)
	}
}

func TestDoStorefrontGETWithRetry_NoRetryForBadRequest(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, "bad request")
	}))
	defer server.Close()

	client := &http.Client{Timeout: binderposAttemptTimeout}
	profile := newStorefrontRequestProfile(server.URL)

	resp, err := doStorefrontGETWithRetry(context.Background(), client, server.URL, profile)
	if err != nil {
		t.Fatalf("expected non-retryable status to return response without transport error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
	if got := attempts.Load(); got != 1 {
		t.Fatalf("expected exactly one attempt for 400 status, got %d", got)
	}
}

func TestApplyStorefrontHeaders_UsesBrowserLikeDefaults(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "https://example.com/search/suggest.json?q=test", nil)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}
	profile := storefrontRequestProfile{
		userAgent:      "Mozilla/5.0 test-agent",
		acceptLanguage: "en-SG,en;q=0.9",
		referer:        "https://example.com/",
	}

	applyStorefrontHeaders(req, profile)

	if req.Header.Get("User-Agent") != profile.userAgent {
		t.Fatalf("unexpected User-Agent header: %s", req.Header.Get("User-Agent"))
	}
	if req.Header.Get("Accept-Language") != profile.acceptLanguage {
		t.Fatalf("unexpected Accept-Language header: %s", req.Header.Get("Accept-Language"))
	}
	if req.Header.Get("Referer") != profile.referer {
		t.Fatalf("unexpected Referer header: %s", req.Header.Get("Referer"))
	}
	if req.Header.Get("X-Requested-With") != "XMLHttpRequest" {
		t.Fatalf("expected X-Requested-With header to mimic xhr requests")
	}
	if req.Header.Get("Origin") != "https://example.com" {
		t.Fatalf("unexpected Origin header: %s", req.Header.Get("Origin"))
	}
}
