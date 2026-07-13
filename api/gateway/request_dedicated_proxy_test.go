package gateway

import (
	"context"
	"testing"

	"github.com/gocolly/colly/v2"
)

func TestRequestDedicatedProxyURL(t *testing.T) {
	t.Run("returns false when unset", func(t *testing.T) {
		if _, ok := RequestDedicatedProxyURL(context.Background()); ok {
			t.Fatal("expected no request dedicated proxy on background context")
		}
	})

	t.Run("round trips pinned proxy URL", func(t *testing.T) {
		const want = "http://user:pass@10.0.0.1:8080"
		ctx := WithRequestDedicatedProxy(context.Background(), want)
		got, ok := RequestDedicatedProxyURL(ctx)
		if !ok || got != want {
			t.Fatalf("expected %q, got %q (ok=%t)", want, got, ok)
		}
	})

	t.Run("ignores blank proxy URL", func(t *testing.T) {
		ctx := WithRequestDedicatedProxy(context.Background(), "  ")
		if _, ok := RequestDedicatedProxyURL(ctx); ok {
			t.Fatal("expected blank proxy URL to be ignored")
		}
	})
}

func TestSelectOutboundProxyUsesRequestDedicatedProxy(t *testing.T) {
	const requestProxy = "http://user:pass@10.0.0.9:8080"
	mode, proxyURL := selectOutboundProxy("http://lease:1", requestProxy)
	if mode != "dedicated" || proxyURL != requestProxy {
		t.Fatalf("expected request dedicated proxy, got mode=%q url=%q", mode, proxyURL)
	}
}

func TestApplyInitialProxyUsesRequestDedicatedProxy(t *testing.T) {
	c := colly.NewCollector()
	const requestProxy = "http://user:pass@10.0.0.9:8080"
	mode, proxyURL := applyInitialProxy(c, "http://lease:1", requestProxy)
	if mode != "dedicated" || proxyURL != requestProxy {
		t.Fatalf("expected request dedicated proxy, got mode=%q url=%q", mode, proxyURL)
	}
}
