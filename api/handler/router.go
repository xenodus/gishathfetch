package handler

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
)

// Handle routes Lambda events to the appropriate handler.
func Handle(ctx context.Context, event json.RawMessage) (any, error) {
	var internalEvent struct {
		Action string `json:"action"`
		ChatID int64  `json:"chatId"`
		Query  string `json:"query"`
	}
	if err := json.Unmarshal(event, &internalEvent); err == nil && internalEvent.Action != "" {
		switch internalEvent.Action {
		case ckPriceRefreshRunAction:
			return nil, runCKPriceRefresh(ctx)
		case analyticsKeywordsExportRunAction:
			return nil, runAnalyticsKeywordsExport(ctx)
		case telegramPriceRunAction:
			return nil, runTelegramPriceRun(ctx, internalEvent.ChatID, internalEvent.Query)
		}
	}

	var apiRequest events.APIGatewayProxyRequest
	apiRequest, err := parseAPIRequest(event)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	return RouteHTTPAPI(ctx, apiRequest)
}
