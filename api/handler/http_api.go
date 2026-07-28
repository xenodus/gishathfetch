package handler

import (
	"context"

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
	default:
		return Search(ctx, request)
	}
}
