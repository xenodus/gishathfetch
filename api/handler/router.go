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
	}
	if err := json.Unmarshal(event, &internalEvent); err == nil && internalEvent.Action != "" {
		switch internalEvent.Action {
		case ckPriceRefreshRunAction:
			return nil, runCKPriceRefresh(ctx)
		case analyticsKeywordsExportRunAction:
			return nil, runAnalyticsKeywordsExport(ctx)
		}
	}

	var apiRequest events.APIGatewayProxyRequest
	if err := json.Unmarshal(event, &apiRequest); err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	if isAdminPath(apiRequest.Path) {
		return routeAdmin(ctx, apiRequest)
	}

	return Search(ctx, apiRequest)
}
