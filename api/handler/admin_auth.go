package handler

import (
	"os"
	"strings"

	"mtg-price-checker-sg/pkg/config"
)

func isAffiliateAdminAuthorized(headers map[string]string) bool {
	expected := strings.TrimSpace(os.Getenv(config.AffiliateAdminAPIKeyEnv))
	if expected == "" {
		return false
	}

	if token := bearerToken(headers); token != "" && token == expected {
		return true
	}

	for _, key := range []string{"x-admin-api-key", "X-Admin-Api-Key"} {
		if headers[key] == expected {
			return true
		}
	}

	return false
}

func bearerToken(headers map[string]string) string {
	authHeader := strings.TrimSpace(headers["authorization"])
	if authHeader == "" {
		authHeader = strings.TrimSpace(headers["Authorization"])
	}
	if authHeader == "" {
		return ""
	}

	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(authHeader, prefix))
}
