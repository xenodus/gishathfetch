package apiauth

import (
	"crypto/subtle"
	"errors"
	"strings"

	"mtg-price-checker-sg/pkg/config"
)

var errBotTokenFailed = errors.New("bot token verification failed")

// VerifyTelegramBotToken validates the Telegram bot API token when
// API_TELEGRAM_BOT_TOKEN is configured. Accepts Authorization: Bearer <token>.
func VerifyTelegramBotToken(headers map[string]string) error {
	expected := config.APITelegramBotToken()
	if expected == "" {
		return errors.New("telegram bot token not configured")
	}

	got := bearerToken(headerValue(headers, "authorization"))
	if got == "" {
		return errBotTokenFailed
	}
	if subtle.ConstantTimeCompare([]byte(got), []byte(expected)) != 1 {
		return errBotTokenFailed
	}
	return nil
}

func bearerToken(authorization string) string {
	authorization = strings.TrimSpace(authorization)
	const prefix = "Bearer "
	if len(authorization) < len(prefix) || !strings.EqualFold(authorization[:len(prefix)], prefix) {
		return ""
	}
	return strings.TrimSpace(authorization[len(prefix):])
}
