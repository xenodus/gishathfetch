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
	// UseBinderposSharedProxyFallbackEnv controls BinderPOS scraper retry path.
	// Default is disabled to use dedicated proxy and then direct. Set to "true"
	// to restore the previous dedicated -> shared -> direct behavior.
	UseBinderposSharedProxyFallbackEnv = "USE_BINDERPOS_SHARED_PROXY_FALLBACK"
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
