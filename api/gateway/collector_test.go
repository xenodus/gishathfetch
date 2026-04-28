package gateway

import (
	"context"
	"strings"
	"testing"
	"time"

	"mtg-price-checker-sg/gateway/util"

	"github.com/gocolly/colly/v2"
)

func TestInitialProxy(t *testing.T) {
	c := colly.NewCollector()
	for i := 1; i <= 7; i++ {
		t.Setenv("DEDICATED_PROXY_"+string(rune('0'+i)), "")
	}

	t.Run("uses leased dedicated when provided", func(t *testing.T) {
		leased := "http://lease:1"
		mode, proxyURL := applyInitialProxy(c, leased)
		if mode != "dedicated" || proxyURL != leased {
			t.Fatalf("expected dedicated leased, got mode=%q url=%q", mode, proxyURL)
		}
	})

	t.Run("picks random dedicated when no lease", func(t *testing.T) {
		c2 := colly.NewCollector()
		for i := 2; i <= 7; i++ {
			t.Setenv("DEDICATED_PROXY_"+string(rune('0'+i)), "")
		}
		t.Setenv("DEDICATED_PROXY_1", "1.1.1.1|8080|user|pass")
		mode, proxyURL := applyInitialProxy(c2, "")
		if mode != "dedicated" {
			t.Fatalf("expected dedicated mode, got %q", mode)
		}
		if proxyURL != "http://user:pass@1.1.1.1:8080" {
			t.Fatalf("unexpected dedicated proxy url: %q", proxyURL)
		}
	})
}

