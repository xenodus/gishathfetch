package binderpos

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"testing"

	"mtg-price-checker-sg/gateway"
)

func TestNextBinderposStorefrontProxyURLRoundRobin(t *testing.T) {
	binderposDedicatedProxySeq.Store(0)

	urls := []string{"http://a:1", "http://b:2"}
	want := []string{"http://a:1", "http://b:2", "http://a:1"}
	for i, w := range want {
		gotURL := nextBinderposStorefrontProxyURL(urls)
		if gotURL != w {
			t.Fatalf("step %d: got %q, want %q", i, gotURL, w)
		}
	}
}

func TestNextBinderposStorefrontProxyURLSingleProxyRepeats(t *testing.T) {
	binderposDedicatedProxySeq.Store(0)
	urls := []string{"http://only:8080"}
	u0 := nextBinderposStorefrontProxyURL(urls)
	if u0 != "http://only:8080" {
		t.Fatalf("first slot: got %q", u0)
	}
	u1 := nextBinderposStorefrontProxyURL(urls)
	if u1 != "http://only:8080" {
		t.Fatalf("second slot: got %q", u1)
	}
}

func TestSearchByStorefrontAPIDynamicRequiresEnv(t *testing.T) {
	t.Setenv("DYNAMIC_PROXY", "")

	_, err := searchByStorefrontAPIDynamic(context.Background(), 1, "Store", "https://example.com", "shopify.example.com", "Abrade")
	if err == nil || err.Error() != "no dynamic proxy configured for binderpos storefront api" {
		t.Fatalf("unexpected error: %v", err)
	}
}

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
