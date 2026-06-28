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
	EnvProd  = "prod"
	EnvLocal = "local"
	UseProxy         = true
	PerSiteTimeout   = 20 * time.Second
	// DynamicProxyEnv contains an authenticated proxy URL used for explicit
	// dynamic-proxy fallback attempts, which BinderPOS now reserves for the
	// final fallback after dedicated and direct/no-proxy attempts.
	DynamicProxyEnv = "DYNAMIC_PROXY"
	// UseDynamicProxyEnv toggles whether DYNAMIC_PROXY may be used for fallback
	// attempts. When false, dynamic proxy is skipped even if configured.
	UseDynamicProxyEnv = "USE_DYNAMIC_PROXY"
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
	// CKDynamoDBTableEnv is the DynamoDB table storing cheapest Card Kingdom prices by card name.
	CKDynamoDBTableEnv = "CK_DYNAMODB_TABLE"
	// CKPriceLookupEnabledEnv toggles Card Kingdom price lookup on search responses.
	CKPriceLookupEnabledEnv = "CK_PRICE_LOOKUP_ENABLED"
	// CKPriceMaxAge is how old a DynamoDB CK listing may be before search omits it.
	CKPriceMaxAge = 24 * time.Hour
	// MTGJSONAllPricesTodayURLEnv overrides the MTGJSON AllPricesToday download URL.
	MTGJSONAllPricesTodayURLEnv = "MTGJSON_ALL_PRICES_TODAY_URL"
	// MTGJSONAllPrintingsURLEnv overrides the MTGJSON AllPrintings download URL.
	MTGJSONAllPrintingsURLEnv = "MTGJSON_ALL_PRINTINGS_URL"
	// AWSRegion is the AWS region used for DynamoDB and other managed services.
	AWSRegion = "ap-southeast-1"
)

// UseLeasedDedicatedProxy enables exclusive per-request leases from the dedicated proxy pool.
// When false, each request picks a random dedicated proxy instead of acquiring a lease.
const UseLeasedDedicatedProxy = false

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

// CKPriceLookupEnabled reports whether search responses should include Card Kingdom prices.
func CKPriceLookupEnabled() bool {
	rawValue := strings.TrimSpace(os.Getenv(CKPriceLookupEnabledEnv))
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
	return []string{
		"https://gishathfetch.com",
		"http://localhost:5173",
		"http://localhost:63342", // JetBrains IDE built-in HTTP server (local dev only)
	}
}
