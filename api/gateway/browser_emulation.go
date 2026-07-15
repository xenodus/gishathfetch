package gateway

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	"mtg-price-checker-sg/pkg/config"

	fhttp "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
)

type browserEmulationContextKey struct{}

// ContextWithBrowserProfile stores the browser emulation profile on ctx for the
// lifetime of one outbound round trip or scrape attempt.
func ContextWithBrowserProfile(ctx context.Context, profile BrowserEmulationProfile) context.Context {
	if !profile.Enabled {
		return ctx
	}
	return context.WithValue(ctx, browserEmulationContextKey{}, profile)
}

// BrowserProfileFromContext returns the browser profile pinned on ctx.
func BrowserProfileFromContext(ctx context.Context) (BrowserEmulationProfile, bool) {
	profile, ok := ctx.Value(browserEmulationContextKey{}).(BrowserEmulationProfile)
	return profile, ok && profile.Enabled
}

// ShouldUseBrowserTLSEmulation reports whether outbound scraping should use a
// browser-like TLS fingerprint instead of Go's default client hello.
func ShouldUseBrowserTLSEmulation(opts OutboundRequestOptions) bool {
	if !config.BrowserTLSEmulationEnabled() {
		return false
	}
	if WebBotAuthEnabled() && !opts.SkipWebBotAuth {
		return false
	}
	return true
}

// ShouldUseBrowserTLSEmulationForScraping reports whether colly/HTML scrapes should
// emulate browser TLS fingerprints.
func ShouldUseBrowserTLSEmulationForScraping() bool {
	if !config.BrowserTLSEmulationEnabled() {
		return false
	}
	return !WebBotAuthEnabled()
}

// ResolveBrowserProfileForOutbound returns the profile to use for an outbound
// request, preferring a value already stored on ctx.
func ResolveBrowserProfileForOutbound(ctx context.Context, opts OutboundRequestOptions) BrowserEmulationProfile {
	if !ShouldUseBrowserTLSEmulation(opts) {
		return BrowserEmulationProfile{}
	}
	if profile, ok := BrowserProfileFromContext(ctx); ok {
		return profile
	}
	return PickBrowserProfile()
}

// ResolveBrowserProfileForScraping returns the profile to use for HTML scraping.
func ResolveBrowserProfileForScraping(ctx context.Context) BrowserEmulationProfile {
	if !ShouldUseBrowserTLSEmulationForScraping() {
		return BrowserEmulationProfile{}
	}
	if profile, ok := BrowserProfileFromContext(ctx); ok {
		return profile
	}
	return PickBrowserProfile()
}

type browserTLSRoundTripper struct {
	inner tls_client.HttpClient
}

func (t *browserTLSRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	fReq, err := toFHTTPRequest(req)
	if err != nil {
		return nil, err
	}
	fResp, err := t.inner.Do(fReq)
	if err != nil {
		return nil, err
	}
	return fromFHTTPResponse(fResp), nil
}

func toFHTTPRequest(req *http.Request) (*fhttp.Request, error) {
	if req == nil {
		return nil, nil
	}

	fReq := &fhttp.Request{
		Method:           req.Method,
		URL:              req.URL,
		Proto:            req.Proto,
		ProtoMajor:       req.ProtoMajor,
		ProtoMinor:       req.ProtoMinor,
		Header:           cloneFHTTPHeader(req.Header),
		Body:             req.Body,
		GetBody:          req.GetBody,
		ContentLength:    req.ContentLength,
		TransferEncoding: append([]string(nil), req.TransferEncoding...),
		Close:            req.Close,
		Host:             req.Host,
		Form:             req.Form,
		PostForm:         req.PostForm,
		MultipartForm:    req.MultipartForm,
		Trailer:          cloneFHTTPHeader(req.Trailer),
		RemoteAddr:       req.RemoteAddr,
		RequestURI:       req.RequestURI,
	}
	if req.Context() != nil {
		fReq = fReq.WithContext(req.Context())
	}
	return fReq, nil
}

