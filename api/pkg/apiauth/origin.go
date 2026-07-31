package apiauth

import (
	"crypto/subtle"
	"errors"
	"strings"

	"mtg-price-checker-sg/pkg/config"
)

var errOriginVerifyFailed = errors.New("origin verification failed")

// VerifyOriginHeader enforces access when API_ORIGIN_VERIFY_SECRET is set.
// Requests must include a matching X-Origin-Verify header (injected by CloudFront
// on the origin request, or by the Vite dev proxy). Allowlisted Origin alone is
// not accepted: it is trivially spoofed on direct execute-api calls that bypass
// CloudFront and WAF.
func VerifyOriginHeader(headers map[string]string) error {
	secret := config.APIOriginVerifySecret()
	if secret == "" {
		return nil
	}

	headerName := strings.ToLower(config.APIOriginVerifyHeader())
	got := strings.TrimSpace(headerValue(headers, headerName))
	if got == "" {
		return errOriginVerifyFailed
	}
	if subtle.ConstantTimeCompare([]byte(got), []byte(secret)) == 1 {
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
