package gateway

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"mtg-price-checker-sg/gateway/util"
)

const (
	// outboundGETMaxAttemptsPerTransport bounds how many times a single transport
	// is tried when the upstream responds with 429 Too Many Requests. Retries
	// back off between sends so a rate-limited host is not hammered.
	outboundGETMaxAttemptsPerTransport = 3
	outboundGETRetryBaseDelay          = 300 * time.Millisecond
	outboundGETRetryMaxDelay           = 2500 * time.Millisecond
)

type outboundAttempt struct {
	strategy string
	proxyURL string
	client   *http.Client
}

// DoOutboundGET performs a GET with direct, dedicated-proxy, and dynamic-proxy
// fallback. 403 responses advance to the next transport; 429 responses retry
// with backoff on the same transport before failing over.
func DoOutboundGET(
	ctx context.Context,
	requestURL string,
	opts OutboundRequestOptions,
	timeout time.Duration,
) (*http.Response, error) {
	return DoOutboundRoundTrip(ctx, opts, timeout, func() (*http.Request, error) {
		return http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	})
}

// DoOutboundRoundTrip performs an HTTP round trip with direct, dedicated-proxy,
// and dynamic-proxy fallback. 403 responses advance to the next transport; 429
// responses retry with backoff on the same transport before failing over.
// buildReq is called for each send so callers can supply a fresh request body
// when needed.
func DoOutboundRoundTrip(
	ctx context.Context,
	opts OutboundRequestOptions,
	timeout time.Duration,
	buildReq func() (*http.Request, error),
) (*http.Response, error) {
	var failures []string
	for _, attempt := range buildOutboundGETAttempts(ctx, timeout, opts.SkipDirect, opts.OnlyProxyURL) {
		resp, failure, ok, err := doOutboundAttempt(ctx, attempt, opts, buildReq)
		if err != nil {
			return nil, err
		}
		if ok {
			return resp, nil
		}
		if failure != "" {
			failures = append(failures, failure)
		}
	}
	if len(failures) == 0 {
		return nil, fmt.Errorf("outbound request failed")
	}
	return nil, fmt.Errorf("outbound request failed: %s", strings.Join(failures, "; "))
}

func doOutboundAttempt(
	ctx context.Context,
	attempt outboundAttempt,
	opts OutboundRequestOptions,
	buildReq func() (*http.Request, error),
) (*http.Response, string, bool, error) {
	proxyDesc := outboundProxyDescription(attempt)
	var lastFailure string

	for retry := range outboundGETMaxAttemptsPerTransport {
		req, err := buildReq()
		if err != nil {
			return nil, "", false, err
		}
		if err := PrepareOutboundRequest(ctx, req, opts); err != nil {
			return nil, "", false, err
		}

		if retry == 0 {
			log.Printf("outbound request: trying %s url=%s", proxyDesc, outboundRequestURL(req))
		} else {
			log.Printf("outbound request: retrying %s attempt=%d url=%s", proxyDesc, retry+1, outboundRequestURL(req))
		}

		resp, err := attempt.client.Do(req)
		if err != nil {
			lastFailure = fmt.Sprintf("%s: %v", attempt.strategy, err)
			log.Printf("outbound request: failed %s: %v", proxyDesc, err)
			return nil, lastFailure, false, nil
		}

		if resp.StatusCode == http.StatusForbidden {
			lastFailure = outboundStatusFailure(attempt.strategy, resp)
			log.Printf("outbound request: failed %s: status %d", proxyDesc, resp.StatusCode)
			return nil, lastFailure, false, nil
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			lastFailure = outboundStatusFailure(attempt.strategy, resp)
			log.Printf("outbound request: failed %s: status %d", proxyDesc, resp.StatusCode)
			if isLastOutboundGETRetry(retry) || !waitBeforeOutboundGETRetry(ctx, outboundGETRetryDelay(retry, resp.Header.Get("Retry-After"))) {
				return nil, lastFailure, false, nil
			}
			continue
		}

		log.Printf("outbound request: succeeded %s status=%d url=%s", proxyDesc, resp.StatusCode, outboundRequestURL(req))
		return resp, "", true, nil
	}

	return nil, lastFailure, false, nil
}

func buildOutboundGETAttempts(ctx context.Context, timeout time.Duration, skipDirect bool, onlyProxyURL string) []outboundAttempt {
	if onlyProxyURL != "" {
		client, err := newProxyHTTPClient(onlyProxyURL, timeout)
		if err != nil {
			return nil
		}
		return []outboundAttempt{{
			strategy: "ck-pricelist-proxy",
			proxyURL: onlyProxyURL,
			client:   client,
		}}
	}

	var attempts []outboundAttempt
	if !skipDirect {
		attempts = append(attempts, outboundAttempt{
			strategy: "direct",
			client:   &http.Client{Timeout: timeout},
		})
	}

	// Match colly's selectOutboundProxy policy: one dedicated proxy per search.
	// When the controller pins a request-scoped lease, reuse that URL instead of
	// picking a new random slot for each outbound store.
	if proxyURL, ok := dedicatedProxyURLForOutbound(ctx); ok {
		client, err := newProxyHTTPClient(proxyURL, timeout)
		if err == nil {
			attempts = append(attempts, outboundAttempt{
				strategy: dedicatedProxyStrategyName(proxyURL),
				proxyURL: proxyURL,
				client:   client,
			})
		}
	}

	if client, ok := dynamicProxyHTTPClient(timeout); ok {
		attempts = append(attempts, outboundAttempt{
			strategy: "dynamic",
			proxyURL: DynamicProxyURL(),
			client:   client,
		})
	}

	return attempts
}

