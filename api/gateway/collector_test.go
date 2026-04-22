package gateway

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"mtg-price-checker-sg/gateway/util"

	"github.com/gocolly/colly/v2"
)

func TestParseRetryAfter(t *testing.T) {
	t.Run("seconds header", func(t *testing.T) {
		wait, ok := parseRetryAfter("3")
		if !ok {
			t.Fatalf("expected retry-after seconds to parse")
		}
		if wait != 3*time.Second {
			t.Fatalf("expected 3s wait, got %s", wait)
		}
	})

	t.Run("invalid header", func(t *testing.T) {
		if _, ok := parseRetryAfter("invalid"); ok {
			t.Fatalf("expected invalid retry-after to fail parsing")
		}
	})
}

func TestRetryDelay(t *testing.T) {
	t.Run("429 with retry-after", func(t *testing.T) {
		wait := retryDelay(429, "2", 0)
		if wait != 2*time.Second {
			t.Fatalf("expected 2s wait for retry-after header, got %s", wait)
		}
	})

	t.Run("429 without retry-after uses stronger base backoff", func(t *testing.T) {
		wait := retryDelay(429, "", 0)
		if wait < 2*time.Second || wait >= 3*time.Second {
			t.Fatalf("expected wait in [2s,3s) for first 429 retry, got %s", wait)
		}
	})

	t.Run("non-429 uses default base backoff", func(t *testing.T) {
		wait := retryDelay(500, "", 0)
		if wait < 1*time.Second || wait >= 1500*time.Millisecond {
			t.Fatalf("expected wait in [1s,1.5s) for first generic retry, got %s", wait)
		}
	})
}

func TestAdjustRetryDelayForContextDeadline(t *testing.T) {
	t.Run("nil context keeps wait duration", func(t *testing.T) {
		wait := 2 * time.Second
		got := adjustRetryDelayForContextDeadline(wait, nil, 1, defaultMaxRetries)
		if got != wait {
			t.Fatalf("expected wait to remain %s, got %s", wait, got)
		}
	})

	t.Run("final retry is immediate", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		got := adjustRetryDelayForContextDeadline(2*time.Second, ctx, defaultMaxRetries, defaultMaxRetries)
		if got != 0 {
			t.Fatalf("expected no wait for final retry, got %s", got)
		}
	})

	t.Run("no deadline keeps wait duration", func(t *testing.T) {
		ctx := context.Background()
		wait := 1500 * time.Millisecond
		got := adjustRetryDelayForContextDeadline(wait, ctx, 1, defaultMaxRetries)
		if got != wait {
			t.Fatalf("expected wait to remain %s, got %s", wait, got)
		}
	})

	t.Run("caps wait when deadline is near", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1200*time.Millisecond)
		defer cancel()

		got := adjustRetryDelayForContextDeadline(3*time.Second, ctx, 1, defaultMaxRetries)
		if got <= 0 {
			t.Fatalf("expected a reduced but positive wait, got %s", got)
		}
		if got >= 3*time.Second {
			t.Fatalf("expected capped wait below original duration, got %s", got)
		}
	})

	t.Run("returns zero when remaining time is too short", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		got := adjustRetryDelayForContextDeadline(1*time.Second, ctx, 1, defaultMaxRetries)
		if got != 0 {
			t.Fatalf("expected zero wait when deadline is too close, got %s", got)
		}
	})
}

func TestApplyProxyForRetryAttempt(t *testing.T) {
	c := colly.NewCollector()
	t.Setenv("DEDICATED_PROXY_1", "")
	t.Setenv("DEDICATED_PROXY_2", "")
	t.Setenv("DEDICATED_PROXY_3", "")
	t.Setenv("DEDICATED_PROXY_4", "")
	t.Setenv("DEDICATED_PROXY_5", "")

	t.Run("retry 0 uses dedicated proxy when available", func(t *testing.T) {
		t.Setenv("DEDICATED_PROXY_1", "1.1.1.1|8080|user|pass")
		mode, proxyURL := applyProxyForRetryAttempt(c, 0, "")
		if mode != "dedicated" {
			t.Fatalf("expected dedicated mode, got %q", mode)
		}
		if proxyURL != "http://user:pass@1.1.1.1:8080" {
			t.Fatalf("unexpected dedicated proxy url: %q", proxyURL)
		}
	})

	t.Run("retry 2 uses direct on final retry", func(t *testing.T) {
		t.Setenv("PROXY_URL", "http://shared:8080")
		mode, proxyURL := applyProxyForRetryAttempt(c, 2, "")
		if mode != "direct" {
			t.Fatalf("expected direct mode on final retry, got %q", mode)
		}
		if proxyURL != "" {
			t.Fatalf("expected empty proxy url on final retry, got %q", proxyURL)
		}
	})

	t.Run("retry 1 still uses dedicated proxy", func(t *testing.T) {
		t.Setenv("DEDICATED_PROXY_1", "2.2.2.2|9090|u1|p1")
		mode, proxyURL := applyProxyForRetryAttempt(c, 1, "")
		if mode != "dedicated" {
			t.Fatalf("expected dedicated mode, got %q", mode)
		}
		if proxyURL != "http://u1:p1@2.2.2.2:9090" {
			t.Fatalf("unexpected dedicated proxy url: %q", proxyURL)
		}
	})

	t.Run("retry 4 remains direct without shared proxy", func(t *testing.T) {
		_ = os.Unsetenv("PROXY_URL")
		mode, proxyURL := applyProxyForRetryAttempt(c, 4, "")
		if mode != "direct" {
			t.Fatalf("expected direct mode, got %q", mode)
		}
		if proxyURL != "" {
			t.Fatalf("expected empty proxy url, got %q", proxyURL)
		}
	})
}

