package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"mtg-price-checker-sg/pkg/telegrambot"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/require"
)

func TestTelegramWebhook_AsyncPriceReturnsOK(t *testing.T) {
	t.Cleanup(resetTelegramServiceForTest)

	var enqueued bool
	svc := telegrambot.NewService(
		"webhook-secret",
		&telegramWebhookStubGishath{t: t},
		&telegramWebhookStubTelegram{},
		&telegramWebhookStubAsync{enqueue: func() { enqueued = true }},
		nil,
	)
	setTelegramServiceForTest(svc, nil)

	body, err := json.Marshal(telegrambot.Update{
		Message: &telegrambot.Message{
			Chat: telegrambot.Chat{ID: 99},
			Text: "/price Opt",
		},
	})
	require.NoError(t, err)

	result, err := TelegramWebhook(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodPost,
		Body:       string(body),
		Headers: map[string]string{
			telegrambot.SecretHeader: "webhook-secret",
		},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, result.StatusCode)
	require.True(t, enqueued)
}

func TestTelegramWebhook_ForbiddenWithoutSecret(t *testing.T) {
	t.Cleanup(resetTelegramServiceForTest)
	setTelegramServiceForTest(telegrambot.NewService("webhook-secret", nil, nil, nil, nil), nil)

	result, err := TelegramWebhook(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodPost,
		Body:       `{}`,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusForbidden, result.StatusCode)
}

func TestTelegramWebhook_NotConfigured(t *testing.T) {
	t.Cleanup(resetTelegramServiceForTest)
	setTelegramServiceForTest(nil, context.Canceled)

	result, err := TelegramWebhook(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodPost,
		Body:       `{}`,
		Headers: map[string]string{
			telegrambot.SecretHeader: "webhook-secret",
		},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusServiceUnavailable, result.StatusCode)
}

func TestRunTelegramPriceRun(t *testing.T) {
	t.Cleanup(resetTelegramServiceForTest)

	var sent string
	svc := telegrambot.NewService(
		"webhook-secret",
		&telegramWebhookStubGishath{
			search: func(_ context.Context, _ string) (*telegrambot.SearchSummary, error) {
				return &telegrambot.SearchSummary{
					ResultCount: 1,
					Cheapest: &telegrambot.CardSummary{
						Name:  "Opt",
						Price: 1.5,
					},
					WebsiteURL: "https://gishathfetch.com/?s=Opt",
				}, nil
			},
		},
		&telegramWebhookStubTelegram{send: func(_ context.Context, _ int64, text, _ string) error {
			sent = text
			return nil
		}},
		nil,
		nil,
	)
	setTelegramServiceForTest(svc, nil)

	require.NoError(t, runTelegramPriceRun(context.Background(), 7, "Opt"))
	require.Contains(t, sent, "S$1.50")
}

type telegramWebhookStubGishath struct {
	t      *testing.T
	search func(context.Context, string) (*telegrambot.SearchSummary, error)
}

func (s *telegramWebhookStubGishath) Search(ctx context.Context, query string) (*telegrambot.SearchSummary, error) {
	if s.search != nil {
		return s.search(ctx, query)
	}
	if s.t != nil {
		s.t.Fatal("unexpected gishath search during webhook")
	}
	return nil, nil
}

type telegramWebhookStubTelegram struct {
	send       func(context.Context, int64, string, string) error
	sendPhoto  func(context.Context, int64, string, string) error
	forceReply func(context.Context, int64, string, string) (int64, error)
}

func (s *telegramWebhookStubTelegram) SendMessage(ctx context.Context, chatID int64, text, linkPreviewURL string) error {
	if s.send != nil {
		return s.send(ctx, chatID, text, linkPreviewURL)
	}
	return nil
}

func (s *telegramWebhookStubTelegram) SendPhoto(ctx context.Context, chatID int64, photoURL, caption string) error {
	if s.sendPhoto != nil {
		return s.sendPhoto(ctx, chatID, photoURL, caption)
	}
	return nil
}

func (s *telegramWebhookStubTelegram) SendForceReply(ctx context.Context, chatID int64, text, placeholder string) (int64, error) {
	if s.forceReply != nil {
		return s.forceReply(ctx, chatID, text, placeholder)
	}
	return 100, nil
}

type telegramWebhookStubAsync struct {
	enqueue func()
}

func (s *telegramWebhookStubAsync) EnqueuePriceSearch(ctx context.Context, chatID int64, query string) error {
	if s.enqueue != nil {
		s.enqueue()
	}
	return nil
}
