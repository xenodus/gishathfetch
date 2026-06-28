package handler

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

// Handle routes Lambda events to the appropriate handler.
func Handle(ctx context.Context, event json.RawMessage) (any, error) {
	var internalEvent struct {
		Action string `json:"action"`
	}
	if err := json.Unmarshal(event, &internalEvent); err == nil && internalEvent.Action == ckPriceRefreshRunAction {
		return nil, runCKPriceRefresh(ctx)
	}

	var apiRequest events.APIGatewayProxyRequest
	if err := json.Unmarshal(event, &apiRequest); err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	path := strings.TrimSuffix(strings.ToLower(apiRequest.Path), "/")
	if path == "/ck-price/refresh" || strings.HasSuffix(path, "/ck-price/refresh") {
		return CKPriceRefresh(ctx, apiRequest)
	}

	return Search(ctx, apiRequest)
}
