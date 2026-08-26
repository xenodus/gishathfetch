package gateway

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"mtg-price-checker-sg/gateway/util"
	"mtg-price-checker-sg/pkg/logger"
)

type outboundAttempt struct {
	strategy string
	proxyURL string
	client   *http.Client
}

// DoOutboundGET performs a GET with direct and dedicated-proxy fallback.
// Client errors (4xx) advance to the next transport without retrying.
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

// DoOutboundRoundTrip performs an HTTP round trip with direct and dedicated-proxy
// fallback. Client errors (4xx) advance to the next transport without retrying.
func DoOutboundRoundTrip(
	ctx context.Context,
	opts OutboundRequestOptions,
	timeout time.Duration,
	buildReq func() (*http.Request, error),
) (*http.Response, error) {
	if ShouldUseBrowserTLSEmulation(opts) {
		ctx = ContextWithBrowserProfile(ctx, PickBrowserProfile())
	}

	attempts := buildOutboundGETAttempts(ctx, timeout, opts)
	var failures []string
	for i, attempt := range attempts {
		resp, failure, ok, err := doOutboundAttempt(ctx, attempt, opts, buildReq)
		if err != nil {
			closeOutboundAttemptClient(attempt)
			closeOutboundAttemptClients(attempts[i+1:])
			return nil, err
		}
		if ok {
			closeOutboundAttemptClients(attempts[i+1:])
			return resp, nil
		}
		closeOutboundAttemptClient(attempt)
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

	req, err := buildReq()
	if err != nil {
		return nil, "", false, err
	}
	if err := PrepareOutboundRequest(ctx, req, opts); err != nil {
		return nil, "", false, err
	}

	logger.From(ctx).InfoContext(ctx, "outbound request: trying", "proxy", proxyDesc, "url", outboundRequestURL(req))

	resp, err := attempt.client.Do(req)
	if err != nil {
		lastFailure := fmt.Sprintf("%s: %v", attempt.strategy, err)
		logger.From(ctx).WarnContext(ctx, "outbound request: failed", "proxy", proxyDesc, "err", err)
		return nil, lastFailure, false, nil
	}

	if isOutboundClientError(resp.StatusCode) {
		lastFailure := outboundStatusFailure(attempt.strategy, resp)
		logger.From(ctx).WarnContext(ctx, "outbound request: failed", "proxy", proxyDesc, "status", resp.StatusCode)
		return nil, lastFailure, false, nil
	}

	logger.From(ctx).InfoContext(ctx, "outbound request: succeeded",
		"proxy", proxyDesc,
		"status", resp.StatusCode,
		"url", outboundRequestURL(req),
	)
	return resp, "", true, nil
}

func isOutboundClientError(statusCode int) bool {
	return statusCode >= http.StatusBadRequest && statusCode < http.StatusInternalServerError
}

func buildOutboundGETAttempts(ctx context.Context, timeout time.Duration, opts OutboundRequestOptions) []outboundAttempt {
	profile := ResolveBrowserProfileForOutbound(ctx, opts)

	if opts.OnlyProxyURL != "" {
		client, err := newOutboundHTTPClient(opts.OnlyProxyURL, timeout, profile)
		if err != nil {
			return nil
		}
		return []outboundAttempt{{
			strategy: "ck-pricelist-proxy",
			proxyURL: opts.OnlyProxyURL,
			client:   client,
		}}
	}

	appendDirect := func(dst []outboundAttempt) []outboundAttempt {
		if opts.SkipDirect {
			return dst
		}
		client, err := newOutboundHTTPClient("", timeout, profile)
		if err != nil {
			return dst
		}
		return append(dst, outboundAttempt{
			strategy: "direct",
			client:   client,
		})
	}

	appendResidential := func(dst []outboundAttempt) []outboundAttempt {
		if !opts.PreferResidentialProxy {
			return dst
		}
		proxyURL, ok := util.GetResidentialProxyURL()
		if !ok {
			return dst
		}
		client, err := newOutboundHTTPClient(proxyURL, timeout, profile)
		if err != nil {
			return dst
		}
		return append(dst, outboundAttempt{
			strategy: "residential-1",
			proxyURL: proxyURL,
			client:   client,
		})
	}

	// Match colly's selectOutboundProxy policy: one dedicated proxy per store search.
	// When searchShop pins a request-scoped lease, reuse that URL instead of
	// picking a new random slot for each outbound store.
	appendDedicated := func(dst []outboundAttempt) []outboundAttempt {
		if !DedicatedProxiesEnabled() {
			return dst
		}
		proxyURL, ok := dedicatedProxyURLForOutbound(ctx)
		if !ok {
			return dst
		}
		client, err := newOutboundHTTPClient(proxyURL, timeout, profile)
		if err != nil {
			return dst
		}
		return append(dst, outboundAttempt{
			strategy: dedicatedProxyStrategyName(proxyURL),
			proxyURL: proxyURL,
			client:   client,
		})
	}

	var attempts []outboundAttempt
	if opts.PreferDedicatedFirst {
		attempts = appendDedicated(attempts)
		attempts = appendDirect(attempts)
		attempts = appendResidential(attempts)
		return attempts
	}

	attempts = appendDirect(attempts)
	attempts = appendResidential(attempts)
	attempts = appendDedicated(attempts)
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
	if resp.StatusCode == http.StatusTooManyRequests {
		return msg
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
	case strings.HasPrefix(strategy, "residential-"):
		return "residential"
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

func closeOutboundAttemptClient(attempt outboundAttempt) {
	if attempt.proxyURL == "" || attempt.client == nil {
		return
	}
	closeProxyOutboundClient(attempt.client)
}

func closeOutboundAttemptClients(attempts []outboundAttempt) {
	for _, attempt := range attempts {
		closeOutboundAttemptClient(attempt)
	}
}

// NewOutboundHTTPClient returns an HTTP client that routes through a random dedicated
// proxy when configured, otherwise direct. The policy matches optimized colly
// collectors used by non-BinderPOS scrapers.
func NewOutboundHTTPClient(timeout time.Duration) (*http.Client, error) {
	_, proxyURL := selectOutboundProxy("", "")
	profile := BrowserEmulationProfile{}
	if ShouldUseBrowserTLSEmulationForScraping() {
		profile = PickBrowserProfile()
	}
	return newOutboundHTTPClient(proxyURL, timeout, profile)
}

func newProxyHTTPClient(proxyURL string, timeout time.Duration) (*http.Client, error) {
	profile := BrowserEmulationProfile{}
	if ShouldUseBrowserTLSEmulationForScraping() {
		profile = PickBrowserProfile()
	}
	return newOutboundHTTPClient(proxyURL, timeout, profile)
}
