package apiauth

import (
	"os"
	"testing"

	"mtg-price-checker-sg/pkg/config"

	"github.com/stretchr/testify/require"
)

func TestVerifyTelegramBotToken(t *testing.T) {
	t.Setenv(config.APITelegramBotTokenEnv, "telegram-bot-secret")

	require.Error(t, VerifyTelegramBotToken(nil))
	require.Error(t, VerifyTelegramBotToken(map[string]string{
		"authorization": "Bearer wrong",
	}))
	require.NoError(t, VerifyTelegramBotToken(map[string]string{
		"authorization": "Bearer telegram-bot-secret",
	}))
	require.NoError(t, VerifyTelegramBotToken(map[string]string{
		"Authorization": "bearer telegram-bot-secret",
	}))
}

func TestVerifyTelegramBotToken_NotConfigured(t *testing.T) {
	require.NoError(t, os.Unsetenv(config.APITelegramBotTokenEnv))
	require.Error(t, VerifyTelegramBotToken(map[string]string{
		"authorization": "Bearer anything",
	}))
}