func TestApplyProxyForRetryAttemptWithPinnedDedicated(t *testing.T) {
	c := colly.NewCollector()
	t.Setenv("DEDICATED_PROXY_1", "9.9.9.9|9000|user|pass")
	t.Setenv("PROXY_URL", "http://shared:8080")

	pinned := "http://pinned:1234"
	for attempt := 0; attempt <= 1; attempt++ {
		mode, proxyURL := applyProxyForRetryAttemptWithPinnedDedicated(c, attempt, "", pinned, dedicatedProxyRetryThreshold, defaultMaxRetries)
		if mode != "dedicated" {
			t.Fatalf("expected dedicated mode for pinned proxy on attempt %d, got %q", attempt, mode)
		}
		if proxyURL != pinned {
			t.Fatalf("expected pinned proxy url %q on attempt %d, got %q", pinned, attempt, proxyURL)
		}
	}

	mode, proxyURL := applyProxyForRetryAttemptWithPinnedDedicated(c, 2, "", pinned, dedicatedProxyRetryThreshold, defaultMaxRetries)
	if mode != "direct" {
		t.Fatalf("expected direct mode on final retry attempt 2, got %q", mode)
	}
	if proxyURL != "" {
		t.Fatalf("expected empty proxy url on final retry attempt 2, got %q", proxyURL)
	}
}

func TestApplyProxyForRetryAttemptWithPinnedDedicatedBinderpos(t *testing.T) {
	t.Run("default strategy uses dedicated then direct", func(t *testing.T) {
		c := colly.NewCollector()
		t.Setenv("DEDICATED_PROXY_1", "9.9.9.9|9000|user|pass")
		t.Setenv("PROXY_URL", "http://shared:8080")
		t.Setenv("USE_BINDERPOS_SHARED_PROXY_FALLBACK", "false")

		pinned := "http://pinned:1234"
		mode, proxyURL := applyProxyForRetryAttemptWithPinnedDedicated(c, 0, "", pinned, binderposDedicatedRetryThreshold(), binderposMaxRetries)
		if mode != "dedicated" || proxyURL != pinned {
			t.Fatalf("expected dedicated on attempt 0, got mode=%q url=%q", mode, proxyURL)
		}

		mode, proxyURL = applyProxyForRetryAttemptWithPinnedDedicated(c, 1, "", pinned, binderposDedicatedRetryThreshold(), binderposMaxRetries)
		if mode != "dedicated" || proxyURL != pinned {
			t.Fatalf("expected dedicated on attempt 1, got mode=%q url=%q", mode, proxyURL)
		}

		mode, proxyURL = applyProxyForRetryAttemptWithPinnedDedicated(c, 2, "", pinned, binderposDedicatedRetryThreshold(), binderposMaxRetries)
		if mode != "direct" {
			t.Fatalf("expected direct mode on final retry attempt 2, got %q", mode)
		}
		if proxyURL != "" {
			t.Fatalf("expected empty proxy url on final retry attempt 2, got %q", proxyURL)
		}
	})

	t.Run("rollback strategy uses shared fallback", func(t *testing.T) {
		c := colly.NewCollector()
		t.Setenv("DEDICATED_PROXY_1", "9.9.9.9|9000|user|pass")
		t.Setenv("PROXY_URL", "http://shared:8080")
		t.Setenv("USE_BINDERPOS_SHARED_PROXY_FALLBACK", "true")

		pinned := "http://pinned:1234"
		mode, proxyURL := applyProxyForRetryAttemptWithPinnedDedicated(c, 0, "", pinned, binderposDedicatedRetryThreshold(), binderposMaxRetries)
		if mode != "dedicated" || proxyURL != pinned {
			t.Fatalf("expected dedicated on attempt 0, got mode=%q url=%q", mode, proxyURL)
		}

		mode, proxyURL = applyProxyForRetryAttemptWithPinnedDedicated(c, 1, "", pinned, binderposDedicatedRetryThreshold(), binderposMaxRetries)
		if mode != "shared" {
			t.Fatalf("expected PROXY_URL on attempt 1, got mode %q", mode)
		}
		if proxyURL != "http://shared:8080" {
			t.Fatalf("expected shared proxy url on attempt 1, got %q", proxyURL)
		}

		mode, proxyURL = applyProxyForRetryAttemptWithPinnedDedicated(c, 2, "", pinned, binderposDedicatedRetryThreshold(), binderposMaxRetries)
		if mode != "direct" {
			t.Fatalf("expected direct mode on final retry attempt 2, got %q", mode)
		}
		if proxyURL != "" {
			t.Fatalf("expected empty proxy url on final retry attempt 2, got %q", proxyURL)
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
