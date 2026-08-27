package handler

import (
	"context"
	"net/http"
	"sync"

	"mtg-price-checker-sg/pkg/telegrambot"

	"github.com/aws/aws-lambda-go/events"
)

var (
	telegramServiceMu  sync.RWMutex
	telegramService    *telegrambot.Service
	telegramServiceErr error
)

func getTelegramService(ctx context.Context) (*telegrambot.Service, error) {
	telegramServiceMu.RLock()
	if telegramService != nil || telegramServiceErr != nil {
		svc, err := telegramService, telegramServiceErr
		telegramServiceMu.RUnlock()
		return svc, err
	}
	telegramServiceMu.RUnlock()

	telegramServiceMu.Lock()
	defer telegramServiceMu.Unlock()
	if telegramService == nil && telegramServiceErr == nil {
		telegramService, telegramServiceErr = telegrambot.NewServiceFromConfig(ctx, telegrambot.LoadConfig(), nil)
	}
	return telegramService, telegramServiceErr
}

func setTelegramServiceForTest(svc *telegrambot.Service, err error) {
	telegramServiceMu.Lock()
	defer telegramServiceMu.Unlock()
	telegramService = svc
	telegramServiceErr = err
}

func resetTelegramServiceForTest() {
	telegramServiceMu.Lock()
	defer telegramServiceMu.Unlock()
	telegramService = nil
	telegramServiceErr = nil
}

// TelegramWebhook accepts Telegram webhook POSTs and enqueues async price searches.
func TelegramWebhook(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var apiRes events.APIGatewayProxyResponse
	origin := request.Headers["origin"]

	if request.HTTPMethod == "OPTIONS" {
		return optionsResponse(origin)
	}

	if request.HTTPMethod != http.MethodPost {
		return errorResponse(apiRes, origin, "method not allowed", http.StatusMethodNotAllowed)
	}

	svc, err := getTelegramService(ctx)
	if err != nil {
		return errorResponse(apiRes, origin, "telegram bot not configured", http.StatusServiceUnavailable)
	}

	secret := headerValue(request.Headers, telegrambot.SecretHeader)
	status := svc.HandleWebhook(ctx, secret, []byte(request.Body))
	apiRes.StatusCode = status
	return apiRes, nil
}
