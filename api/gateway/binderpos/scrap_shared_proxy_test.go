package binderpos

import (
	"context"
	"strings"
	"testing"
)

func TestScrapSharedProxy(t *testing.T) {
	i := impl{}

	t.Run("returns error when shared proxy is missing", func(t *testing.T) {
		t.Setenv("PROXY_URL", "")

		_, err := i.scrapSharedProxy(
			context.Background(),
			2,
			"Test Store",
			"https://example.com",
			"/search?q=%s",
			"abrade",
		)
		if err == nil {
			t.Fatalf("expected missing shared proxy to return an error")
		}
		if !strings.Contains(err.Error(), "no shared proxy configured for binderpos scraper") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("returns error when shared proxy is invalid", func(t *testing.T) {
		t.Setenv("PROXY_URL", "://bad-proxy")

		_, err := i.scrapSharedProxy(
			context.Background(),
			2,
			"Test Store",
			"https://example.com",
			"/search?q=%s",
			"abrade",
		)
		if err == nil {
			t.Fatalf("expected invalid shared proxy to return an error")
		}
		if !strings.Contains(err.Error(), "invalid shared proxy configured for binderpos scraper") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
