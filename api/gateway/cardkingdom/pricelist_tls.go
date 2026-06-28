package cardkingdom

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	fhttp "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/util"
)

const ckTLSUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"

type ckTLSAttempt struct {
	strategy string
	proxyURL string
}

func fetchPricelistBodyTLS(ctx context.Context) ([]byte, error) {
	requestURL, err := url.Parse(pricelistURL)
	if err != nil {
		return nil, err
	}
	if err := gateway.WaitForDomainRequestSlot(ctx, requestURL); err != nil {
		return nil, err
	}

	var failures []string
	for _, attempt := range buildCKTLSAttempts() {
		body, status, detail, err := doCKTLSGET(ctx, attempt.proxyURL)
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", attempt.strategy, err))
			continue
		}
		if status == fhttp.StatusOK && !looksLikeCloudflareChallenge(body) {
			return body, nil
		}
		failures = append(failures, fmt.Sprintf("%s: status %d%s", attempt.strategy, status, detail))
	}
	if len(failures) == 0 {
		return nil, fmt.Errorf("tls get failed")
	}
	return nil, fmt.Errorf("tls get failed: %s", strings.Join(failures, "; "))
}

func buildCKTLSAttempts() []ckTLSAttempt {
	var attempts []ckTLSAttempt
	for idx, proxyURL := range util.GetDedicatedProxyURLs() {
		if proxyURL == "" {
			continue
		}
		attempts = append(attempts, ckTLSAttempt{
			strategy: fmt.Sprintf("dedicated-%d", idx+1),
			proxyURL: proxyURL,
		})
	}
	if proxyURL := gateway.DynamicProxyURL(); proxyURL != "" {
		attempts = append(attempts, ckTLSAttempt{strategy: "dynamic", proxyURL: proxyURL})
	}
	attempts = append(attempts, ckTLSAttempt{strategy: "direct", proxyURL: ""})
	return attempts
}

func doCKTLSGET(ctx context.Context, proxyURL string) ([]byte, int, string, error) {
	client, err := newCKTLSClient(proxyURL, pricelistTimeout)
	if err != nil {
		return nil, 0, "", err
	}

	req, err := fhttp.NewRequestWithContext(ctx, fhttp.MethodGet, pricelistURL, nil)
	if err != nil {
		return nil, 0, "", err
	}
	req.Header = fhttp.Header{
		"accept":          {"application/json"},
		"accept-language": {"en-US,en;q=0.9"},
		"user-agent":      {ckTLSUserAgent},
		fhttp.HeaderOrderKey: {
			"accept",
			"accept-language",
			"user-agent",
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, "", err
	}
	return body, resp.StatusCode, tlsFailureDetail(resp, body), nil
}

func newCKTLSClient(proxyURL string, timeout time.Duration) (tls_client.HttpClient, error) {
	timeoutSeconds := int(timeout.Seconds())
	if timeoutSeconds < 1 {
		timeoutSeconds = 1
	}

	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(timeoutSeconds),
		tls_client.WithClientProfile(profiles.Chrome_131),
		tls_client.WithCookieJar(tls_client.NewCookieJar()),
	}
	if proxyURL != "" {
		options = append(options, tls_client.WithProxyUrl(proxyURL))
	}
	return tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
}

func tlsFailureDetail(resp *fhttp.Response, body []byte) string {
	if resp == nil {
		return ""
	}

	var parts []string
	if cfRay := resp.Header.Get("cf-ray"); cfRay != "" {
		parts = append(parts, "cf-ray="+cfRay)
	}
	if looksLikeCloudflareChallenge(body) {
		parts = append(parts, "cloudflare-challenge")
		return formatTLSFailureParts(parts)
	}

	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return formatTLSFailureParts(parts)
	}
	if len(trimmed) > 120 {
		trimmed = trimmed[:120] + "..."
	}
	parts = append(parts, "("+trimmed+")")
	return formatTLSFailureParts(parts)
}

func formatTLSFailureParts(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	return " " + strings.Join(parts, " ")
}

func looksLikeCloudflareChallenge(body []byte) bool {
	prefix := string(body)
	if len(prefix) > 512 {
		prefix = prefix[:512]
	}
	lower := strings.ToLower(prefix)
	return strings.Contains(lower, "just a moment") || strings.Contains(lower, "cloudflare")
}
