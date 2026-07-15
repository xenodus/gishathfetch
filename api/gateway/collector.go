package gateway

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"mtg-price-checker-sg/gateway/util"
	"mtg-price-checker-sg/pkg/config"
	"os"
	"strings"
	"sync"

	"github.com/gocolly/colly/v2"
)

var browserUserAgents = func() []string {
	agents := make([]string, 0, len(browserEmulationProfiles))
	for _, profile := range browserEmulationProfiles {
		agents = append(agents, profile.UserAgent)
	}
	return agents
}()

var dedicatedProxyLeases = newDedicatedProxyLeasePool()

type dedicatedProxyLeasePool struct {
	mu    sync.Mutex
	cond  *sync.Cond
	inUse map[string]bool
}

func newDedicatedProxyLeasePool() *dedicatedProxyLeasePool {
	pool := &dedicatedProxyLeasePool{
		inUse: make(map[string]bool),
	}
	pool.cond = sync.NewCond(&pool.mu)
	return pool
}

func (p *dedicatedProxyLeasePool) acquire(proxyURLs []string) (string, bool) {
	if len(proxyURLs) == 0 {
		return "", false
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	for {
		if proxyURL, ok := p.tryAcquireLocked(proxyURLs); ok {
			return proxyURL, true
		}
		p.cond.Wait()
	}
}

// tryAcquireLocked picks a free dedicated proxy in random order among proxyURLs.
// Caller must hold p.mu.
func (p *dedicatedProxyLeasePool) tryAcquireLocked(proxyURLs []string) (string, bool) {
	for _, idx := range rand.Perm(len(proxyURLs)) {
		proxyURL := proxyURLs[idx]
		if proxyURL == "" {
			continue
		}
		if !p.inUse[proxyURL] {
			p.inUse[proxyURL] = true
			return proxyURL, true
		}
	}
	return "", false
}

// acquireCtx is like acquire but returns when ctx is cancelled or timed out if no slot is available.
func (p *dedicatedProxyLeasePool) acquireCtx(ctx context.Context, proxyURLs []string) (string, error) {
	if len(proxyURLs) == 0 {
		return "", errors.New("no dedicated proxy URLs")
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}

	stop := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			p.mu.Lock()
			p.cond.Broadcast()
			p.mu.Unlock()
		case <-stop:
		}
	}()
	defer close(stop)

	p.mu.Lock()
	defer p.mu.Unlock()

	for {
		if proxyURL, ok := p.tryAcquireLocked(proxyURLs); ok {
			return proxyURL, nil
		}
		if err := ctx.Err(); err != nil {
			return "", err
		}
		p.cond.Wait()
		if err := ctx.Err(); err != nil {
			return "", err
		}
	}
}

