package apiauth

import (
	"crypto/subtle"
	"errors"
	"strings"

	"mtg-price-checker-sg/pkg/config"
)

var errOriginVerifyFailed = errors.New("origin verification failed")

// VerifyOriginHeader ensures the CloudFront origin secret header matches when configured.
func VerifyOriginHeader(headers map[string]string) error {
	secret := config.APIOriginVerifySecret()
	if secret == "" {
		return nil
	}

	headerName := strings.ToLower(config.APIOriginVerifyHeader())
	got := strings.TrimSpace(headerValue(headers, headerName))
	if got == "" || subtle.ConstantTimeCompare([]byte(got), []byte(secret)) != 1 {
		return errOriginVerifyFailed
	}

	return nil
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
