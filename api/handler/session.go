package handler

import (
	"context"
	"errors"
	"net/http"
	"os"
	"time"

	"mtg-price-checker-sg/pkg/apiauth"
	"mtg-price-checker-sg/pkg/config"

	"github.com/aws/aws-lambda-go/events"
)

var sessionTokenFunc = apiauth.NewSessionToken

var verifyTurnstileFunc = apiauth.VerifyTurnstile

// Session mints an HttpOnly cookie the browser must send before search requests.
func Session(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var apiRes events.APIGatewayProxyResponse
	origin := request.Headers["origin"]

	if request.HTTPMethod == "OPTIONS" {
		return optionsResponse(origin)
	}

	if request.HTTPMethod != http.MethodGet {
		return errorResponse(apiRes, origin, "method not allowed", http.StatusMethodNotAllowed)
	}

	if res, ok := enforceOriginVerify(apiRes, origin, request.Headers); !ok {
		return res, nil
	}

	if config.APISessionSecret() == "" {
		return errorResponse(apiRes, origin, "session not configured", http.StatusServiceUnavailable)
	}

	turnstileToken := turnstileTokenFromRequest(request)
	remoteIP := request.RequestContext.Identity.SourceIP
	if err := verifyTurnstileFunc(ctx, turnstileToken, remoteIP); err != nil {
		message := "verification failed"
		if errors.Is(err, apiauth.ErrTurnstileTokenMissing) {
			message = "verification required"
		}
		return accessDeniedResponse(apiRes, origin, message), nil
	}

	token, err := sessionTokenFunc(time.Now().UTC())
	if err != nil {
		return errorResponse(apiRes, origin, "err creating session", http.StatusInternalServerError)
	}

	secure := os.Getenv("ENV") == config.EnvProd
	apiRes.StatusCode = http.StatusNoContent
	applyCORSHeaders(&apiRes, origin)
	headers := ensureResponseHeaders(&apiRes)
	headers["Set-Cookie"] = sessionCookieString(token, secure)
	headers["Cache-Control"] = "no-store"
	return apiRes, nil
}
