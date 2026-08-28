package handler

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

// RouteHTTPAPI dispatches API Gateway proxy requests to handlers.
func RouteHTTPAPI(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	origin := request.Headers["origin"]
	if request.HTTPMethod == "OPTIONS" {
		return optionsResponse(origin)
	}

	switch normalizeAPIPath(request) {
	case "session":
		return Session(ctx, request)
	case "search":
		return Search(ctx, request)
	case "telegram/search":
		return TelegramSearch(ctx, request)
	case "telegram/ck":
		return TelegramCK(ctx, request)
	case "telegram/webhook":
		return TelegramWebhook(ctx, request)
	default:
		var apiRes events.APIGatewayProxyResponse
		return errorResponse(apiRes, origin, "not found", http.StatusNotFound)
	}
}
