package gateway

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"mtg-price-checker-sg/gateway/util"
)

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
	var failures []string
	for _, attempt := range buildOutboundGETAttempts(timeout, opts.SkipDirect) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
		if err != nil {
			return nil, err
		}
		if err := PrepareOutboundRequest(ctx, req, opts); err != nil {
			return nil, err
		}

		resp, err := attempt.client.Do(req)
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", attempt.strategy, err))
			continue
		}
		if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests {
			failures = append(failures, outboundStatusFailure(attempt.strategy, resp))
			continue
		}
		return resp, nil
	}
	if len(failures) == 0 {
		return nil, fmt.Errorf("outbound get failed")
	}
	return nil, fmt.Errorf("outbound get failed: %s", strings.Join(failures, "; "))
}

func buildOutboundGETAttempts(timeout time.Duration, skipDirect bool) []outboundAttempt {
	var attempts []outboundAttempt
	if !skipDirect {
		attempts = append(attempts, outboundAttempt{
			strategy: "direct",
			client:   &http.Client{Timeout: timeout},
		})
	}

	for idx, proxyURL := range util.GetDedicatedProxyURLs() {
		if proxyURL == "" {
			continue
		}
		client, err := newProxyHTTPClient(proxyURL, timeout)
		if err != nil {
			continue
		}
		attempts = append(attempts, outboundAttempt{
			strategy: fmt.Sprintf("dedicated-%d", idx+1),
			client:   client,
		})
	}

	if client, ok := dynamicProxyHTTPClient(timeout); ok {
		attempts = append(attempts, outboundAttempt{strategy: "dynamic", client: client})
	}

	return attempts
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

// NewOutboundHTTPClient returns an HTTP client that routes through a random dedicated
// proxy when configured, otherwise dynamic proxy, otherwise direct. The policy matches
// optimized colly collectors used by non-BinderPOS scrapers.
func NewOutboundHTTPClient(timeout time.Duration) (*http.Client, error) {
	_, proxyURL := selectOutboundProxy("")
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
