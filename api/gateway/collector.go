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

var browserUserAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 14_4_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:137.0) Gecko/20100101 Firefox/137.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 14.4; rv:137.0) Gecko/20100101 Firefox/137.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 14_4_0) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.4 Safari/605.1.15",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Edg/135.0.3179.98 Chrome/135.0.0.0 Safari/537.36",
}

func RandomBrowserUserAgent() string {
	if len(browserUserAgents) == 0 {
		return "Mozilla/5.0"
	}
	return browserUserAgents[rand.IntN(len(browserUserAgents))]
}

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
	configureRequestOptimizations(c, false)
	return c
}

func NewOptimizedCollectorNoRetry(ctx context.Context) *colly.Collector {
	c := colly.NewCollector(
		colly.StdlibContext(ctx),
	)
	configureRequestOptimizations(c, false)
	return c
}

func NewOptimizedCollectorNoRetryDirect(ctx context.Context) *colly.Collector {
	c := colly.NewCollector(
		colly.StdlibContext(ctx),
	)
	configureRequestOptimizations(c, false)
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
	configureRequestOptimizations(c, false)
	if err := forceCollectorProxy(c, "dynamic", proxyURL); err != nil {
		return nil, fmt.Errorf("invalid dynamic proxy configured: %w", err)
	}
	return c, nil
}

func NewOptimizedCollectorForBinderpos(ctx context.Context) *colly.Collector {
	c := colly.NewCollector(
		colly.StdlibContext(ctx),
	)
	configureRequestOptimizations(c, config.UseLeasedDedicatedProxy)
	return c
}

func ConfigureRequestOptimizations(c *colly.Collector) {
	configureRequestOptimizations(c, false)
}

func ConfigureRequestOptimizationsNoRetry(c *colly.Collector) {
	configureRequestOptimizations(c, false)
}

func forceCollectorDirectProxy(c *colly.Collector) {
	c.SetProxyFunc(nil)
	c.OnRequest(func(r *colly.Request) {
		if r == nil || r.Ctx == nil {
			return
		}
		r.Ctx.Put("last_proxy_mode", "direct")
		r.Ctx.Put("last_proxy_url", "")
	})
}

func forceCollectorProxy(c *colly.Collector, mode, proxyURL string) error {
	if err := c.SetProxy(proxyURL); err != nil {
		return err
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
func configureRequestOptimizations(c *colly.Collector, enforceDedicatedProxyLease bool) {
	applyCollectorDefaults(c)

	leasedDedicatedProxyURL, releaseDedicatedProxy := leaseDedicatedProxyIfNeeded(enforceDedicatedProxyLease)
	registerRequestHandler(c, leasedDedicatedProxyURL)
	registerScrapedHandler(c, releaseDedicatedProxy)
	registerNoRetryErrorHandler(c, releaseDedicatedProxy)
}

// Base collector defaults used by all optimized collectors.
func applyCollectorDefaults(c *colly.Collector) {
	c.DisableCookies()
	c.SetRequestTimeout(config.SearchAttemptTimeout)
}

// Dedicated proxy lease lifecycle helpers.
func leaseDedicatedProxyIfNeeded(enforceDedicatedProxyLease bool) (string, func()) {
	var leasedDedicatedProxyURL string
	releaseLeasedDedicatedProxy := func() {}
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

// applyInitialProxy configures the transport for the first and only request for this collector, mirroring
// the prior "initial attempt" proxy policy without any follow-up attempts.
func applyInitialProxy(c *colly.Collector, leasedDedicatedProxyURL string) (string, string) {
	if !config.UseProxy {
		return clearProxy(c)
	}
	if leasedDedicatedProxyURL != "" {
		c.SetProxy(leasedDedicatedProxyURL)
		return "dedicated", leasedDedicatedProxyURL
	}
	if proxyURL, ok := randomDedicatedProxyURL(""); ok {
		c.SetProxy(proxyURL)
		return "dedicated", proxyURL
	}
	if proxyURL := DynamicProxyURL(); proxyURL != "" {
		if err := c.SetProxy(proxyURL); err == nil {
			return "dynamic", proxyURL
		} else {
			log.Printf("invalid dynamic proxy configured: %v", err)
		}
	}
	return clearProxy(c)
}

// Colly callback registration helpers.
func registerRequestHandler(c *colly.Collector, leasedDedicatedProxyURL string) {
	initialProxyMode, initialProxyURL := applyInitialProxy(c, leasedDedicatedProxyURL)
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
			r.Headers.Set("User-Agent", OutboundUserAgent())
			if err := SignWebBotAuthCollyRequest(r); err != nil {
				log.Printf("Web Bot Auth signing failed for %s: %v", r.URL, err)
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
