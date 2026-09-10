package handler

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"mtg-price-checker-sg/pkg/apiauth"
	"mtg-price-checker-sg/pkg/config"

	"github.com/aws/aws-lambda-go/events"
)

// TurnstileTokenQueryParam carries a one-time Cloudflare Turnstile response on GET /session.
// Query string avoids a CORS preflight for a custom header; API Gateway OPTIONS does not
// forward preflights to Lambda with our current CORS wiring.
// TODO(api-abuse): migrate to POST /session with a JSON body once API Gateway exposes POST.
const TurnstileTokenQueryParam = "turnstileToken"

var sessionTokenFunc = apiauth.NewSessionToken
var turnstileVerifyFunc = apiauth.VerifyTurnstileToken

// Session mints an HttpOnly cookie the browser must send before search requests.
func Session(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var apiRes events.APIGatewayProxyResponse
	origin := request.Headers["origin"]

	if request.HTTPMethod == "OPTIONS" {
		return optionsResponse(origin)
	}

	if res, ok := enforceOriginVerify(apiRes, origin, request.Headers); !ok {
		return res, nil
	}

	if config.APISessionSecret() == "" {
		return errorResponse(apiRes, origin, "session not configured", http.StatusServiceUnavailable)
	}

	turnstileToken, err := parseSessionTurnstileToken(request)
	if err != nil {
		return errorResponse(apiRes, origin, err.Error(), http.StatusBadRequest)
	}
	if res, ok := enforceTurnstile(ctx, apiRes, origin, turnstileToken, request); !ok {
		return res, nil
	}

	token, err := sessionTokenFunc(time.Now().UTC())
	if err != nil {
		return errorResponse(apiRes, origin, "err creating session", http.StatusInternalServerError)
	}

	secure := os.Getenv("ENV") == config.EnvProd
	apiRes, err = jsonResponse(apiRes, origin, http.StatusOK, buildSiteStatusResponse())
	if err != nil {
		return errorResponse(apiRes, origin, "err marshalling response", http.StatusInternalServerError)
	}
	applyMaintenanceHeaders(&apiRes)
	headers := ensureResponseHeaders(&apiRes)
	headers["Set-Cookie"] = sessionCookieString(token, secure)
	headers["Cache-Control"] = "no-store"
	return apiRes, nil
}

func parseSessionTurnstileToken(request events.APIGatewayProxyRequest) (string, error) {
	if request.HTTPMethod != http.MethodGet {
		return "", errSessionMethodNotAllowed
	}

	if config.TurnstileSecretKey() == "" {
		return "", nil
	}

	if request.QueryStringParameters != nil {
		token := strings.TrimSpace(request.QueryStringParameters[TurnstileTokenQueryParam])
		if token != "" {
			return token, nil
		}
	}
	return "", errSessionVerificationRequired
}

func enforceTurnstile(
	ctx context.Context,
	apiRes events.APIGatewayProxyResponse,
	origin string,
	turnstileToken string,
	request events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, bool) {
	if config.TurnstileSecretKey() == "" {
		return apiRes, true
	}

	if err := turnstileVerifyFunc(ctx, turnstileToken, request.RequestContext.Identity.SourceIP); err != nil {
		if errors.Is(err, apiauth.ErrTurnstileVerificationFailed) {
			return accessDeniedResponse(apiRes, origin, "verification failed"), false
		}
		res, _ := errorResponse(apiRes, origin, "verification unavailable", http.StatusServiceUnavailable)
		return res, false
	}

	return apiRes, true
}

var (
	errSessionVerificationRequired = errors.New("verification required")
	errSessionMethodNotAllowed     = errors.New("method not allowed")
)
