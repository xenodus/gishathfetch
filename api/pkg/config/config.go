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
	// DirectSearchAttemptTimeout bounds a single direct-egress search attempt.
	DirectSearchAttemptTimeout = 3 * time.Second
	// DedicatedSearchAttemptTimeout bounds a single dedicated-proxy search attempt
	// (BinderPOS dedicated step or default colly scrape via proxy).
	DedicatedSearchAttemptTimeout = 5 * time.Second
	// AgoraSearchAttemptTimeout is the per-attempt cap for Agora Hobby only (matches per-store deadline).
	AgoraSearchAttemptTimeout = PerSiteTimeout
	// MoxAndLotusSearchAttemptTimeout is the per-attempt cap for Mox & Lotus.
	MoxAndLotusSearchAttemptTimeout = 10 * time.Second
	// UseDedicatedProxyEnv toggles whether DEDICATED_PROXY_* may be used for outbound
	// scrapes and API calls. When false, dedicated proxy transports are skipped even
	// if configured. Defaults to enabled when unset or invalid.
	UseDedicatedProxyEnv = "USE_DEDICATED_PROXY"
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
	// CKPriceLookupTimeout caps Card Kingdom enrichment on /search so a slow
	// Scryfall/DynamoDB path cannot delay returning store results.
	CKPriceLookupTimeout = 2 * time.Second
	// CKPriceMaxAge is how old a DynamoDB CK listing may be before search omits it.
	CKPriceMaxAge = 48 * time.Hour
	// CKPricelistURLEnv overrides the Card Kingdom singles pricelist download URL.
	CKPricelistURLEnv = "CK_PRICELIST_URL"
	// ResidentialProxyEnv is an optional residential proxy used by stores that
	// rate-limit datacenter IPs. Format matches DEDICATED_PROXY_*.
	ResidentialProxyEnv = "RESIDENTIAL_PROXY_1"
	// BrowserTLSEmulationEnabledEnv toggles browser TLS fingerprinting and
	// matched User-Agent emulation for outbound scrapers. Defaults to enabled.
	BrowserTLSEmulationEnabledEnv = "BROWSER_TLS_EMULATION_ENABLED"
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

	// APIOriginVerifySecretEnv is the shared secret CloudFront adds as a custom origin header.
	APIOriginVerifySecretEnv = "API_ORIGIN_VERIFY_SECRET"
	// APIOriginVerifyHeaderEnv overrides the origin verify header name (default X-Origin-Verify).
	APIOriginVerifyHeaderEnv = "API_ORIGIN_VERIFY_HEADER"
	// APISessionSecretEnv signs HttpOnly browser session cookies for search requests.
	APISessionSecretEnv = "API_SESSION_SECRET"
	// APISessionTTLEnv overrides session lifetime in seconds (default 15 minutes).
	APISessionTTLEnv = "API_SESSION_TTL_SECONDS"
	// APIMaintenanceModeEnv toggles maintenance mode for /search. When true, search
	// requests are rejected and the API advertises maintenance via response headers.
	APIMaintenanceModeEnv = "API_MAINTENANCE_MODE"
	// APIMaintenanceMessageEnv overrides the user-visible maintenance message when
	// API_MAINTENANCE_MODE is enabled. Ignored when maintenance mode is off.
	APIMaintenanceMessageEnv = "API_MAINTENANCE_MESSAGE"
	// APINoticeMessageEnv is optional site-wide notice text advertised via /session
	// whenever non-empty. Unlike maintenance mode, search remains available.
	APINoticeMessageEnv = "API_NOTICE_MESSAGE"
	// DefaultAPIMaintenanceMessage is shown when maintenance mode is on but no
	// custom message is configured.
	DefaultAPIMaintenanceMessage = "Search is temporarily unavailable. Please try again later."
	// CardsCentralKeyEnv is the API key for Cards Central LGS search requests.
	CardsCentralKeyEnv = "CARDS_CENTRAL_KEY"
	// CardsCentralKeyHeader is the request header Cards Central expects for aggregator access.
	CardsCentralKeyHeader = "x-gishath-key"
	// APITelegramBotTokenEnv authorizes GET /telegram/search for the Telegram bot service.
	APITelegramBotTokenEnv = "API_TELEGRAM_BOT_TOKEN"
)