func TestDedicatedProxyLeasePoolAcquireRelease(t *testing.T) {
	pool := newDedicatedProxyLeasePool()
	proxyURLs := []string{"http://a:1", "http://b:2"}

	first, ok := pool.acquire(proxyURLs)
	if !ok || first == "" {
		t.Fatalf("expected first acquire to succeed")
	}
	second, ok := pool.acquire(proxyURLs)
	if !ok || second == "" {
		t.Fatalf("expected second acquire to succeed")
	}
	if first == second {
		t.Fatalf("expected distinct leased proxies, got duplicate %q", first)
	}

	acquired := make(chan string, 1)
	go func() {
		third, ok := pool.acquire(proxyURLs)
		if ok {
			acquired <- third
			return
		}
		acquired <- ""
	}()

	select {
	case v := <-acquired:
		t.Fatalf("expected third acquire to block before release, got %q", v)
	case <-time.After(50 * time.Millisecond):
		// expected: blocked
	}

	pool.release(first)

	select {
	case third := <-acquired:
		if third != first {
			t.Fatalf("expected released proxy %q to be re-acquired, got %q", first, third)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("expected third acquire to unblock after release")
	}

	pool.release(second)
	pool.release(first)
}

func TestDedicatedProxyURLHelpers(t *testing.T) {
	t.Run("build dedicated proxy url", func(t *testing.T) {
		proxyURL, ok := util.BuildDedicatedProxyURL(util.DedicatedProxy{
			Host:     "1.1.1.1",
			Port:     "8080",
			Username: "user",
			Password: "pass",
		})
		if !ok {
			t.Fatalf("expected proxy url build to succeed")
		}
		if proxyURL != "http://user:pass@1.1.1.1:8080" {
			t.Fatalf("unexpected proxy url: %q", proxyURL)
		}
	})
}

func TestFormatProxyContext(t *testing.T) {
	t.Run("empty mode and proxy default values", func(t *testing.T) {
		got := formatProxyContext("", "")
		if got != "proxy_mode=unknown proxy=none" {
			t.Fatalf("unexpected proxy context: %q", got)
		}
	})

	t.Run("uses dedicated proxy env label when mapped", func(t *testing.T) {
		t.Setenv("DEDICATED_PROXY_1", "4.4.4.4|8080|user|pass")
		got := formatProxyContext("dedicated", "http://user:pass@4.4.4.4:8080")
		if got != "proxy_mode=dedicated proxy=DEDICATED_PROXY_1" {
			t.Fatalf("unexpected proxy context: %q", got)
		}
	})

	t.Run("uses shared proxy env label", func(t *testing.T) {
		t.Setenv("PROXY_URL", "http://shared-proxy:8080")
		got := formatProxyContext("shared", "http://shared-proxy:8080")
		if got != "proxy_mode=shared proxy=PROXY_URL" {
			t.Fatalf("unexpected proxy context: %q", got)
		}
	})
}

func TestResolveProxyLabel(t *testing.T) {
	t.Run("returns none for empty proxy", func(t *testing.T) {
		if got := resolveProxyLabel("dedicated", ""); got != "none" {
			t.Fatalf("expected none, got %q", got)
		}
	})

	t.Run("matches dedicated env key by URL", func(t *testing.T) {
		t.Setenv("DEDICATED_PROXY_1", "10.0.0.1|1111|u1|p1")
		t.Setenv("DEDICATED_PROXY_2", "10.0.0.2|2222|u2|p2")
		if got := resolveProxyLabel("dedicated", "http://u2:p2@10.0.0.2:2222"); got != "DEDICATED_PROXY_2" {
			t.Fatalf("expected DEDICATED_PROXY_2, got %q", got)
		}
	})

	t.Run("matches shared env key by URL", func(t *testing.T) {
		t.Setenv("PROXY_URL", "http://shared:3333")
		if got := resolveProxyLabel("shared", "http://shared:3333"); got != "PROXY_URL" {
			t.Fatalf("expected PROXY_URL, got %q", got)
		}
	})

	t.Run("falls back to mode label when unmapped", func(t *testing.T) {
		if got := resolveProxyLabel("dedicated", "http://unknown:4444"); got != "dedicated-configured" {
			t.Fatalf("expected dedicated-configured fallback, got %q", got)
		}
		if got := resolveProxyLabel("shared", "http://unknown:5555"); got != "shared-configured" {
			t.Fatalf("expected shared-configured fallback, got %q", got)
		}
		if got := resolveProxyLabel("unknown", "http://unknown:6666"); got != "configured" {
			t.Fatalf("expected configured fallback, got %q", got)
		}
	})
}

func TestVisitWithProxyInfo(t *testing.T) {
	c := NewOptimizedCollectorNoRetry(context.Background())
	err := VisitWithProxyInfo(c, "http://[::1")
	if err == nil {
		t.Fatalf("expected visit error for malformed URL")
	}
	if !strings.Contains(err.Error(), "proxy_mode=") {
		t.Fatalf("expected proxy context in error, got %q", err)
	}
}

func TestVisitWithProxyInfoDirectCollector(t *testing.T) {
	t.Setenv("DEDICATED_PROXY_1", "4.4.4.4|8080|user|pass")

	c := NewOptimizedCollectorNoRetryDirect(context.Background())
	err := VisitWithProxyInfo(c, "http://127.0.0.1:1")
	if err == nil {
		t.Fatalf("expected visit error for unreachable endpoint")
	}
	if !strings.Contains(err.Error(), "proxy_mode=direct") {
		t.Fatalf("expected direct proxy context in error, got %q", err)
	}
	if !strings.Contains(err.Error(), "proxy=none") {
		t.Fatalf("expected no proxy label for direct collector, got %q", err)
	}
}

func TestSeedProxyContextIfMissing(t *testing.T) {
	t.Run("initializes mode and url when missing", func(t *testing.T) {
		ctx := colly.NewContext()
		seedProxyContextIfMissing(ctx, "dedicated", "http://proxy:8080")

		if got := ctx.Get("last_proxy_mode"); got != "dedicated" {
			t.Fatalf("expected dedicated mode, got %q", got)
		}
		if got := ctx.Get("last_proxy_url"); got != "http://proxy:8080" {
			t.Fatalf("expected proxy URL to be seeded, got %q", got)
		}
	})

	t.Run("does not overwrite direct context with empty url", func(t *testing.T) {
		ctx := colly.NewContext()
		ctx.Put("last_proxy_mode", "direct")
		ctx.Put("last_proxy_url", "")

		seedProxyContextIfMissing(ctx, "shared", "http://shared:8080")

		if got := ctx.Get("last_proxy_mode"); got != "direct" {
			t.Fatalf("expected direct mode to remain, got %q", got)
		}
		if got := ctx.Get("last_proxy_url"); got != "" {
			t.Fatalf("expected direct mode proxy URL to stay empty, got %q", got)
		}
	})

	t.Run("ignores nil context", func(t *testing.T) {
		seedProxyContextIfMissing(nil, "dedicated", "http://proxy:8080")
	})
}

func TestRandomBrowserUserAgent(t *testing.T) {
	got := RandomBrowserUserAgent()
	if strings.TrimSpace(got) == "" {
		t.Fatalf("expected random browser user-agent to be non-empty")
	}

	found := false
	for _, candidate := range browserUserAgents {
		if got == candidate {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected user-agent to be chosen from browserUserAgents list, got %q", got)
	}
}
