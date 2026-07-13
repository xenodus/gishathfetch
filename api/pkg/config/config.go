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
	PerSiteTimeout   = 16 * time.Second
	// SearchAttemptTimeout bounds a single search strategy attempt (BinderPOS step
	// or default colly scrape).
	SearchAttemptTimeout = 5 * time.Second
	// AgoraSearchAttemptTimeout is the per-attempt cap for Agora Hobby only.
	AgoraSearchAttemptTimeout = 10 * time.Second
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
	CKPriceMaxAge = 48 * time.Hour
	// MTGJSONAllPricesTodayURLEnv overrides the MTGJSON AllPricesToday download URL.
	MTGJSONAllPricesTodayURLEnv = "MTGJSON_ALL_PRICES_TODAY_URL"
	// MTGJSONAllPrintingsURLEnv overrides the MTGJSON AllPrintings download URL.
	MTGJSONAllPrintingsURLEnv = "MTGJSON_ALL_PRINTINGS_URL"
	// GA4PropertyIDEnv is the numeric GA4 property ID used by the Data API.
	GA4PropertyIDEnv = "GA4_PROPERTY_ID"
	// GA4CredentialsJSONEnv holds a Google service account JSON key with Analytics read access.
	GA4CredentialsJSONEnv = "GA4_CREDENTIALS_JSON"
	// AnalyticsS3BucketEnv overrides the destination bucket for exported analytics reports.
	AnalyticsS3BucketEnv = "ANALYTICS_S3_BUCKET"
	// AnalyticsS3DefaultBucket is the frontend S3 bucket served by CloudFront.
	AnalyticsS3DefaultBucket = "gishathfetch.com"
	// AnalyticsS3KeyPrefixEnv is the object key prefix for exported analytics reports.
	AnalyticsS3KeyPrefixEnv = "ANALYTICS_S3_KEY_PREFIX"
	// AnalyticsS3DefaultKeyPrefix is the default object key prefix under the frontend bucket.
	AnalyticsS3DefaultKeyPrefix = "analytics/top-search-keywords"
	// CKPriceChangesS3Bucket is the frontend S3 bucket for exported CK price change reports.
	CKPriceChangesS3Bucket = AnalyticsS3DefaultBucket
	// CKPriceChangesS3KeyPrefix is the object key prefix for exported CK price change reports.
	CKPriceChangesS3KeyPrefix = "analytics/ck-price-changes"
	// CKPriceChangesLatestJSONCacheControl is applied to latest.json so CloudFront can cache it
	// between daily exports without a separate invalidation.
	CKPriceChangesLatestJSONCacheControl = AnalyticsLatestJSONCacheControl
	// AnalyticsLatestJSONCacheControl is applied to latest.json so CloudFront can cache it
	// between daily exports without a separate invalidation.
	AnalyticsLatestJSONCacheControl = "public, max-age=3600"
	// RobotsTxtCacheControl is applied to robots.txt so CloudFront can cache it between daily exports.
	RobotsTxtCacheControl = "public, max-age=3600"
	// SiteBaseURL is the public frontend origin used when generating robots.txt search URLs.
	SiteBaseURL = "https://gishathfetch.com/"
	// AWSRegion is the AWS region used for DynamoDB and other managed services.
	AWSRegion = "ap-southeast-1"
)

// UseLeasedDedicatedProxy enables exclusive per-request leases from the dedicated proxy pool.
// When false, each request picks a random dedicated proxy instead of acquiring a lease.
const UseLeasedDedicatedProxy = false

// BinderposScrapOnly routes BinderPOS storefront search through scrape strategies only.
// Decklist API code remains available for structure probes and live integration tests.
var BinderposScrapOnly = true

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
		return strings.TrimSpace(os.Getenv(CKDynamoDBTableEnv)) != ""
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
