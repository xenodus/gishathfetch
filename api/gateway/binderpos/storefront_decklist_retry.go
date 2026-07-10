package binderpos

import (
	"context"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"strconv"
	"strings"
	"time"

	"mtg-price-checker-sg/gateway"
)

// retriableDecklistStatuses are upstream responses that signal a transient
// load or rate-limit condition on the shared portal host. They are retried with
// backoff (honoring Retry-After) instead of immediately failing over to a
// heavier, more expensive strategy, which reduces the user-visible 503 rate
// without changing card data.
var retriableDecklistStatuses = map[int]struct{}{
	http.StatusTooManyRequests:     {}, // 429
	http.StatusInternalServerError: {}, // 500
	http.StatusBadGateway:          {}, // 502
	http.StatusServiceUnavailable:  {}, // 503
	http.StatusGatewayTimeout:      {}, // 504
}

func isRetriableDecklistStatus(status int) bool {
	_, ok := retriableDecklistStatuses[status]
	return ok
}

// doDecklistRequestWithRetry sends the decklist request, retrying transient
// rate-limit/5xx responses and network errors with jittered exponential backoff
// bounded by ctx. newRequest rebuilds the request for every send so each retry
// carries a fresh User-Agent (set by the caller) and a re-readable body. On the
// dynamic-proxy attempt every retry egresses from a fresh rotating IP, which is
// the most effective mitigation for per-IP throttling on the shared portal host.
//
// On success it returns the response with its body still open for the caller to
// decode. On failure it returns the last observed error and no response.
func doDecklistRequestWithRetry(ctx context.Context, client *http.Client, newRequest func() (*http.Request, error)) (*http.Response, error) {
	var lastErr error

	for attempt := range binderposDecklistMaxAttempts {
		req, err := newRequest()
		if err != nil {
			return nil, err
		}
		if err := gateway.WaitForDomainRequestSlot(ctx, req.URL); err != nil {
			return nil, err
		}

		res, err := client.Do(req)
		if err != nil {
			lastErr = err
			if isLastDecklistAttempt(attempt) || !waitBeforeDecklistRetry(ctx, decklistBackoffDelay(attempt)) {
				return nil, lastErr
			}
			continue
		}

		if res.StatusCode == http.StatusOK {
			return res, nil
		}

		body, _ := io.ReadAll(res.Body)
		res.Body.Close()
		lastErr = fmt.Errorf("binderpos decklist request failed status=%d body=%s", res.StatusCode, strings.TrimSpace(string(body)))

		if !isRetriableDecklistStatus(res.StatusCode) || isLastDecklistAttempt(attempt) {
			return nil, lastErr
		}

		delay := parseRetryAfter(res.Header.Get("Retry-After"))
		if delay <= 0 {
			delay = decklistBackoffDelay(attempt)
		}
		if !waitBeforeDecklistRetry(ctx, delay) {
			return nil, lastErr
		}
	}

	return nil, lastErr
}

func isLastDecklistAttempt(attempt int) bool {
	return attempt >= binderposDecklistMaxAttempts-1
}

// decklistBackoffDelay returns an equal-jitter exponential backoff for the given
// zero-based attempt, capped at binderposDecklistRetryMaxDelay. Equal jitter
// keeps a guaranteed minimum spacing while still desynchronizing concurrent
// callers that all retry against the same host.
func decklistBackoffDelay(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}

	base := binderposDecklistRetryBaseDelay << attempt
	if base <= 0 || base > binderposDecklistRetryMaxDelay {
		base = binderposDecklistRetryMaxDelay
	}

	half := base / 2
	if half <= 0 {
		return base
	}
	return half + time.Duration(rand.Int64N(int64(half)+1))
}

// waitBeforeDecklistRetry sleeps for delay unless ctx is cancelled or the
// remaining ctx budget cannot accommodate it. It returns true only when the
// wait completed and a retry should proceed.
func waitBeforeDecklistRetry(ctx context.Context, delay time.Duration) bool {
	if ctx == nil {
		ctx = context.Background()
	}
	if delay <= 0 {
		return ctx.Err() == nil
	}
	if deadline, ok := ctx.Deadline(); ok && time.Until(deadline) <= delay {
		return false
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-timer.C:
		return true
	case <-ctx.Done():
		return false
	}
}

// parseRetryAfter parses a Retry-After header expressed either as a number of
// seconds or an HTTP date. It returns 0 when the header is absent or
// unparseable, and caps the wait so a huge value cannot stall the attempt.
func parseRetryAfter(value string) time.Duration {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}

	if seconds, err := strconv.Atoi(value); err == nil {
		if seconds <= 0 {
			return 0
		}
		return capRetryAfter(time.Duration(seconds) * time.Second)
	}

	if when, err := http.ParseTime(value); err == nil {
		delay := time.Until(when)
		if delay <= 0 {
			return 0
		}
		return capRetryAfter(delay)
	}

	return 0
}

func capRetryAfter(delay time.Duration) time.Duration {
	if delay > binderposDecklistRetryMaxDelay {
		return binderposDecklistRetryMaxDelay
	}
	return delay
}