func fromFHTTPResponse(resp *fhttp.Response) *http.Response {
	if resp == nil {
		return nil
	}

	return &http.Response{
		Status:           resp.Status,
		StatusCode:       resp.StatusCode,
		Proto:            resp.Proto,
		ProtoMajor:       resp.ProtoMajor,
		ProtoMinor:       resp.ProtoMinor,
		Header:           cloneHTTPHeaderFromFHTTP(resp.Header),
		Body:             resp.Body,
		ContentLength:    resp.ContentLength,
		TransferEncoding: append([]string(nil), resp.TransferEncoding...),
		Close:            resp.Close,
		Uncompressed:     resp.Uncompressed,
		Trailer:          cloneHTTPHeaderFromFHTTP(resp.Trailer),
		Request:          fromFHTTPRequest(resp.Request),
	}
}

func fromFHTTPRequest(req *fhttp.Request) *http.Request {
	if req == nil {
		return nil
	}

	return &http.Request{
		Method:           req.Method,
		URL:              req.URL,
		Proto:            req.Proto,
		ProtoMajor:       req.ProtoMajor,
		ProtoMinor:       req.ProtoMinor,
		Header:           cloneHTTPHeaderFromFHTTP(req.Header),
		Body:             req.Body,
		GetBody:          req.GetBody,
		ContentLength:    req.ContentLength,
		TransferEncoding: append([]string(nil), req.TransferEncoding...),
		Close:            req.Close,
		Host:             req.Host,
		Form:             req.Form,
		PostForm:         req.PostForm,
		MultipartForm:    req.MultipartForm,
		Trailer:          cloneHTTPHeaderFromFHTTP(req.Trailer),
		RemoteAddr:       req.RemoteAddr,
		RequestURI:       req.RequestURI,
	}
}

func cloneFHTTPHeader(h http.Header) fhttp.Header {
	if h == nil {
		return nil
	}
	cloned := make(fhttp.Header, len(h))
	for key, values := range h {
		cloned[key] = append([]string(nil), values...)
	}
	return cloned
}

func cloneHTTPHeaderFromFHTTP(h fhttp.Header) http.Header {
	if h == nil {
		return nil
	}
	cloned := make(http.Header, len(h))
	for key, values := range h {
		cloned[key] = append([]string(nil), values...)
	}
	return cloned
}

// IsBrowserEmulatedTransport reports whether rt was created by newBrowserEmulatedHTTPClient.
func IsBrowserEmulatedTransport(rt http.RoundTripper) bool {
	_, ok := rt.(*browserTLSRoundTripper)
	return ok
}

// newOutboundHTTPClient returns an HTTP client for outbound scraping. When profile
// is enabled the client uses a browser-matched TLS fingerprint and HTTP/2 settings.
func newOutboundHTTPClient(proxyURL string, timeout time.Duration, profile BrowserEmulationProfile) (*http.Client, error) {
	if profile.Enabled {
		return newBrowserEmulatedHTTPClient(proxyURL, timeout, profile)
	}
	return newStdlibHTTPClient(proxyURL, timeout), nil
}

func newStdlibHTTPClient(proxyURL string, timeout time.Duration) *http.Client {
	proxyURL = strings.TrimSpace(proxyURL)
	if proxyURL == "" {
		return &http.Client{Timeout: timeout}
	}

	parsedProxyURL, err := url.Parse(proxyURL)
	if err != nil {
		return &http.Client{Timeout: timeout}
	}

	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			Proxy: http.ProxyURL(parsedProxyURL),
		},
	}
}

func newBrowserEmulatedHTTPClient(proxyURL string, timeout time.Duration, profile BrowserEmulationProfile) (*http.Client, error) {
	timeoutSeconds := max(int(timeout.Seconds()), 1)

	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(timeoutSeconds),
		tls_client.WithClientProfile(profile.TLSProfile),
		tls_client.WithRandomTLSExtensionOrder(),
	}
	if proxyURL = strings.TrimSpace(proxyURL); proxyURL != "" {
		options = append(options, tls_client.WithProxyUrl(proxyURL))
	}

	inner, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		return nil, err
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: &browserTLSRoundTripper{inner: inner},
	}, nil
}

// NewBinderposHTTPClient returns an HTTP client for BinderPOS storefront API calls.
func NewBinderposHTTPClient(proxyURL string, profile BrowserEmulationProfile) (*http.Client, error) {
	return newOutboundHTTPClient(proxyURL, config.SearchAttemptTimeout, profile)
}
