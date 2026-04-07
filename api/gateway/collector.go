package gateway

import (
	"context"
	"fmt"
	"math/rand/v2"
	"mtg-price-checker-sg/gateway/util"
	"mtg-price-checker-sg/pkg/config"
	"os"
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
	applyProxyForRetryAttempt(c, 0)
	c.OnRequest(func(r *colly.Request) {
		// Keep gzip only. Go's default client does not transparently decode brotli ("br").
		r.Headers.Set("Accept-Encoding", "gzip")
		r.Headers.Set("User-Agent", browserUserAgents[rand.IntN(len(browserUserAgents))])
	})

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		RandomDelay: 2 * time.Second, // adds 0–2s on top of Delay
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

		if retries < maxRetries {
			nextRetry := retries + 1
			r.Ctx.Put("retries", nextRetry)
			applyProxyForRetryAttempt(c, nextRetry)
			fmt.Printf("Retrying %s (attempt %d)...\n", r.Request.URL, nextRetry)
			time.Sleep(time.Duration(retries+1) * time.Second) // backoff
			if retryErr := r.Request.Retry(); retryErr != nil {
				fmt.Printf("Retry failed for %s: %v\n", r.Request.URL, retryErr)
			}
		} else {
			fmt.Printf("Failed after %d retries: %s\n", maxRetries, r.Request.URL)
		}
	})
}

func applyProxyForRetryAttempt(c *colly.Collector, retryAttempt int) {
	if !config.UseProxy {
		c.SetProxyFunc(nil)
		return
	}

	switch retryAttempt {
	case 0, 1:
		if proxyURL, ok := randomDedicatedProxyURL(); ok {
			c.SetProxy(proxyURL)
			return
		}
		fallthrough
	case 2:
		if proxyURL := os.Getenv("PROXY_URL"); proxyURL != "" {
			c.SetProxy(proxyURL)
			return
		}
		fallthrough
	default:
		c.SetProxyFunc(nil)
	}
}

func randomDedicatedProxyURL() (string, bool) {
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

	p := valid[rand.IntN(len(valid))]
	if p.Username != "" || p.Password != "" {
		return fmt.Sprintf("http://%s:%s@%s:%s", p.Username, p.Password, p.Host, p.Port), true
	}

	return fmt.Sprintf("http://%s:%s", p.Host, p.Port), true
}
