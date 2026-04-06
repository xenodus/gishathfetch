package gateway

import (
	"context"
	"math/rand/v2"

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
	c := colly.NewCollector(colly.StdlibContext(ctx))
	ConfigureRequestOptimizations(c)
	return c
}

func ConfigureRequestOptimizations(c *colly.Collector) {
	c.DisableCookies()
	c.OnRequest(func(r *colly.Request) {
		// Keep gzip only. Go's default client does not transparently decode brotli ("br").
		r.Headers.Set("Accept-Encoding", "gzip")
		r.Headers.Set("User-Agent", browserUserAgents[rand.IntN(len(browserUserAgents))])
	})
}
