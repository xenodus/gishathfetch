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
	// SearchAttemptTimeout bounds a single search strategy attempt (BinderPOS step,
	// Shopify suggest transport, or default colly scrape).
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
	// AnalyticsLatestJSONCacheControl is applied to latest.json so CloudFront can cache it
	// between daily exports without a separate invalidation.
	AnalyticsLatestJSONCacheControl = "public, max-age=3600"
	// RobotsTxtCacheControl is applied to robots.txt so CloudFront can cache it between daily exports.
	RobotsTxtCacheControl = "public, max-age=3600"
	// SiteBaseURL is the public frontend origin used when generating robots.txt search URLs.
	SiteBaseURL = "https://gishathfetch.com/"
	// AWSRegion is the AWS region used for DynamoDB and other managed services.
	AWSRegion = "ap-southeast-1"
	// AdminUsernameEnv is the admin login username.
	AdminUsernameEnv = "ADMIN_USERNAME"
	// AdminPasswordEnv is the admin login password.
	AdminPasswordEnv = "ADMIN_PASSWORD"
	// AdminSessionSecretEnv signs admin session cookies.
	AdminSessionSecretEnv = "ADMIN_SESSION_SECRET"
	// AdminLoginDynamoDBTableEnv stores login attempt logs and rate-limit state.
	AdminLoginDynamoDBTableEnv = "ADMIN_LOGIN_DYNAMODB_TABLE"
	// AdminLoginMaxFailuresPerIPEnv caps failed logins per IP inside the IP window.
	AdminLoginMaxFailuresPerIPEnv = "ADMIN_LOGIN_MAX_FAILURES_PER_IP"
	// AdminLoginIPWindowSecondsEnv is the rolling failure window for IP rate limits.
	AdminLoginIPWindowSecondsEnv = "ADMIN_LOGIN_IP_WINDOW_SECONDS"
	// AdminLoginIPLockoutSecondsEnv is how long an IP stays locked after exceeding the cap.
	AdminLoginIPLockoutSecondsEnv = "ADMIN_LOGIN_IP_LOCKOUT_SECONDS"
	// AdminLoginMaxFailuresPerUserEnv caps failed logins per username inside the user window.
	AdminLoginMaxFailuresPerUserEnv = "ADMIN_LOGIN_MAX_FAILURES_PER_USER"
	// AdminLoginUserWindowSecondsEnv is the rolling failure window for username rate limits.
	AdminLoginUserWindowSecondsEnv = "ADMIN_LOGIN_USER_WINDOW_SECONDS"
	// AdminLoginUserLockoutSecondsEnv is how long a username stays locked after exceeding the cap.
	AdminLoginUserLockoutSecondsEnv = "ADMIN_LOGIN_USER_LOCKOUT_SECONDS"
	// AdminSessionTTLEnv overrides admin session lifetime in seconds (default 8h).
	AdminSessionTTLEnv = "ADMIN_SESSION_TTL_SECONDS"
	// AdminAttemptLogRetentionDaysEnv overrides attempt-log TTL in days (default 90).
	AdminAttemptLogRetentionDaysEnv = "ADMIN_ATTEMPT_LOG_RETENTION_DAYS"
	// DefaultAdminLoginMaxFailuresPerIP is the default failed-login cap per IP.
	DefaultAdminLoginMaxFailuresPerIP = 5
	// DefaultAdminLoginIPWindow is the default rolling window for IP failures.
	DefaultAdminLoginIPWindow = 15 * time.Minute
	// DefaultAdminLoginIPLockout is the default IP lockout duration.
	DefaultAdminLoginIPLockout = 15 * time.Minute
	// DefaultAdminLoginMaxFailuresPerUser is the default failed-login cap per username.
	DefaultAdminLoginMaxFailuresPerUser = 10
	// DefaultAdminLoginUserWindow is the default rolling window for username failures.
	DefaultAdminLoginUserWindow = 30 * time.Minute
	// DefaultAdminLoginUserLockout is the default username lockout duration.
	DefaultAdminLoginUserLockout = 30 * time.Minute
	// DefaultAdminSessionTTL is the default admin session lifetime.
	DefaultAdminSessionTTL = 8 * time.Hour
	// DefaultAdminAttemptLogRetention is how long attempt logs are retained.
	DefaultAdminAttemptLogRetention = 90 * 24 * time.Hour
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

// AdminEnabled reports whether admin login is fully configured.
func AdminEnabled() bool {
	return strings.TrimSpace(os.Getenv(AdminUsernameEnv)) != "" &&
		strings.TrimSpace(os.Getenv(AdminPasswordEnv)) != "" &&
		strings.TrimSpace(os.Getenv(AdminSessionSecretEnv)) != "" &&
		strings.TrimSpace(os.Getenv(AdminLoginDynamoDBTableEnv)) != ""
}

func AdminSessionTTL() time.Duration {
	return durationFromEnv(AdminSessionTTLEnv, DefaultAdminSessionTTL)
}

func AdminAttemptLogRetention() time.Duration {
	days := intFromEnv(AdminAttemptLogRetentionDaysEnv, int(DefaultAdminAttemptLogRetention.Hours()/24))
	if days <= 0 {
		days = int(DefaultAdminAttemptLogRetention.Hours() / 24)
	}
	return time.Duration(days) * 24 * time.Hour
}

type AdminLoginRateLimits struct {
	MaxFailuresPerIP     int
	IPWindow             time.Duration
	IPLockout            time.Duration
	MaxFailuresPerUser   int
	UserWindow           time.Duration
	UserLockout          time.Duration
}

func AdminLoginRateLimitsFromEnv() AdminLoginRateLimits {
	return AdminLoginRateLimits{
		MaxFailuresPerIP: intFromEnv(
			AdminLoginMaxFailuresPerIPEnv,
			DefaultAdminLoginMaxFailuresPerIP,
		),
		IPWindow: durationFromEnv(
			AdminLoginIPWindowSecondsEnv,
			DefaultAdminLoginIPWindow,
		),
		IPLockout: durationFromEnv(
			AdminLoginIPLockoutSecondsEnv,
			DefaultAdminLoginIPLockout,
		),
		MaxFailuresPerUser: intFromEnv(
			AdminLoginMaxFailuresPerUserEnv,
			DefaultAdminLoginMaxFailuresPerUser,
		),
		UserWindow: durationFromEnv(
			AdminLoginUserWindowSecondsEnv,
			DefaultAdminLoginUserWindow,
		),
		UserLockout: durationFromEnv(
			AdminLoginUserLockoutSecondsEnv,
			DefaultAdminLoginUserLockout,
		),
	}
}

func durationFromEnv(name string, fallback time.Duration) time.Duration {
	seconds := intFromEnv(name, int(fallback.Seconds()))
	if seconds <= 0 {
		return fallback
	}
	return time.Duration(seconds) * time.Second
}

func intFromEnv(name string, fallback int) int {
	rawValue := strings.TrimSpace(os.Getenv(name))
	if rawValue == "" {
		return fallback
	}
	value, err := strconv.Atoi(rawValue)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}