func (p *dedicatedProxyLeasePool) release(proxyURL string) {
	if proxyURL == "" {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.inUse[proxyURL] {
		delete(p.inUse, proxyURL)
		p.cond.Signal()
	}
}

// LeaseDedicatedProxyURL acquires one free dedicated proxy URL from proxyURLs, trying URLs in a
// random order until one is available. The caller must invoke release exactly once when finished.
// If ctx is cancelled or times out before a slot is available, err is non-nil and release is a no-op.
func LeaseDedicatedProxyURL(ctx context.Context, proxyURLs []string) (leasedURL string, release func(), err error) {
	if len(proxyURLs) == 0 {
		return "", func() {}, errors.New("no dedicated proxy URLs")
	}
	leasedURL, err = dedicatedProxyLeases.acquireCtx(ctx, proxyURLs)
	if err != nil {
		return "", func() {}, err
	}
	u := leasedURL
	return leasedURL, func() {
		dedicatedProxyLeases.release(u)
	}, nil
}

// Collector constructors and top-level configuration.
func NewOptimizedCollector(ctx context.Context) *colly.Collector {
	c := colly.NewCollector(
		colly.StdlibContext(ctx),
	)
	configureRequestOptimizations(c, ctx, false)
	return c
}

func NewOptimizedCollectorNoRetry(ctx context.Context) *colly.Collector {
	c := colly.NewCollector(
		colly.StdlibContext(ctx),
	)
	configureRequestOptimizations(c, ctx, false)
	return c
}

func NewOptimizedCollectorNoRetryDirect(ctx context.Context) *colly.Collector {
	c := colly.NewCollector(
		colly.StdlibContext(ctx),
	)
	configureRequestOptimizations(c, ctx, false)
	forceCollectorDirectProxy(c)
	return c
}

func NewOptimizedCollectorNoRetryDynamic(ctx context.Context) (*colly.Collector, error) {
	proxyURL := DynamicProxyURL()
	if proxyURL == "" {
		return nil, errors.New("no dynamic proxy configured")
	}

	c := colly.NewCollector(
		colly.StdlibContext(ctx),
	)
	configureRequestOptimizations(c, ctx, false)
	if err := forceCollectorProxy(c, "dynamic", proxyURL); err != nil {
		return nil, fmt.Errorf("invalid dynamic proxy configured: %w", err)
	}
	return c, nil
}

func NewOptimizedCollectorForBinderpos(ctx context.Context) *colly.Collector {
	c := colly.NewCollector(
		colly.StdlibContext(ctx),
	)
	configureRequestOptimizations(c, ctx, config.UseLeasedDedicatedProxy)
	return c
}

func ConfigureRequestOptimizations(c *colly.Collector) {
	configureRequestOptimizations(c, context.Background(), false)
}

func ConfigureRequestOptimizationsNoRetry(c *colly.Collector) {
	configureRequestOptimizations(c, context.Background(), false)
}

func forceCollectorDirectProxy(c *colly.Collector) {
	profile := PickBrowserProfile()
	if !ShouldUseBrowserTLSEmulationForScraping() {
		profile = BrowserEmulationProfile{}
	}
	client, err := newOutboundHTTPClient("", config.SearchAttemptTimeout, profile)
	if err == nil {
		c.SetClient(client)
	} else {
		c.SetProxyFunc(nil)
	}
	c.OnRequest(func(r *colly.Request) {
		if r == nil || r.Ctx == nil {
			return
		}
		r.Ctx.Put("last_proxy_mode", "direct")
		r.Ctx.Put("last_proxy_url", "")
	})
}

func forceCollectorProxy(c *colly.Collector, mode, proxyURL string) error {
	profile := PickBrowserProfile()
	if !ShouldUseBrowserTLSEmulationForScraping() {
		profile = BrowserEmulationProfile{}
	}
	client, err := newOutboundHTTPClient(proxyURL, config.SearchAttemptTimeout, profile)
	if err != nil {
		if setErr := c.SetProxy(proxyURL); setErr != nil {
			return setErr
		}
	} else {
		c.SetClient(client)
	}
	c.OnRequest(func(r *colly.Request) {
		if r == nil || r.Ctx == nil {
			return
		}
		r.Ctx.Put("last_proxy_mode", mode)
		r.Ctx.Put("last_proxy_url", proxyURL)
	})
	return nil
}

// Core request optimization pipeline. Gateway colly requests do not retry; each lookup is a single attempt.
func configureRequestOptimizations(c *colly.Collector, ctx context.Context, enforceDedicatedProxyLease bool) {
	applyCollectorDefaults(c)

	requestDedicatedProxyURL, _ := RequestDedicatedProxyURL(ctx)
	leasedDedicatedProxyURL, releaseDedicatedProxy := leaseDedicatedProxyIfNeeded(enforceDedicatedProxyLease, requestDedicatedProxyURL)
	profile := ResolveBrowserProfileForScraping(ctx)
	ctx = ContextWithBrowserProfile(ctx, profile)
	applyCollectorHTTPClient(c, leasedDedicatedProxyURL, requestDedicatedProxyURL, profile)
	registerRequestHandler(c, leasedDedicatedProxyURL, requestDedicatedProxyURL, profile)
	registerScrapedHandler(c, releaseDedicatedProxy)
	registerNoRetryErrorHandler(c, releaseDedicatedProxy)
}

func applyCollectorHTTPClient(
	c *colly.Collector,
	leasedDedicatedProxyURL, requestDedicatedProxyURL string,
	profile BrowserEmulationProfile,
) {
	mode, proxyURL := selectOutboundProxy(leasedDedicatedProxyURL, requestDedicatedProxyURL)
	client, err := newOutboundHTTPClient(proxyURL, config.SearchAttemptTimeout, profile)
	if err != nil {
		log.Printf("browser TLS client setup failed for colly (%s): %v", mode, err)
		if proxyURL != "" {
			if setErr := c.SetProxy(proxyURL); setErr != nil {
				log.Printf("invalid proxy configured for colly (%s): %v", mode, setErr)
			}
		}
		return
	}
	c.SetClient(client)
}

// Base collector defaults used by all optimized collectors.
func applyCollectorDefaults(c *colly.Collector) {
	c.DisableCookies()
	c.SetRequestTimeout(config.SearchAttemptTimeout)
}

// Dedicated proxy lease lifecycle helpers.
func leaseDedicatedProxyIfNeeded(enforceDedicatedProxyLease bool, requestDedicatedProxyURL string) (string, func()) {
	var leasedDedicatedProxyURL string
	releaseLeasedDedicatedProxy := func() {}
	if requestDedicatedProxyURL != "" {
		return "", releaseLeasedDedicatedProxy
	}
	if enforceDedicatedProxyLease && config.UseProxy && config.UseLeasedDedicatedProxy {
		if proxyURL, ok := dedicatedProxyLeases.acquire(util.GetDedicatedProxyURLs()); ok {
			leasedDedicatedProxyURL = proxyURL
			releaseLeasedDedicatedProxy = func() {
				dedicatedProxyLeases.release(proxyURL)
			}
		}
	}
	var releaseOnce sync.Once
	releaseDedicatedProxy := func() {
		releaseOnce.Do(releaseLeasedDedicatedProxy)
	}

	return leasedDedicatedProxyURL, releaseDedicatedProxy
}

// selectOutboundProxy picks the proxy mode and URL for a single outbound attempt.
// When requestDedicatedProxyURL is set the store search holds one dedicated lease;
// otherwise a per-collector lease or random dedicated proxy is chosen when
// configured, then dynamic proxy, then direct.
func selectOutboundProxy(leasedDedicatedProxyURL, requestDedicatedProxyURL string) (mode string, proxyURL string) {
	if !config.UseProxy {
		return "direct", ""
	}
	if requestDedicatedProxyURL != "" {
		return "dedicated", requestDedicatedProxyURL
	}
	if leasedDedicatedProxyURL != "" {
		return "dedicated", leasedDedicatedProxyURL
	}
	if proxyURL, ok := randomDedicatedProxyURL(""); ok {
		return "dedicated", proxyURL
	}
	if proxyURL := DynamicProxyURL(); proxyURL != "" {
		return "dynamic", proxyURL
	}
	return "direct", ""
}

// applyInitialProxy configures the transport for the first and only request for this collector, mirroring
// the prior "initial attempt" proxy policy without any follow-up attempts.
func applyInitialProxy(c *colly.Collector, leasedDedicatedProxyURL, requestDedicatedProxyURL string) (string, string) {
	mode, proxyURL := selectOutboundProxy(leasedDedicatedProxyURL, requestDedicatedProxyURL)
	if proxyURL == "" {
		return clearProxy(c)
	}
	if err := c.SetProxy(proxyURL); err != nil {
		if mode == "dynamic" {
			log.Printf("invalid dynamic proxy configured: %v", err)
		}
		return clearProxy(c)
	}
	return mode, proxyURL
}

// Colly callback registration helpers.
func registerRequestHandler(
	c *colly.Collector,
	leasedDedicatedProxyURL, requestDedicatedProxyURL string,
	profile BrowserEmulationProfile,
) {
	initialProxyMode, initialProxyURL := selectOutboundProxy(leasedDedicatedProxyURL, requestDedicatedProxyURL)
	c.OnRequest(func(r *colly.Request) {
		if r == nil || r.URL == nil {
			return
		}
		if err := waitForDomainRequestSlot(c.Context, r.URL); err != nil {
			log.Printf("Skipping request pacing for %s due to context cancellation: %v", r.URL, err)
			r.Abort()
			return
		}
		if r.Headers != nil {
			ApplyBrowserLikeHTMLHeaders(r.Headers, r.URL)
			// Keep gzip only. Go's default client does not transparently decode brotli ("br").
			r.Headers.Set("Accept-Encoding", "gzip")
			if profile.Enabled {
				r.Headers.Set("User-Agent", profile.UserAgent)
				ApplyBrowserProfileHeaders(r.Headers, profile)
			} else {
				r.Headers.Set("User-Agent", OutboundUserAgent())
				if err := SignWebBotAuthCollyRequest(r); err != nil {
					log.Printf("Web Bot Auth signing failed for %s: %v", r.URL, err)
				}
			}
		}
		seedProxyContextIfMissing(r.Ctx, initialProxyMode, initialProxyURL)
	})
}

func seedProxyContextIfMissing(ctx *colly.Context, initialProxyMode, initialProxyURL string) {
	if ctx == nil {
		return
	}

	// Proxy context is considered initialized once mode exists.
	// Direct mode intentionally uses an empty proxy URL.
	if ctx.Get("last_proxy_mode") != "" {
		return
	}

	ctx.Put("last_proxy_mode", initialProxyMode)
	ctx.Put("last_proxy_url", initialProxyURL)
}

func registerScrapedHandler(c *colly.Collector, releaseDedicatedProxy func()) {
	c.OnScraped(func(_ *colly.Response) {
		releaseDedicatedProxy()
	})
}

func registerNoRetryErrorHandler(c *colly.Collector, releaseDedicatedProxy func()) {
	c.OnError(func(_ *colly.Response, _ error) {
		releaseDedicatedProxy()
	})
}

// VisitWithProxyInfo wraps collector visit errors with the selected proxy context.
func VisitWithProxyInfo(c *colly.Collector, targetURL string) error {
	var proxyMu sync.Mutex
	var lastProxyMode string
	var lastProxyURL string
	var responseStatusCode int

	c.OnRequest(func(r *colly.Request) {
		if r == nil || r.Ctx == nil {
			return
		}

		proxyMu.Lock()
		lastProxyMode = r.Ctx.Get("last_proxy_mode")
		lastProxyURL = r.Ctx.Get("last_proxy_url")
		proxyMu.Unlock()
	})

	c.OnError(func(r *colly.Response, _ error) {
		if r == nil || r.StatusCode < 100 {
			return
		}

		proxyMu.Lock()
		responseStatusCode = r.StatusCode
		proxyMu.Unlock()
	})

	err := c.Visit(targetURL)
	if err == nil {
		return nil
	}

	proxyMu.Lock()
	mode := lastProxyMode
	proxyURL := lastProxyURL
	statusCode := responseStatusCode
	proxyMu.Unlock()

	err = EnrichErrorWithHTTPStatus(err, statusCode)
	return fmt.Errorf("%w (%s)", err, formatProxyContext(mode, proxyURL))
}

// Proxy strategy helpers.
func clearProxy(c *colly.Collector) (string, string) {
	c.SetProxyFunc(nil)
	return "direct", ""
}

func DynamicProxyURL() string {
	if !config.UseDynamicProxy() {
		return ""
	}

	proxyURL, ok := util.BuildProxyURL(os.Getenv(config.DynamicProxyEnv))
	if !ok {
		return ""
	}
	return strings.TrimSpace(proxyURL)
}

// RandomDedicatedProxyURL picks one configured dedicated proxy uniformly at random.
func RandomDedicatedProxyURL() (string, bool) {
	return randomDedicatedProxyURL("")
}

func randomDedicatedProxyURL(avoidProxyURL string) (string, bool) {
	dedicatedProxies := util.GetDedicatedProxy()
	valid := make([]util.DedicatedProxy, 0, len(dedicatedProxies))
	for _, proxy := range dedicatedProxies {
		if proxy.Host == "" || proxy.Port == "" {
			continue
		}
		valid = append(valid, proxy)
	}
	if len(valid) == 0 {
		return "", false
	}

	candidates := valid
	if avoidProxyURL != "" && len(valid) > 1 {
		filtered := make([]util.DedicatedProxy, 0, len(valid))
		for _, p := range valid {
			proxyURL, ok := util.BuildDedicatedProxyURL(p)
			if !ok {
				continue
			}
			if proxyURL != avoidProxyURL {
				filtered = append(filtered, p)
			}
		}
		if len(filtered) > 0 {
			candidates = filtered
		}
	}

	p := candidates[rand.IntN(len(candidates))]
	proxyURL, ok := util.BuildDedicatedProxyURL(p)
	return proxyURL, ok
}

func formatProxyContext(mode, proxyURL string) string {
	if mode == "" {
		mode = "unknown"
	}

	return fmt.Sprintf("proxy_mode=%s proxy=%s", mode, resolveProxyLabel(mode, proxyURL))
}

func resolveProxyLabel(mode, proxyURL string) string {
	if proxyURL == "" {
		return "none"
	}

	switch mode {
	case "direct":
		return "none"
	case "dynamic":
		if dynamicProxyURL := DynamicProxyURL(); dynamicProxyURL != "" && dynamicProxyURL == proxyURL {
			return config.DynamicProxyEnv
		}
		return "dynamic-configured"
	case "dedicated":
		if label := dedicatedProxyEnvLabel(proxyURL); label != "" {
			return label
		}
		return "dedicated-configured"
	case "residential":
		if configuredURL, ok := util.GetResidentialProxyURL(); ok && configuredURL == proxyURL {
			return config.ResidentialProxyEnv
		}
		return "residential-configured"
	case "ck-pricelist":
		if configuredURL, ok := util.GetCKPricelistProxyURL(); ok && configuredURL == proxyURL {
			return config.CKPricelistProxyEnv
		}
		return "ck-pricelist-configured"
	default:
		return "configured"
	}
}

func dedicatedProxyEnvLabel(proxyURL string) string {
	dedicatedProxies := util.GetDedicatedProxy()
	for idx, proxy := range dedicatedProxies {
		candidateURL, ok := util.BuildDedicatedProxyURL(proxy)
		if !ok {
			continue
		}
		if candidateURL == proxyURL {
			return fmt.Sprintf("DEDICATED_PROXY_%d", idx+1)
		}
	}

	return ""
}
