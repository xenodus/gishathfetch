package telegrambot

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPriceRunEventJSON(t *testing.T) {
	payload, err := json.Marshal(PriceRunEvent{
		Action: PriceRunAction,
		ChatID: 42,
		Query:  "Opt",
	})
	require.NoError(t, err)
	require.JSONEq(t, `{"action":"telegram-price-run","chatId":42,"query":"Opt"}`, string(payload))
}

func TestNewServiceFromConfig_WithoutLambdaNameUsesSync(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "tg")
	t.Setenv("TELEGRAM_WEBHOOK_SECRET", "secret")
	t.Setenv("API_TELEGRAM_BOT_TOKEN", "api")
	t.Setenv("TELEGRAM_BOT_LAMBDA_FUNCTION", "")
	t.Setenv("AWS_LAMBDA_FUNCTION_NAME", "")

	svc, err := NewServiceFromConfig(context.Background(), LoadConfig(), nil)
	require.NoError(t, err)
	require.NotNil(t, svc)
}
