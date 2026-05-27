package gateway

import (
	"context"
	"net/url"
	"testing"
	"time"
)

func TestDomainRequestLimiterWaitSameDomain(t *testing.T) {
	limiter := newDomainRequestLimiter(40 * time.Millisecond)
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
	limiter := newDomainRequestLimiter(50 * time.Millisecond)
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
	limiter := newDomainRequestLimiter(150 * time.Millisecond)
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

func TestDomainRequestLimiterWaitBypassesDisabledPacing(t *testing.T) {
	limiter := newDomainRequestLimiter(80 * time.Millisecond)
	targetURL, err := url.Parse("https://example.com/items")
	if err != nil {
		t.Fatalf("failed to parse URL: %v", err)
	}

	disabledCtx := WithDomainRequestPacingDisabled(context.Background())
	start := time.Now()
	if err := limiter.wait(disabledCtx, targetURL); err != nil {
		t.Fatalf("first disabled wait returned error: %v", err)
	}
	if err := limiter.wait(disabledCtx, targetURL); err != nil {
		t.Fatalf("second disabled wait returned error: %v", err)
	}
	if elapsed := time.Since(start); elapsed >= 40*time.Millisecond {
		t.Fatalf("expected disabled pacing waits to return immediately, got elapsed=%s", elapsed)
	}

	if _, exists := limiter.nextAllowed["example.com"]; exists {
		t.Fatalf("expected disabled pacing waits to avoid reserving a domain slot")
	}
}

func TestDomainRequestLimiterReserveDelayFirstRequestImmediate(t *testing.T) {
	limiter := newDomainRequestLimiter(200 * time.Millisecond)
	now := time.Unix(100, 0)

	delay, reservedUntil := limiter.reserveDelay("portal.binderpos.com", now)
	if delay != 0 {
		t.Fatalf("expected first request delay to be 0, got %s", delay)
	}
	if want := now.Add(200 * time.Millisecond); !reservedUntil.Equal(want) {
		t.Fatalf("expected first reservation until %s, got %s", want, reservedUntil)
	}
}

func TestDomainRequestLimiterReserveDelayUsesFixedMinimumInterval(t *testing.T) {
	limiter := newDomainRequestLimiter(200 * time.Millisecond)
	now := time.Unix(100, 0)

	_, firstReservedUntil := limiter.reserveDelay("portal.binderpos.com", now)
	secondNow := now.Add(100 * time.Millisecond)
	delay, secondReservedUntil := limiter.reserveDelay("portal.binderpos.com", secondNow)

	if want := 100 * time.Millisecond; delay != want {
		t.Fatalf("expected second request delay %s, got %s", want, delay)
	}
	if want := firstReservedUntil.Add(200 * time.Millisecond); !secondReservedUntil.Equal(want) {
		t.Fatalf("expected second reservation until %s, got %s", want, secondReservedUntil)
	}
}
