package gateway

import (
	"context"
	"net/url"
	"testing"
	"time"
)

func TestDomainRequestLimiterWaitSameDomain(t *testing.T) {
	limiter := newDomainRequestLimiter(40*time.Millisecond, 0)
	targetURL, err := url.Parse("https://example.com/products")
	if err != nil {
		t.Fatalf("failed to parse URL: %v", err)
	}

	start := time.Now()
	if err := limiter.wait(context.Background(), targetURL); err != nil {
		t.Fatalf("first wait returned error: %v", err)
	}
	if err := limiter.wait(context.Background(), targetURL); err != nil {
		t.Fatalf("second wait returned error: %v", err)
	}

	elapsed := time.Since(start)
	if elapsed < 35*time.Millisecond {
		t.Fatalf("expected second request to be delayed, got elapsed=%s", elapsed)
	}
}

func TestDomainRequestLimiterWaitDifferentDomains(t *testing.T) {
	limiter := newDomainRequestLimiter(50*time.Millisecond, 0)
	firstURL, err := url.Parse("https://example.com/a")
	if err != nil {
		t.Fatalf("failed to parse first URL: %v", err)
	}
	secondURL, err := url.Parse("https://other.com/b")
	if err != nil {
		t.Fatalf("failed to parse second URL: %v", err)
	}

	start := time.Now()
	if err := limiter.wait(context.Background(), firstURL); err != nil {
		t.Fatalf("first wait returned error: %v", err)
	}
	if err := limiter.wait(context.Background(), secondURL); err != nil {
		t.Fatalf("second wait returned error: %v", err)
	}

	elapsed := time.Since(start)
	if elapsed >= 35*time.Millisecond {
		t.Fatalf("expected different domains to avoid throttling delay, got elapsed=%s", elapsed)
	}
}

func TestDomainRequestLimiterRollbackOnCancellation(t *testing.T) {
	limiter := newDomainRequestLimiter(150*time.Millisecond, 0)
	targetURL, err := url.Parse("https://example.com/items")
	if err != nil {
		t.Fatalf("failed to parse URL: %v", err)
	}

	if err := limiter.wait(context.Background(), targetURL); err != nil {
		t.Fatalf("first wait returned error: %v", err)
	}

	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := limiter.wait(canceledCtx, targetURL); err == nil {
		t.Fatalf("expected canceled wait to return an error")
	}

	start := time.Now()
	if err := limiter.wait(context.Background(), targetURL); err != nil {
		t.Fatalf("post-cancellation wait returned error: %v", err)
	}
	elapsed := time.Since(start)

	if elapsed >= 60*time.Millisecond {
		t.Fatalf("expected rollback to clear pending reservation, got elapsed=%s", elapsed)
	}
}
