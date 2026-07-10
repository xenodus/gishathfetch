package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	"mtg-price-checker-sg/pkg/adminauth"
	"mtg-price-checker-sg/pkg/config"
	"mtg-price-checker-sg/store/adminlogin"

	"github.com/aws/aws-lambda-go/events"
)

const (
	adminLoginPath   = "/admin/login"
	adminSessionPath = "/admin/session"
	adminLogoutPath  = "/admin/logout"
)

type adminLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type adminSessionResponse struct {
	Authenticated bool `json:"authenticated"`
	Enabled       bool `json:"enabled"`
}

type adminDisabledBody struct {
	Error   string `json:"error"`
	Enabled bool   `json:"enabled"`
}

var newAdminLoginService = func(ctx context.Context) (*adminlogin.Service, error) {
	store, err := adminlogin.NewDynamoDBStore(ctx)
	if err != nil {
		return nil, err
	}
	limits := config.AdminLoginRateLimitsFromEnv()
	return adminlogin.NewService(store, adminlogin.RateLimits{
		MaxFailuresPerIP:   limits.MaxFailuresPerIP,
		IPWindow:           limits.IPWindow,
		IPLockout:          limits.IPLockout,
		MaxFailuresPerUser: limits.MaxFailuresPerUser,
		UserWindow:         limits.UserWindow,
		UserLockout:        limits.UserLockout,
	}), nil
}

func routeAdmin(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	path := normalizeAdminPath(request.Path)
	origin := request.Headers["origin"]

	if request.HTTPMethod == "OPTIONS" {
		return adminOptionsResponse(origin)
	}

	switch {
	case path == adminLoginPath && request.HTTPMethod == http.MethodPost:
		return adminLogin(ctx, request, origin)
	case path == adminSessionPath && request.HTTPMethod == http.MethodGet:
		return adminSession(request, origin)
	case path == adminLogoutPath && request.HTTPMethod == http.MethodPost:
		return adminLogout(origin)
	default:
		var apiRes events.APIGatewayProxyResponse
		return adminErrorResponse(apiRes, origin, "not found", http.StatusNotFound)
	}
}

func adminLogin(ctx context.Context, request events.APIGatewayProxyRequest, origin string) (events.APIGatewayProxyResponse, error) {
	var apiRes events.APIGatewayProxyResponse
	if !config.AdminEnabled() {
		return buildAdminDisabledResponse(apiRes, origin)
	}

	loginDelay()
	now := time.Now().UTC()

	var body adminLoginRequest
	if err := json.Unmarshal([]byte(request.Body), &body); err != nil {
		return adminErrorResponse(apiRes, origin, "invalid request body", http.StatusBadRequest)
	}

	ip := clientIP(request)
	userAgent := request.Headers["user-agent"]
	service, err := newAdminLoginService(ctx)
	if err != nil {
		return adminErrorResponse(apiRes, origin, "admin login unavailable", http.StatusServiceUnavailable)
	}

	lockout, err := service.CheckLockout(ctx, ip, body.Username, now)
	if err != nil {
		return adminErrorResponse(apiRes, origin, "admin login unavailable", http.StatusServiceUnavailable)
	}
	if lockout.Locked {
		_ = service.RecordAttempt(ctx, adminlogin.Attempt{
			IP:        ip,
			Username:  body.Username,
			Success:   false,
			Blocked:   true,
			UserAgent: userAgent,
			CreatedAt: now,
		}, config.AdminAttemptLogRetention())
		return adminTooManyAttemptsResponse(apiRes, origin)
	}

	expectedUsername := os.Getenv(config.AdminUsernameEnv)
	expectedPassword := os.Getenv(config.AdminPasswordEnv)
	success := adminauth.CredentialsMatch(
		expectedUsername,
		expectedPassword,
		body.Username,
		body.Password,
	)

	if success {
		_ = service.RecordAttempt(ctx, adminlogin.Attempt{
			IP:        ip,
			Username:  body.Username,
			Success:   true,
			UserAgent: userAgent,
			CreatedAt: now,
		}, config.AdminAttemptLogRetention())
		_ = service.ClearFailures(ctx, ip, body.Username)

		sessionTTL := config.AdminSessionTTL()
		token, err := adminauth.IssueSessionToken(
			os.Getenv(config.AdminSessionSecretEnv),
			expectedUsername,
			sessionTTL,
		)
		if err != nil {
			return adminErrorResponse(apiRes, origin, "admin login unavailable", http.StatusServiceUnavailable)
		}

		response, err := jsonResponse(apiRes, origin, http.StatusOK, map[string]bool{"ok": true})
		if err != nil {
			return response, err
		}
		applyAdminResponseHeaders(&response, origin)
		response.Headers["Set-Cookie"] = adminauth.SessionCookie(token, int(sessionTTL.Seconds()))
		return response, nil
	}

	_ = service.RecordAttempt(ctx, adminlogin.Attempt{
		IP:        ip,
		Username:  body.Username,
		Success:   false,
		UserAgent: userAgent,
		CreatedAt: now,
	}, config.AdminAttemptLogRetention())
	_ = service.RecordFailure(ctx, ip, body.Username, now)

	return adminErrorResponse(apiRes, origin, "invalid credentials", http.StatusUnauthorized)
}

