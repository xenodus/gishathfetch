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
	// DynamicProxyEnv contains an authenticated proxy URL used for explicit
	// dynamic-proxy fallback attempts, which BinderPOS now reserves for the
	// final fallback after dedicated and direct/no-proxy attempts.
	DynamicProxyEnv = "DYNAMIC_PROXY"
	// UseDynamicProxyEnv toggles whether DYNAMIC_PROXY may be used for fallback
	// attempts. When false, dynamic proxy is skipped even if configured.
	UseDynamicProxyEnv = "USE_DYNAMIC_PROXY"
	// UseBinderposSharedProxyFallbackEnv is reserved for future BinderPOS proxy
	// routing options. It no longer changes scraper behavior (lookups are single-attempt).
	UseBinderposSharedProxyFallbackEnv = "USE_BINDERPOS_SHARED_PROXY_FALLBACK"
)

// UseLeasedDedicatedProxy enables exclusive per-request leases from the dedicated proxy pool.
// When false, each request picks a random dedicated proxy instead of acquiring a lease.
const UseLeasedDedicatedProxy = false

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

func UseDynamicProxy() bool {
	rawValue := strings.TrimSpace(os.Getenv(UseDynamicProxyEnv))
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
