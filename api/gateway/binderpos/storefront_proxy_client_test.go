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
	want := []struct {
		url    string
		direct bool
	}{
		{url: "http://a:1", direct: false},
		{url: "http://b:2", direct: false},
		{url: "", direct: true},
		{url: "http://a:1", direct: false},
	}
	for i, w := range want {
		gotURL, gotDirect := nextBinderposStorefrontProxyURL(urls)
		if gotURL != w.url || gotDirect != w.direct {
			t.Fatalf("step %d: got (%q, %v), want (%q, %v)", i, gotURL, gotDirect, w.url, w.direct)
		}
	}
}

func TestNextBinderposStorefrontProxyURLSingleProxyIncludesDirect(t *testing.T) {
	binderposDedicatedProxySeq.Store(0)
	urls := []string{"http://only:8080"}
	u0, d0 := nextBinderposStorefrontProxyURL(urls)
	if d0 || u0 != "http://only:8080" {
		t.Fatalf("first slot: got (%q, %v)", u0, d0)
	}
	u1, d1 := nextBinderposStorefrontProxyURL(urls)
	if !d1 || u1 != "" {
		t.Fatalf("second slot (direct): got (%q, %v)", u1, d1)
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
