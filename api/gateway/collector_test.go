package gateway

import (
	"os"
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

	t.Run("retry 3 uses direct on final retry", func(t *testing.T) {
		t.Setenv("PROXY_URL", "http://shared:8080")
		mode, proxyURL := applyProxyForRetryAttempt(c, 3, "")
		if mode != "direct" {
			t.Fatalf("expected direct mode on final retry, got %q", mode)
		}
		if proxyURL != "" {
			t.Fatalf("expected empty proxy url on final retry, got %q", proxyURL)
		}
	})

	t.Run("retry 2 uses shared proxy", func(t *testing.T) {
		t.Setenv("PROXY_URL", "http://shared:8080")
		mode, proxyURL := applyProxyForRetryAttempt(c, 2, "")
		if mode != "shared" {
			t.Fatalf("expected shared mode, got %q", mode)
		}
		if proxyURL != "http://shared:8080" {
			t.Fatalf("unexpected shared proxy url: %q", proxyURL)
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
		mode, proxyURL := applyProxyForRetryAttemptWithPinnedDedicated(c, attempt, "", pinned, dedicatedProxyRetryThreshold)
		if mode != "dedicated" {
			t.Fatalf("expected dedicated mode for pinned proxy on attempt %d, got %q", attempt, mode)
		}
		if proxyURL != pinned {
			t.Fatalf("expected pinned proxy url %q on attempt %d, got %q", pinned, attempt, proxyURL)
		}
	}

	mode, proxyURL := applyProxyForRetryAttemptWithPinnedDedicated(c, 2, "", pinned, dedicatedProxyRetryThreshold)
	if mode != "shared" {
		t.Fatalf("expected shared mode on attempt 2, got %q", mode)
	}
	if proxyURL != "http://shared:8080" {
		t.Fatalf("expected shared proxy url on attempt 2, got %q", proxyURL)
	}

	mode, proxyURL = applyProxyForRetryAttemptWithPinnedDedicated(c, 3, "", pinned, dedicatedProxyRetryThreshold)
	if mode != "direct" {
		t.Fatalf("expected direct mode on final retry attempt 3, got %q", mode)
	}
	if proxyURL != "" {
		t.Fatalf("expected empty proxy url on final retry attempt 3, got %q", proxyURL)
	}
}

func TestApplyProxyForRetryAttemptWithPinnedDedicatedBinderpos(t *testing.T) {
	c := colly.NewCollector()
	t.Setenv("DEDICATED_PROXY_1", "9.9.9.9|9000|user|pass")
	t.Setenv("PROXY_URL", "http://shared:8080")

	pinned := "http://pinned:1234"
	for attempt := 0; attempt <= 1; attempt++ {
		mode, proxyURL := applyProxyForRetryAttemptWithPinnedDedicated(c, attempt, "", pinned, binderposDedicatedProxyRetryThreshold)
		if mode != "dedicated" || proxyURL != pinned {
			t.Fatalf("expected dedicated on attempt %d, got mode=%q url=%q", attempt, mode, proxyURL)
		}
	}

	mode, proxyURL := applyProxyForRetryAttemptWithPinnedDedicated(c, 2, "", pinned, binderposDedicatedProxyRetryThreshold)
	if mode != "shared" {
		t.Fatalf("expected PROXY_URL on attempt 2, got mode %q", mode)
	}
	if proxyURL != "http://shared:8080" {
		t.Fatalf("expected shared proxy url on attempt 2, got %q", proxyURL)
	}

	mode, proxyURL = applyProxyForRetryAttemptWithPinnedDedicated(c, 3, "", pinned, binderposDedicatedProxyRetryThreshold)
	if mode != "direct" {
		t.Fatalf("expected direct mode on final retry attempt 3, got %q", mode)
	}
	if proxyURL != "" {
		t.Fatalf("expected empty proxy url on final retry attempt 3, got %q", proxyURL)
	}
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
