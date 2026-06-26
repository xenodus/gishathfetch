package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	UtmSource        = "gishathfetch"
	// MinSearchStringLength is the minimum number of characters required for a search.
	MinSearchStringLength = 3
	// MaxSearchStringLength caps card name searches. The longest MTG card name is
	// ~141 characters (Unhinged); 150 allows any real card name while rejecting
	// bot paragraph spam.
	MaxSearchStringLength = 150
	MaxPagesToSearch      = 3
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
	// WebBotAuthEnabledEnv toggles RFC 9421 Web Bot Auth signing on outbound gateway requests.
	WebBotAuthEnabledEnv = "WEB_BOT_AUTH_ENABLED"
	// WebBotAuthPrivateKeyEnv holds a PEM (or base64-encoded PEM) Ed25519 PKCS8 private key.
	WebBotAuthPrivateKeyEnv = "WEB_BOT_AUTH_PRIVATE_KEY"
	// WebBotAuthPrivateKeyFileEnv holds a filesystem path to PEM key material.
	// Prefer this in CI so the raw key is not kept in process environment variables.
	WebBotAuthPrivateKeyFileEnv = "WEB_BOT_AUTH_PRIVATE_KEY_FILE"
	// WebBotAuthSignatureAgentEnv is the Signature-Agent directory URL published by this bot.
	WebBotAuthSignatureAgentEnv = "WEB_BOT_AUTH_SIGNATURE_AGENT"
	// WebBotAuthUserAgentEnv optionally overrides the stable bot User-Agent when signing is enabled.
	WebBotAuthUserAgentEnv = "WEB_BOT_AUTH_USER_AGENT"
	// WebBotAuthTTLEnv optionally overrides signature validity in seconds (default 24h).
	WebBotAuthTTLEnv = "WEB_BOT_AUTH_TTL_SECONDS"
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

// WebBotAuthTTL returns how long outbound Web Bot Auth signatures remain valid.
func WebBotAuthTTL() time.Duration {
	const defaultTTL = 24 * time.Hour
	rawValue := strings.TrimSpace(os.Getenv(WebBotAuthTTLEnv))
	if rawValue == "" {
		return defaultTTL
	}
	seconds, err := strconv.Atoi(rawValue)
	if err != nil || seconds <= 0 {
		return defaultTTL
	}
	return time.Duration(seconds) * time.Second
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
