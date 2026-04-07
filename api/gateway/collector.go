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

func NewOptimizedCollector(ctx context.Context) *colly.Collector {
	c := colly.NewCollector(
		colly.StdlibContext(ctx),
	)
	ConfigureRequestOptimizations(c)
	return c
}

func NewOptimizedCollectorNoRetry(ctx context.Context) *colly.Collector {
	c := colly.NewCollector(
		colly.StdlibContext(ctx),
	)
	ConfigureRequestOptimizationsNoRetry(c)
	return c
}

func ConfigureRequestOptimizations(c *colly.Collector) {
	configureRequestOptimizations(c, true)
}

func ConfigureRequestOptimizationsNoRetry(c *colly.Collector) {
	configureRequestOptimizations(c, false)
}

func configureRequestOptimizations(c *colly.Collector, enableRetry bool) {
	c.DisableCookies()
	c.SetRequestTimeout(config.PerSiteTimeout)
	initialProxyMode, initialProxyURL := applyProxyForRetryAttempt(c, 0, "")
	c.OnRequest(func(r *colly.Request) {
		// Keep gzip only. Go's default client does not transparently decode brotli ("br").
		r.Headers.Set("Accept-Encoding", "gzip")
		r.Headers.Set("User-Agent", browserUserAgents[rand.IntN(len(browserUserAgents))])
		if r.Ctx != nil && r.Ctx.Get("last_proxy_url") == "" {
			r.Ctx.Put("last_proxy_mode", initialProxyMode)
			r.Ctx.Put("last_proxy_url", initialProxyURL)
		}
	})

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: 2 * time.Second, // adds 0–2s jitter between matching-domain requests
	})

	if !enableRetry {
		return
	}

	const maxRetries = 3

	c.OnError(func(r *colly.Response, err error) {
		// Colly may call OnError without a full response/request on network/proxy failures.
		// Guard all dereferences to prevent intermittent panics that bubble up as 500s.
		if r == nil || r.Ctx == nil || r.Request == nil || r.Request.URL == nil {
			fmt.Printf("Request failed before response was available: %v\n", err)
			return
		}

		retries, _ := r.Ctx.GetAny("retries").(int)
		statusCode := r.StatusCode

		if retries < maxRetries {
			nextRetry := retries + 1
			r.Ctx.Put("retries", nextRetry)
			previousProxyURL := r.Ctx.Get("last_proxy_url")
			proxyMode, proxyURL := applyProxyForRetryAttempt(c, nextRetry, previousProxyURL)
			r.Ctx.Put("last_proxy_mode", proxyMode)
			r.Ctx.Put("last_proxy_url", proxyURL)

			retryAfterHeader := ""
			if r.Headers != nil {
				retryAfterHeader = r.Headers.Get("Retry-After")
			}
			waitDuration := retryDelay(statusCode, retryAfterHeader, retries)
			log.Printf("Retrying %s (attempt=%d status=%d wait=%s proxy_mode=%s)", r.Request.URL, nextRetry, statusCode, waitDuration, proxyMode)
			time.Sleep(waitDuration)
			if retryErr := r.Request.Retry(); retryErr != nil {
				log.Printf("Retry failed for %s (attempt=%d status=%d): %v", r.Request.URL, nextRetry, statusCode, retryErr)
			}
		} else {
			log.Printf("Failed after %d retries: %s (status=%d)", maxRetries, r.Request.URL, statusCode)
		}
	})
}

func applyProxyForRetryAttempt(c *colly.Collector, retryAttempt int, avoidProxyURL string) (string, string) {
	if !config.UseProxy {
		c.SetProxyFunc(nil)
		return "direct", ""
	}

	switch retryAttempt {
	case 0, 1:
		if proxyURL, ok := randomDedicatedProxyURL(avoidProxyURL); ok {
			c.SetProxy(proxyURL)
			return "dedicated", proxyURL
		}
	default:
		if proxyURL := os.Getenv("PROXY_URL"); proxyURL != "" {
			c.SetProxy(proxyURL)
			return "shared", proxyURL
		}
	}

	c.SetProxyFunc(nil)
	return "direct", ""
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
			proxyURL := fmt.Sprintf("http://%s:%s", p.Host, p.Port)
			if p.Username != "" || p.Password != "" {
				proxyURL = fmt.Sprintf("http://%s:%s@%s:%s", p.Username, p.Password, p.Host, p.Port)
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
	if p.Username != "" || p.Password != "" {
		return fmt.Sprintf("http://%s:%s@%s:%s", p.Username, p.Password, p.Host, p.Port), true
	}

	return fmt.Sprintf("http://%s:%s", p.Host, p.Port), true
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
