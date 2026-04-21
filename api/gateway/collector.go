package gateway

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"mtg-price-checker-sg/gateway/util"
	"mtg-price-checker-sg/pkg/config"
	"os"
	"strconv"
	"sync"
	"time"

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

const (
	// Default flow allows two retries after the initial request:
	// - first retry (attempt 1): dedicated proxy
	// - second/final retry (attempt 2): direct (no proxy)
	defaultMaxRetries = 2
	// Binderpos also uses two retries after the initial request, but with a different first retry path:
	// - first retry (attempt 1): shared/dynamic PROXY_URL
	// - second/final retry (attempt 2): direct (no proxy)
	binderposMaxRetries = defaultMaxRetries
	// Dedicated (leased or random pool) is used while retryAttempt <= this value (initial request is attempt 0).
	// First retry = attempt 1 (still dedicated).
	dedicatedProxyRetryThreshold          = 1
	binderposDedicatedProxyRetryThreshold = 0
	// Keep a small reserve so a retry request can still be dispatched before context cancellation.
	retryExecutionBuffer = 300 * time.Millisecond
)

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
		p.cond.Wait()
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

// Collector constructors and top-level configuration.
func NewOptimizedCollector(ctx context.Context) *colly.Collector {
	c := colly.NewCollector(
		colly.StdlibContext(ctx),
	)
	configureRequestOptimizations(c, true, false)
	return c
}

func NewOptimizedCollectorNoRetry(ctx context.Context) *colly.Collector {
	c := colly.NewCollector(
		colly.StdlibContext(ctx),
	)
	configureRequestOptimizations(c, false, false)
	return c
}

func NewOptimizedCollectorForBinderpos(ctx context.Context) *colly.Collector {
	c := colly.NewCollector(
		colly.StdlibContext(ctx),
	)
	configureRequestOptimizationsWithDedicatedThreshold(c, true, true, binderposDedicatedProxyRetryThreshold, binderposMaxRetries)
	return c
}

func ConfigureRequestOptimizations(c *colly.Collector) {
	configureRequestOptimizations(c, true, false)
}

func ConfigureRequestOptimizationsNoRetry(c *colly.Collector) {
	configureRequestOptimizations(c, false, false)
}

// Core request optimization pipeline.
func configureRequestOptimizations(c *colly.Collector, enableRetry, enforceDedicatedProxyLease bool) {
	configureRequestOptimizationsWithDedicatedThreshold(c, enableRetry, enforceDedicatedProxyLease, dedicatedProxyRetryThreshold, defaultMaxRetries)
}

func configureRequestOptimizationsWithDedicatedThreshold(c *colly.Collector, enableRetry, enforceDedicatedProxyLease bool, dedicatedRetryThreshold, maxRetries int) {
	applyCollectorDefaults(c)

	leasedDedicatedProxyURL, releaseDedicatedProxy := leaseDedicatedProxyIfNeeded(enforceDedicatedProxyLease)
	registerRequestHandler(c, leasedDedicatedProxyURL, dedicatedRetryThreshold, maxRetries)
	registerScrapedHandler(c, releaseDedicatedProxy)

	if !enableRetry {
		registerNoRetryErrorHandler(c, releaseDedicatedProxy)
		return
	}

	registerRetryErrorHandler(c, leasedDedicatedProxyURL, releaseDedicatedProxy, dedicatedRetryThreshold, maxRetries)
}

// Base collector defaults used by all optimized collectors.
func applyCollectorDefaults(c *colly.Collector) {
	c.DisableCookies()
	c.SetRequestTimeout(config.PerSiteTimeout)
}

