package webbotauth

import (
	"fmt"
	"os"
	"strings"

	"mtg-price-checker-sg/pkg/config"
)

// LoadPrivateKeyPEM reads signing key material from WEB_BOT_AUTH_PRIVATE_KEY_FILE
// or WEB_BOT_AUTH_PRIVATE_KEY. Prefer the file path in CI so the raw key is not
// kept in the process environment longer than necessary.
func LoadPrivateKeyPEM() (string, error) {
	if keyFile := strings.TrimSpace(os.Getenv(config.WebBotAuthPrivateKeyFileEnv)); keyFile != "" {
		data, err := os.ReadFile(keyFile)
		if err != nil {
			return "", fmt.Errorf("read %s: %w", config.WebBotAuthPrivateKeyFileEnv, err)
		}
		return strings.TrimSpace(string(data)), nil
	}

	if pemData := strings.TrimSpace(os.Getenv(config.WebBotAuthPrivateKeyEnv)); pemData != "" {
		return pemData, nil
	}

	return "", fmt.Errorf("%s or %s is required", config.WebBotAuthPrivateKeyFileEnv, config.WebBotAuthPrivateKeyEnv)
}

// PrivateKeyConfigured reports whether signing key material is available.
func PrivateKeyConfigured() bool {
	if keyFile := strings.TrimSpace(os.Getenv(config.WebBotAuthPrivateKeyFileEnv)); keyFile != "" {
		if _, err := os.Stat(keyFile); err == nil {
			return true
		}
	}
	return strings.TrimSpace(os.Getenv(config.WebBotAuthPrivateKeyEnv)) != ""
}