func dedicatedProxyURLForOutbound(ctx context.Context) (string, bool) {
	if pinned, ok := RequestDedicatedProxyURL(ctx); ok {
		return pinned, true
	}
	return RandomDedicatedProxyURL()
}

func dedicatedProxyStrategyName(proxyURL string) string {
	for idx, configuredURL := range util.GetDedicatedProxyURLs() {
		if configuredURL == proxyURL {
			return fmt.Sprintf("dedicated-%d", idx+1)
		}
	}
	return "dedicated"
}

func dynamicProxyHTTPClient(timeout time.Duration) (*http.Client, bool) {
	proxyURL := DynamicProxyURL()
	if proxyURL == "" {
		return nil, false
	}

	client, err := newProxyHTTPClient(proxyURL, timeout)
	if err != nil {
		return nil, false
	}
	return client, true
}

func isLastOutboundGETRetry(retry int) bool {
	return retry >= outboundGETMaxAttemptsPerTransport-1
}

func outboundGETRetryDelay(retry int, retryAfter string) time.Duration {
	if delay := parseOutboundRetryAfter(retryAfter); delay > 0 {
		return delay
	}
	return outboundGETBackoffDelay(retry)
}

func outboundGETBackoffDelay(retry int) time.Duration {
	if retry < 0 {
		retry = 0
	}

	base := outboundGETRetryBaseDelay << retry
	if base <= 0 || base > outboundGETRetryMaxDelay {
		base = outboundGETRetryMaxDelay
	}

	half := base / 2
	if half <= 0 {
		return base
	}
	return half + time.Duration(rand.Int64N(int64(half)+1))
}

func waitBeforeOutboundGETRetry(ctx context.Context, delay time.Duration) bool {
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

func parseOutboundRetryAfter(value string) time.Duration {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}

	if seconds, err := strconv.Atoi(value); err == nil {
		if seconds <= 0 {
			return 0
		}
		return capOutboundRetryAfter(time.Duration(seconds) * time.Second)
	}

	if when, err := http.ParseTime(value); err == nil {
		delay := time.Until(when)
		if delay <= 0 {
			return 0
		}
		return capOutboundRetryAfter(delay)
	}

	return 0
}

func capOutboundRetryAfter(delay time.Duration) time.Duration {
	if delay > outboundGETRetryMaxDelay {
		return outboundGETRetryMaxDelay
	}
	return delay
}

func outboundStatusFailure(strategy string, resp *http.Response) string {
	msg := fmt.Sprintf("%s: status %d", strategy, resp.StatusCode)
	if resp == nil {
		return msg
	}

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 200))
	resp.Body.Close()

	if cfRay := resp.Header.Get("cf-ray"); cfRay != "" {
		msg += " cf-ray=" + cfRay
	}
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return msg
	}
	if len(trimmed) > 120 {
		trimmed = trimmed[:120] + "..."
	}
	return msg + " (" + trimmed + ")"
}

func outboundProxyDescription(attempt outboundAttempt) string {
	return formatProxyContext(outboundProxyMode(attempt.strategy), attempt.proxyURL)
}

func outboundProxyMode(strategy string) string {
	switch {
	case strategy == "direct":
		return "direct"
	case strategy == "dynamic":
		return "dynamic"
	case strings.HasPrefix(strategy, "dedicated-"):
		return "dedicated"
	case strategy == "ck-pricelist-proxy":
		return "ck-pricelist"
	default:
		return strategy
	}
}

func outboundRequestURL(req *http.Request) string {
	if req == nil || req.URL == nil {
		return ""
	}
	return req.URL.Redacted()
}

// NewOutboundHTTPClient returns an HTTP client that routes through a random dedicated
// proxy when configured, otherwise dynamic proxy, otherwise direct. The policy matches
// optimized colly collectors used by non-BinderPOS scrapers.
func NewOutboundHTTPClient(timeout time.Duration) (*http.Client, error) {
	_, proxyURL := selectOutboundProxy("", "")
	if proxyURL == "" {
		return &http.Client{Timeout: timeout}, nil
	}
	return newProxyHTTPClient(proxyURL, timeout)
}

func newProxyHTTPClient(proxyURL string, timeout time.Duration) (*http.Client, error) {
	parsedProxyURL, err := url.Parse(proxyURL)
	if err != nil {
		return nil, err
	}

	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			Proxy: http.ProxyURL(parsedProxyURL),
		},
	}, nil
}
