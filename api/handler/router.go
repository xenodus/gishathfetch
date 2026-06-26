package handler

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

// Handle routes Lambda events to the appropriate handler.
func Handle(ctx context.Context, event json.RawMessage) (any, error) {
	var cloudWatchEvent events.CloudWatchEvent
	if err := json.Unmarshal(event, &cloudWatchEvent); err == nil && cloudWatchEvent.Source == "aws.events" {
		return nil, RefreshCKPrices(ctx, cloudWatchEvent)
	}

	var apiRequest events.APIGatewayProxyRequest
	if err := json.Unmarshal(event, &apiRequest); err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	path := strings.TrimSuffix(strings.ToLower(apiRequest.Path), "/")
	if path == "/ck-price" || strings.HasSuffix(path, "/ck-price") {
		return CKPrice(ctx, apiRequest)
	}

	return Search(ctx, apiRequest)
}
