package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"mtg-price-checker-sg/pkg/apiauth"
	"mtg-price-checker-sg/pkg/config"

	"github.com/aws/aws-lambda-go/events"
)

var sessionTokenFunc = apiauth.NewSessionToken
var turnstileVerifyFunc = apiauth.VerifyTurnstileToken

type sessionRequestBody struct {
	TurnstileToken string `json:"turnstileToken"`
}

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
	if config.TurnstileSecretKey() == "" {
		if request.HTTPMethod != http.MethodGet {
			return "", errSessionMethodNotAllowed
		}
		return "", nil
	}

	if request.HTTPMethod != http.MethodPost {
		return "", errSessionVerificationRequired
	}

	var body sessionRequestBody
	if strings.TrimSpace(request.Body) == "" {
		return "", errSessionVerificationRequired
	}
	if err := json.Unmarshal([]byte(request.Body), &body); err != nil {
		return "", errSessionVerificationRequired
	}
	if strings.TrimSpace(body.TurnstileToken) == "" {
		return "", errSessionVerificationRequired
	}
	return strings.TrimSpace(body.TurnstileToken), nil
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
