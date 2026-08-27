package handler

import (
	"context"
	"encoding/json"
	"testing"

	"mtg-price-checker-sg/pkg/telegrambot"

	"github.com/stretchr/testify/require"
)

func TestHandle_TelegramPriceRunAction(t *testing.T) {
	t.Cleanup(resetTelegramServiceForTest)

	var sent string
	svc := telegrambot.NewService(
		"webhook-secret",
		&telegramWebhookStubGishath{
			search: func(_ context.Context, _ string) (*telegrambot.SearchSummary, error) {
				return &telegrambot.SearchSummary{
					ResultCount: 1,
					Cheapest:    &telegrambot.CardSummary{Name: "Bolt", Price: 2},
					WebsiteURL:  "https://gishathfetch.com/?s=Bolt",
				}, nil
			},
		},
		&telegramWebhookStubTelegram{send: func(_ context.Context, _ int64, text string) error {
			sent = text
			return nil
		}},
		nil,
		nil,
	)
	setTelegramServiceForTest(svc, nil)

	payload, err := json.Marshal(telegrambot.PriceRunEvent{
		Action: telegrambot.PriceRunAction,
		ChatID: 5,
		Query:  "Bolt",
	})
	require.NoError(t, err)

	_, err = Handle(context.Background(), payload)
	require.NoError(t, err)
	require.Contains(t, sent, "S$2.00")
}
