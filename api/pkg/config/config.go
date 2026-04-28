package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	UtmSource        = "gishathfetch"
	MaxPagesToSearch = 3
	EnvProd          = "prod"
	EnvStaging       = "staging"
	EnvLocal         = "local"
	UseProxy         = true
	PerSiteTimeout   = 20 * time.Second
	// UseBinderposStorefrontAPIEnv toggles BinderPOS storefront API search mode.
	// Default is enabled; set to "false" to force legacy scraping.
	UseBinderposStorefrontAPIEnv = "USE_BINDERPOS_STOREFRONT_API"
	// UseBinderposSharedProxyFallbackEnv is reserved for future BinderPOS proxy
	// routing options. It no longer changes scraper behavior (lookups are single-attempt).
	UseBinderposSharedProxyFallbackEnv = "USE_BINDERPOS_SHARED_PROXY_FALLBACK"
	// UseLeasedDedicatedProxyEnv enables exclusive per-request leases from the dedicated proxy pool.
	// When false (default), each request picks a random dedicated proxy instead of acquiring a lease.
	UseLeasedDedicatedProxyEnv = "USE_LEASED_DEDICATED_PROXY"
)

func UseBinderposStorefrontAPI() bool {
	rawValue := strings.TrimSpace(os.Getenv(UseBinderposStorefrontAPIEnv))
	if rawValue == "" {
		return true
	}

	enabled, err := strconv.ParseBool(rawValue)
	if err != nil {
		return true
	}

	return enabled
}

func UseBinderposSharedProxyFallback() bool {
	rawValue := strings.TrimSpace(os.Getenv(UseBinderposSharedProxyFallbackEnv))
	if rawValue == "" {
		return false
	}

	enabled, err := strconv.ParseBool(rawValue)
	if err != nil {
		return false
	}

	return enabled
}

// UseLeasedDedicatedProxy returns whether dedicated proxy usage should acquire an exclusive lease
// from the pool. Default is false (random selection among configured dedicated proxies).
func UseLeasedDedicatedProxy() bool {
	rawValue := strings.TrimSpace(os.Getenv(UseLeasedDedicatedProxyEnv))
	if rawValue == "" {
		return false
	}

	enabled, err := strconv.ParseBool(rawValue)
	if err != nil {
		return false
	}

	return enabled
}

func GetAllowedOrigins() []string {
	if os.Getenv("ENV") == EnvProd {
		return []string{
			"https://gishathfetch.com",
		}
	}

	return []string{
		"https://gishathfetch.com",
		"https://staging.gishathfetch.com",
		"http://localhost:5173",
		"http://localhost:63342", // JetBrains IDE built-in HTTP server (local dev only)
	}
}