// Dedicated proxy lease lifecycle helpers.
func leaseDedicatedProxyIfNeeded(enforceDedicatedProxyLease bool) (string, func()) {
	var leasedDedicatedProxyURL string
	releaseLeasedDedicatedProxy := func() {}
	if enforceDedicatedProxyLease && config.UseProxy {
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

// Colly callback registration helpers.
func registerRequestHandler(c *colly.Collector, leasedDedicatedProxyURL string, dedicatedRetryThreshold, maxRetries int) {
	initialProxyMode, initialProxyURL := applyProxyForRetryAttemptWithPinnedDedicated(c, 0, "", leasedDedicatedProxyURL, dedicatedRetryThreshold, maxRetries)
	c.OnRequest(func(r *colly.Request) {
		if r == nil || r.URL == nil {
			return
		}
		if err := waitForDomainRequestSlot(c.Context, r.URL); err != nil {
			log.Printf("Skipping request pacing for %s due to context cancellation: %v", r.URL, err)
			r.Abort()
			return
		}
		// Keep gzip only. Go's default client does not transparently decode brotli ("br").
		r.Headers.Set("Accept-Encoding", "gzip")
		r.Headers.Set("User-Agent", RandomBrowserUserAgent())
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

func registerRetryErrorHandler(c *colly.Collector, leasedDedicatedProxyURL string, releaseDedicatedProxy func(), dedicatedRetryThreshold, maxRetries int) {
	c.OnError(func(r *colly.Response, err error) {
		// Colly may call OnError without a full response/request on network/proxy failures.
		// Guard all dereferences to prevent intermittent panics that bubble up as 500s.
		if r == nil || r.Ctx == nil || r.Request == nil || r.Request.URL == nil {
			mode := "unknown"
			if leasedDedicatedProxyURL != "" {
				mode = "dedicated"
			}
			log.Printf("Request failed before response was available: %v (%s)", err, formatProxyContext(mode, leasedDedicatedProxyURL))
			releaseDedicatedProxy()
			return
		}

		retries, _ := r.Ctx.GetAny("retries").(int)
		statusCode := r.StatusCode

		if retries < maxRetries {
			nextRetry := retries + 1
			r.Ctx.Put("retries", nextRetry)
			previousProxyURL := r.Ctx.Get("last_proxy_url")
			proxyMode, proxyURL := applyProxyForRetryAttemptWithPinnedDedicated(c, nextRetry, previousProxyURL, leasedDedicatedProxyURL, dedicatedRetryThreshold, maxRetries)
			r.Ctx.Put("last_proxy_mode", proxyMode)
			r.Ctx.Put("last_proxy_url", proxyURL)

			waitDuration := retryDelay(statusCode, retryAfterHeaderValue(r), retries)
			waitDuration = adjustRetryDelayForContextDeadline(waitDuration, c.Context, nextRetry, maxRetries)
			log.Printf(
				"Retrying %s (attempt=%d status=%d wait=%s %s)",
				r.Request.URL,
				nextRetry,
				statusCode,
				waitDuration,
				formatProxyContext(proxyMode, proxyURL),
			)
			time.Sleep(waitDuration)
			if retryErr := r.Request.Retry(); retryErr != nil {
				log.Printf(
					"Retry failed for %s (attempt=%d status=%d): %v (%s)",
					r.Request.URL,
					nextRetry,
					statusCode,
					retryErr,
					formatProxyContext(proxyMode, proxyURL),
				)
				releaseDedicatedProxy()
			}
		} else {
			log.Printf(
				"Failed after %d retries: %s (status=%d %s)",
				maxRetries,
				r.Request.URL,
				statusCode,
				formatProxyContext(r.Ctx.Get("last_proxy_mode"), r.Ctx.Get("last_proxy_url")),
			)
			releaseDedicatedProxy()
		}
	})
}

// VisitWithProxyInfo wraps collector visit errors with the selected proxy context.
func VisitWithProxyInfo(c *colly.Collector, targetURL string) error {
	var proxyMu sync.Mutex
	var lastProxyMode string
	var lastProxyURL string

	c.OnRequest(func(r *colly.Request) {
		if r == nil || r.Ctx == nil {
			return
		}

		proxyMu.Lock()
		lastProxyMode = r.Ctx.Get("last_proxy_mode")
		lastProxyURL = r.Ctx.Get("last_proxy_url")
		proxyMu.Unlock()
	})

	err := c.Visit(targetURL)
	if err == nil {
		return nil
	}

	proxyMu.Lock()
	mode := lastProxyMode
	proxyURL := lastProxyURL
	proxyMu.Unlock()

	return fmt.Errorf("%w (%s)", err, formatProxyContext(mode, proxyURL))
}

// Retry metadata helpers.
func retryAfterHeaderValue(r *colly.Response) string {
	if r == nil || r.Headers == nil {
		return ""
	}
	return r.Headers.Get("Retry-After")
}

// Proxy strategy helpers.
func applyProxyForRetryAttempt(c *colly.Collector, retryAttempt int, avoidProxyURL string) (string, string) {
	return applyProxyForRetryAttemptWithPinnedDedicated(c, retryAttempt, avoidProxyURL, "", dedicatedProxyRetryThreshold, defaultMaxRetries)
}

func clearProxy(c *colly.Collector) (string, string) {
	c.SetProxyFunc(nil)
	return "direct", ""
}

func isDedicatedRetryAttempt(retryAttempt int, dedicatedRetryThreshold int) bool {
	return retryAttempt <= dedicatedRetryThreshold
}

func isFinalRetryAttempt(retryAttempt, maxRetries int) bool {
	return retryAttempt >= maxRetries
}

func applyProxyForRetryAttemptWithPinnedDedicated(c *colly.Collector, retryAttempt int, avoidProxyURL, pinnedDedicatedProxyURL string, dedicatedRetryThreshold, maxRetries int) (string, string) {
	if !config.UseProxy {
		return clearProxy(c)
	}

	// Final retry is always direct to bypass potentially bad proxy routes.
	if isFinalRetryAttempt(retryAttempt, maxRetries) {
		return clearProxy(c)
	}

	// Keep pinned dedicated proxy only while retryAttempt <= dedicatedRetryThreshold (see constants).
	// After that, prefer shared PROXY_URL fallback when available.
	if pinnedDedicatedProxyURL != "" && isDedicatedRetryAttempt(retryAttempt, dedicatedRetryThreshold) {
		c.SetProxy(pinnedDedicatedProxyURL)
		return "dedicated", pinnedDedicatedProxyURL
	}

	if isDedicatedRetryAttempt(retryAttempt, dedicatedRetryThreshold) {
		if proxyURL, ok := randomDedicatedProxyURL(avoidProxyURL); ok {
			c.SetProxy(proxyURL)
			return "dedicated", proxyURL
		}
		return clearProxy(c)
	}

	if proxyURL := os.Getenv("PROXY_URL"); proxyURL != "" {
		c.SetProxy(proxyURL)
		return "shared", proxyURL
	}

	return clearProxy(c)
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
	case "shared":
		if sharedProxyURL := os.Getenv("PROXY_URL"); sharedProxyURL != "" && sharedProxyURL == proxyURL {
			return "PROXY_URL"
		}
		return "shared-configured"
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

func retryDelay(statusCode int, retryAfterHeader string, retries int) time.Duration {
	if statusCode == 429 {
		if retryAfter, ok := parseRetryAfter(retryAfterHeader); ok {
			return retryAfter
		}

		// Stronger backoff for explicit throttling.
		base := time.Duration(retries+2) * time.Second
		jitter := time.Duration(rand.IntN(1000)) * time.Millisecond
		return base + jitter
	}

	base := time.Duration(retries+1) * time.Second
	jitter := time.Duration(rand.IntN(500)) * time.Millisecond
	return base + jitter
}

func adjustRetryDelayForContextDeadline(waitDuration time.Duration, requestCtx context.Context, nextRetry, maxRetries int) time.Duration {
	if requestCtx == nil {
		return waitDuration
	}

	if isFinalRetryAttempt(nextRetry, maxRetries) {
		// Prioritize issuing the final retry (which is direct/no proxy) before the context expires.
		return 0
	}

	deadline, hasDeadline := requestCtx.Deadline()
	if !hasDeadline {
		return waitDuration
	}

	remaining := time.Until(deadline)
	if remaining <= retryExecutionBuffer {
		return 0
	}

	maxWait := remaining - retryExecutionBuffer
	if waitDuration > maxWait {
		return maxWait
	}

	return waitDuration
}

func parseRetryAfter(value string) (time.Duration, bool) {
	if value == "" {
		return 0, false
	}

	seconds, err := strconv.Atoi(value)
	if err == nil {
		if seconds <= 0 {
			return 0, false
		}
		return time.Duration(seconds) * time.Second, true
	}

	retryAt, err := time.Parse(time.RFC1123, value)
	if err != nil {
		return 0, false
	}
	wait := time.Until(retryAt)
	if wait <= 0 {
		return 0, false
	}
	return wait, true
}
