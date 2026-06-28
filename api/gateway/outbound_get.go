package gateway

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"

	"mtg-price-checker-sg/gateway/util"
)

var outboundDedicatedProxySeq atomic.Uint32

type outboundAttempt struct {
	strategy string
	client   *http.Client
}

// DoOutboundGET performs a GET with direct, dedicated-proxy, and dynamic-proxy
// fallback. Transient 403/429 responses advance to the next transport.
func DoOutboundGET(
	ctx context.Context,
	requestURL string,
	opts OutboundRequestOptions,
	timeout time.Duration,
) (*http.Response, error) {
	var lastErr error
	for _, attempt := range buildOutboundGETAttempts(timeout) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
		if err != nil {
			return nil, err
		}
		if err := PrepareOutboundRequest(ctx, req, opts); err != nil {
			return nil, err
		}

		resp, err := attempt.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("%s: %w", attempt.strategy, err)
			continue
		}
		if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests {
			lastErr = fmt.Errorf("%s: status %d", attempt.strategy, resp.StatusCode)
			resp.Body.Close()
			continue
		}
		return resp, nil
	}
	if lastErr == nil {
		return nil, fmt.Errorf("outbound get failed")
	}
	return nil, lastErr
}

func buildOutboundGETAttempts(timeout time.Duration) []outboundAttempt {
	attempts := []outboundAttempt{{
		strategy: "direct",
		client:   &http.Client{Timeout: timeout},
	}}

	if client, ok := dedicatedProxyHTTPClient(timeout); ok {
		attempts = append(attempts, outboundAttempt{strategy: "dedicated", client: client})
	}

	if client, ok := dynamicProxyHTTPClient(timeout); ok {
		attempts = append(attempts, outboundAttempt{strategy: "dynamic", client: client})
	}

	return attempts
}

func dedicatedProxyHTTPClient(timeout time.Duration) (*http.Client, bool) {
	proxyURLs := util.GetDedicatedProxyURLs()
	if len(proxyURLs) == 0 {
		return nil, false
	}

	client, err := newProxyHTTPClient(nextOutboundDedicatedProxyURL(proxyURLs), timeout)
	if err != nil {
		return nil, false
	}
	return client, true
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

func nextOutboundDedicatedProxyURL(proxyURLs []string) string {
	v := outboundDedicatedProxySeq.Add(1) - 1
	return proxyURLs[int(v%uint32(len(proxyURLs)))]
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