// UseLeasedDedicatedProxy enables exclusive per-request leases from the dedicated proxy pool.
// When false, each request picks a random dedicated proxy instead of acquiring a lease.
const UseLeasedDedicatedProxy = false

// AgoraSearchEnabled toggles Agora Hobby search in the search Lambda.
const AgoraSearchEnabled = true

// UseDedicatedProxy reports whether DEDICATED_PROXY_* env vars may be used.
// Defaults to enabled when unset or invalid.
func UseDedicatedProxy() bool {
	rawValue := strings.TrimSpace(os.Getenv(UseDedicatedProxyEnv))
	if rawValue == "" {
		return true
	}

	enabled, err := strconv.ParseBool(rawValue)
	if err != nil {
		return true
	}

	return enabled
}

// BrowserTLSEmulationEnabled reports whether outbound scrapers should use
// browser-matched TLS fingerprints instead of Go's default client hello.
func BrowserTLSEmulationEnabled() bool {
	rawValue := strings.TrimSpace(os.Getenv(BrowserTLSEmulationEnabledEnv))
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

const (
	defaultAPIOriginVerifyHeader = "X-Origin-Verify"
	defaultAPISessionTTL         = 15 * time.Minute
)

// APIOriginVerifySecret returns the CloudFront origin verification secret when set.
func APIOriginVerifySecret() string {
	return strings.TrimSpace(os.Getenv(APIOriginVerifySecretEnv))
}

// APIOriginVerifyHeader returns the header CloudFront must send to API Gateway.
func APIOriginVerifyHeader() string {
	if v := strings.TrimSpace(os.Getenv(APIOriginVerifyHeaderEnv)); v != "" {
		return v
	}
	return defaultAPIOriginVerifyHeader
}

// APISessionSecret returns the HMAC secret for browser session cookies when set.
func APISessionSecret() string {
	return strings.TrimSpace(os.Getenv(APISessionSecretEnv))
}

// APISessionTTL is how long a minted browser session remains valid.
func APISessionTTL() time.Duration {
	rawValue := strings.TrimSpace(os.Getenv(APISessionTTLEnv))
	if rawValue == "" {
		return defaultAPISessionTTL
	}
	seconds, err := strconv.Atoi(rawValue)
	if err != nil || seconds <= 0 {
		return defaultAPISessionTTL
	}
	return time.Duration(seconds) * time.Second
}

// APIMaintenanceMode reports whether /search is disabled for maintenance.
func APIMaintenanceMode() bool {
	rawValue := strings.TrimSpace(os.Getenv(APIMaintenanceModeEnv))
	if rawValue == "" {
		return false
	}

	enabled, err := strconv.ParseBool(rawValue)
	if err != nil {
		return false
	}

	return enabled
}

// APIMaintenanceMessage returns the user-visible maintenance message when
// maintenance mode is enabled.
func APIMaintenanceMessage() string {
	if message := strings.TrimSpace(os.Getenv(APIMaintenanceMessageEnv)); message != "" {
		return message
	}
	return DefaultAPIMaintenanceMessage
}

// APINoticeMessage returns optional site-wide notice text when configured.
func APINoticeMessage() string {
	return strings.TrimSpace(os.Getenv(APINoticeMessageEnv))
}

// APITelegramBotToken returns the bearer token required for /telegram/search.
func APITelegramBotToken() string {
	return strings.TrimSpace(os.Getenv(APITelegramBotTokenEnv))
}

// APIAccessControlEnabled is true when origin verification or session enforcement is configured.
func APIAccessControlEnabled() bool {
	return APIOriginVerifySecret() != "" || APISessionSecret() != ""
}
