package handler

import (
	"mtg-price-checker-sg/pkg/apiauth"

	"github.com/aws/aws-lambda-go/events"
)

func turnstileTokenFromRequest(request events.APIGatewayProxyRequest) string {
	if token := headerValue(request.Headers, apiauth.TurnstileResponseHeader); token != "" {
		return token
	}
	if request.QueryStringParameters == nil {
		return ""
	}
	return request.QueryStringParameters[apiauth.TurnstileResponseQueryParam]
}
