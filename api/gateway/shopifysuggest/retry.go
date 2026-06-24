package shopifysuggest

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

const (
	// suggestRetryMaxAttempts bounds how many times a single suggest or product
	// JSON request is sent (initial try plus retries) when Shopify responds with
	// a transient rate-limit/5xx status or a network error.
	suggestRetryMaxAttempts = 3
	// suggestRetryBaseDelay is the first backoff step when Retry-After is absent.
	suggestRetryBaseDelay = 300 * time.Millisecond
	// suggestRetryMaxDelay caps a single backoff/Retry-After wait so a large or
	// hostile Retry-After value cannot stall the attempt.
	suggestRetryMaxDelay = 2500 * time.Millisecond
)

// retriableSuggestStatuses are upstream responses that signal a transient load
// or rate-limit condition on a Shopify storefront. They are retried with
// backoff honoring Retry-After instead of immediately failing over to the next
// transport, which reduces unnecessary proxy churn.
var retriableSuggestStatuses = map[int]struct{}{
	http.StatusTooManyRequests:     {}, // 429
	http.StatusInternalServerError: {}, // 500
	http.StatusBadGateway:          {}, // 502
	http.StatusServiceUnavailable:  {}, // 503
	http.StatusGatewayTimeout:      {}, // 504
}

func isRetriableSuggestStatus(status int) bool {
	_, ok := retriableSuggestStatuses[status]
	return ok
}

// doSuggestGETWithRetry sends a GET request, retrying transient rate-limit/5xx
// responses and network errors with jittered exponential backoff bounded by
// ctx. Each send uses a fresh User-Agent. On success it returns the response
// body. On failure it returns the last observed error.
func doSuggestGETWithRetry(ctx context.Context, client *http.Client, requestURL string) ([]byte, error) {
	var lastErr error

	for attempt := 0; attempt < suggestRetryMaxAttempts; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", gateway.RandomBrowserUserAgent())

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			if isLastSuggestRetryAttempt(attempt) || !waitBeforeSuggestRetry(ctx, suggestBackoffDelay(attempt)) {
				return nil, lastErr
			}
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			if isLastSuggestRetryAttempt(attempt) || !waitBeforeSuggestRetry(ctx, suggestBackoffDelay(attempt)) {
				return nil, lastErr
			}
			continue
		}

		if resp.StatusCode == http.StatusOK {
			return body, nil
		}

		lastErr = fmt.Errorf("shopifysuggest: unexpected status %d", resp.StatusCode)
		if !isRetriableSuggestStatus(resp.StatusCode) || isLastSuggestRetryAttempt(attempt) {
			return nil, lastErr
		}

		delay := parseSuggestRetryAfter(resp.Header.Get("Retry-After"))
		if delay <= 0 {
			delay = suggestBackoffDelay(attempt)
		}
		if !waitBeforeSuggestRetry(ctx, delay) {
			return nil, lastErr
		}
	}

	return nil, lastErr
}

func isLastSuggestRetryAttempt(attempt int) bool {
	return attempt >= suggestRetryMaxAttempts-1
}

// suggestBackoffDelay returns an equal-jitter exponential backoff for the given
// zero-based attempt, capped at suggestRetryMaxDelay.
func suggestBackoffDelay(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}

	base := suggestRetryBaseDelay << attempt
	if base <= 0 || base > suggestRetryMaxDelay {
		base = suggestRetryMaxDelay
	}

	half := base / 2
	if half <= 0 {
		return base
	}
	return half + time.Duration(rand.Int64N(int64(half)+1))
}

// waitBeforeSuggestRetry sleeps for delay unless ctx is cancelled or the
// remaining ctx budget cannot accommodate it.
func waitBeforeSuggestRetry(ctx context.Context, delay time.Duration) bool {
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

// parseSuggestRetryAfter parses a Retry-After header expressed either as a
// number of seconds or an HTTP date. It returns 0 when the header is absent or
// unparseable, and caps the wait so a huge value cannot stall the attempt.
func parseSuggestRetryAfter(value string) time.Duration {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}

	if seconds, err := strconv.Atoi(value); err == nil {
		if seconds <= 0 {
			return 0
		}
		return capSuggestRetryAfter(time.Duration(seconds) * time.Second)
	}

	if when, err := http.ParseTime(value); err == nil {
		delay := time.Until(when)
		if delay <= 0 {
			return 0
		}
		return capSuggestRetryAfter(delay)
	}

	return 0
}

func capSuggestRetryAfter(delay time.Duration) time.Duration {
	if delay > suggestRetryMaxDelay {
		return suggestRetryMaxDelay
	}
	return delay
}
