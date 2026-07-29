package apiauth

import (
	"crypto/subtle"
	"errors"
	"slices"
	"strings"

	"mtg-price-checker-sg/pkg/config"
)

var errOriginVerifyFailed = errors.New("origin verification failed")

// VerifyOriginHeader enforces access when API_ORIGIN_VERIFY_SECRET is set.
// Requests with a matching X-Origin-Verify header pass (e.g. CloudFront origin).
// Browser calls to api.gishathfetch.com may omit that header and use an allowlisted Origin instead.
func VerifyOriginHeader(headers map[string]string) error {
	secret := config.APIOriginVerifySecret()
	if secret == "" {
		return nil
	}

	headerName := strings.ToLower(config.APIOriginVerifyHeader())
	got := strings.TrimSpace(headerValue(headers, headerName))
	if got != "" {
		if subtle.ConstantTimeCompare([]byte(got), []byte(secret)) == 1 {
			return nil
		}
		return errOriginVerifyFailed
	}

	origin := strings.TrimSpace(headerValue(headers, "origin"))
	if origin != "" && slices.Contains(config.GetAllowedOrigins(), origin) {
		return nil
	}

	return errOriginVerifyFailed
}

func headerValue(headers map[string]string, lowerName string) string {
	if headers == nil {
		return ""
	}
	if v := headers[lowerName]; v != "" {
		return v
	}
	for k, v := range headers {
		if strings.EqualFold(k, lowerName) {
			return v
		}
	}
	return ""
}