func adminSession(request events.APIGatewayProxyRequest, origin string) (events.APIGatewayProxyResponse, error) {
	var apiRes events.APIGatewayProxyResponse
	if !config.AdminEnabled() {
		return adminSessionJSON(apiRes, origin, adminSessionResponse{
			Authenticated: false,
			Enabled:       false,
		})
	}

	token := adminauth.SessionTokenFromCookie(request.Headers["cookie"])
	if token == "" {
		return adminSessionJSON(apiRes, origin, adminSessionResponse{
			Authenticated: false,
			Enabled:       true,
		})
	}

	_, err := adminauth.ValidateSessionToken(os.Getenv(config.AdminSessionSecretEnv), token)
	if err != nil {
		return adminSessionJSON(apiRes, origin, adminSessionResponse{
			Authenticated: false,
			Enabled:       true,
		})
	}

	return adminSessionJSON(apiRes, origin, adminSessionResponse{
		Authenticated: true,
		Enabled:       true,
	})
}

func adminLogout(origin string) (events.APIGatewayProxyResponse, error) {
	var apiRes events.APIGatewayProxyResponse
	if !config.AdminEnabled() {
		return buildAdminDisabledResponse(apiRes, origin)
	}

	response, err := jsonResponse(apiRes, origin, http.StatusOK, map[string]bool{"ok": true})
	if err != nil {
		return response, err
	}
	applyAdminResponseHeaders(&response, origin)
	response.Headers["Set-Cookie"] = adminauth.ClearSessionCookie()
	return response, nil
}

func buildAdminDisabledResponse(apiRes events.APIGatewayProxyResponse, origin string) (events.APIGatewayProxyResponse, error) {
	response, err := jsonResponse(apiRes, origin, http.StatusServiceUnavailable, adminDisabledBody{
		Error:   "admin is not configured",
		Enabled: false,
	})
	if err != nil {
		return response, err
	}
	applyAdminResponseHeaders(&response, origin)
	return response, nil
}

func adminTooManyAttemptsResponse(apiRes events.APIGatewayProxyResponse, origin string) (events.APIGatewayProxyResponse, error) {
	response, err := jsonResponse(apiRes, origin, http.StatusTooManyRequests, ErrorResponse{
		Error:      "too many login attempts, try again later",
		StatusCode: http.StatusTooManyRequests,
	})
	if err != nil {
		return response, err
	}
	applyAdminResponseHeaders(&response, origin)
	return response, nil
}

func adminErrorResponse(
	apiRes events.APIGatewayProxyResponse,
	origin string,
	message string,
	statusCode int,
) (events.APIGatewayProxyResponse, error) {
	response, err := errorResponse(apiRes, origin, message, statusCode)
	if err != nil {
		return response, err
	}
	applyAdminResponseHeaders(&response, origin)
	return response, nil
}

func adminSessionJSON(
	apiRes events.APIGatewayProxyResponse,
	origin string,
	payload adminSessionResponse,
) (events.APIGatewayProxyResponse, error) {
	response, err := jsonResponse(apiRes, origin, http.StatusOK, payload)
	if err != nil {
		return response, err
	}
	applyAdminResponseHeaders(&response, origin)
	return response, nil
}

func normalizeAdminPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return strings.TrimSuffix(path, "/")
}

func isAdminPath(path string) bool {
	normalized := normalizeAdminPath(path)
	return normalized == adminLoginPath ||
		normalized == adminSessionPath ||
		normalized == adminLogoutPath
}

func clientIP(request events.APIGatewayProxyRequest) string {
	if ip := strings.TrimSpace(request.RequestContext.Identity.SourceIP); ip != "" {
		return ip
	}
	if xff := request.Headers["x-forwarded-for"]; xff != "" {
		return strings.TrimSpace(strings.Split(xff, ",")[0])
	}
	return "unknown"
}

func loginDelay() {
	loginDelayFunc()
}

var loginDelayFunc = func() {
	jitter := time.Duration(time.Now().UnixNano()%200) * time.Millisecond
	time.Sleep(200*time.Millisecond + jitter)
}

func applyAdminResponseHeaders(apiResponse *events.APIGatewayProxyResponse, origin string) {
	applyAdminCORSHeaders(apiResponse, origin)
	if apiResponse.Headers == nil {
		apiResponse.Headers = map[string]string{}
	}
	apiResponse.Headers["Cache-Control"] = "no-store"
}

func adminOptionsResponse(origin string) (events.APIGatewayProxyResponse, error) {
	apiResponse := events.APIGatewayProxyResponse{StatusCode: http.StatusNoContent}
	applyAdminCORSHeaders(&apiResponse, origin)
	return apiResponse, nil
}

func applyAdminCORSHeaders(apiResponse *events.APIGatewayProxyResponse, origin string) {
	if !isAllowedOrigin(origin) {
		return
	}
	apiResponse.Headers = map[string]string{
		"Access-Control-Allow-Origin":      origin,
		"Access-Control-Allow-Methods":     "GET, POST, OPTIONS",
		"Access-Control-Allow-Headers":     "Content-Type, Cookie",
		"Access-Control-Allow-Credentials": "true",
		"Vary":                             "Origin",
	}
}

func isAllowedOrigin(origin string) bool {
	return slices.Contains(config.GetAllowedOrigins(), origin)
}
