package gateway

import (
	"context"
	"math/rand/v2"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	domainRequestMinInterval = 300 * time.Millisecond
	domainRequestMaxJitter   = 600 * time.Millisecond
)

var sharedDomainRequestLimiter = newDomainRequestLimiter(domainRequestMinInterval, domainRequestMaxJitter)

type domainRequestLimiter struct {
	mu          sync.Mutex
	nextAllowed map[string]time.Time
	minInterval time.Duration
	maxJitter   time.Duration
}

func newDomainRequestLimiter(minInterval, maxJitter time.Duration) *domainRequestLimiter {
	if minInterval < 0 {
		minInterval = 0
	}
	if maxJitter < 0 {
		maxJitter = 0
	}

	return &domainRequestLimiter{
		nextAllowed: make(map[string]time.Time),
		minInterval: minInterval,
		maxJitter:   maxJitter,
	}
}

// waitForDomainRequestSlot blocks until the next request for this domain can be made.
func waitForDomainRequestSlot(ctx context.Context, targetURL *url.URL) error {
	return sharedDomainRequestLimiter.wait(ctx, targetURL)
}

// WaitForDomainRequestSlot exposes shared per-domain pacing for non-Colly requests.
func WaitForDomainRequestSlot(ctx context.Context, targetURL *url.URL) error {
	return waitForDomainRequestSlot(ctx, targetURL)
}

func (l *domainRequestLimiter) wait(ctx context.Context, targetURL *url.URL) error {
	if l == nil || targetURL == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}

	domain := canonicalDomain(targetURL)
	if domain == "" {
		return nil
	}

	delay, reservedUntil := l.reserveDelay(domain, time.Now())
	if delay <= 0 {
		return nil
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		l.rollbackReservation(domain, reservedUntil)
		return ctx.Err()
	}
}

func (l *domainRequestLimiter) reserveDelay(domain string, now time.Time) (time.Duration, time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()

	nextAllowed := l.nextAllowed[domain]
	if nextAllowed.Before(now) {
		nextAllowed = now
	}

	jitter := randomDuration(l.maxJitter)
	reservedUntil := nextAllowed.Add(l.minInterval + jitter)
	l.nextAllowed[domain] = reservedUntil
	return nextAllowed.Sub(now), reservedUntil
}

func (l *domainRequestLimiter) rollbackReservation(domain string, reservedUntil time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if currentReservation, exists := l.nextAllowed[domain]; exists && currentReservation.Equal(reservedUntil) {
		l.nextAllowed[domain] = time.Now()
	}
}

func canonicalDomain(targetURL *url.URL) string {
	return strings.ToLower(strings.TrimSpace(targetURL.Hostname()))
}

func randomDuration(max time.Duration) time.Duration {
	if max <= 0 {
		return 0
	}

	return time.Duration(rand.Int64N(int64(max)))
}
