package gateway

import (
	"os"
	"testing"
	"time"

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

	t.Run("retry 3 uses shared proxy", func(t *testing.T) {
		t.Setenv("PROXY_URL", "http://shared:8080")
		mode, proxyURL := applyProxyForRetryAttempt(c, 3, "")
		if mode != "shared" {
			t.Fatalf("expected shared mode, got %q", mode)
		}
		if proxyURL != "http://shared:8080" {
			t.Fatalf("unexpected shared proxy url: %q", proxyURL)
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

	t.Run("retry 3 falls back to direct without shared proxy", func(t *testing.T) {
		_ = os.Unsetenv("PROXY_URL")
		mode, proxyURL := applyProxyForRetryAttempt(c, 3, "")
		if mode != "direct" {
			t.Fatalf("expected direct mode, got %q", mode)
		}
		if proxyURL != "" {
			t.Fatalf("expected empty proxy url, got %q", proxyURL)
		}
	})
}
