package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"mtg-price-checker-sg/pkg/apiauth"
	"mtg-price-checker-sg/pkg/config"

	"github.com/aws/aws-lambda-go/events"
)

func enforceOriginVerify(
	apiRes events.APIGatewayProxyResponse,
	origin string,
	headers map[string]string,
	domainName string,
) (events.APIGatewayProxyResponse, bool) {
	if err := apiauth.VerifyOriginHeader(headers, domainName); err != nil {
		return accessDeniedResponse(apiRes, origin, "forbidden"), false
	}
	return apiRes, true
}

func enforceSession(
	apiRes events.APIGatewayProxyResponse,
	origin string,
	headers map[string]string,
) (events.APIGatewayProxyResponse, bool) {
	if config.APISessionSecret() == "" {
		return apiRes, true
	}

	cookieHeader := headerValue(headers, "cookie")
	token := cookieValue(cookieHeader, apiauth.SessionCookieName())
	if err := apiauth.ValidateSessionToken(token, time.Now().UTC()); err != nil {
		message := "session required"
		if errors.Is(err, apiauth.ErrSessionExpired) {
			message = "session expired"
		}
		return accessDeniedResponse(apiRes, origin, message), false
	}

	return apiRes, true
}

func accessDeniedResponse(
	apiRes events.APIGatewayProxyResponse,
	origin string,
	message string,
) events.APIGatewayProxyResponse {
	return mustErrorResponse(apiRes, origin, message, http.StatusForbidden)
}

func mustErrorResponse(
	apiRes events.APIGatewayProxyResponse,
	origin string,
	message string,
	statusCode int,
) events.APIGatewayProxyResponse {
	res, err := errorResponse(apiRes, origin, message, statusCode)
	if err != nil {
		res.StatusCode = http.StatusInternalServerError
		res.Body = `{"error":"internal error","statusCode":500}` + "\n"
	}
	return res
}

func headerValue(headers map[string]string, lowerName string) string {
	if headers == nil {
		return ""
	}
	if v := headers[lowerName]; v != "" {
		return v
	}
	for k, v := range headers {
		if strings.EqualFold(k, lowerName) {
			return v
		}
	}
	return ""
}

func cookieValue(cookieHeader, name string) string {
	if cookieHeader == "" {
		return ""
	}
	for part := range strings.SplitSeq(cookieHeader, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		key, value, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(key), name) {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func sessionCookieString(token string, secure bool) string {
	maxAge := int(config.APISessionTTL().Seconds())
	parts := []string{
		fmt.Sprintf("%s=%s", apiauth.SessionCookieName(), token),
		"Path=/",
		fmt.Sprintf("Max-Age=%d", maxAge),
		"HttpOnly",
		"SameSite=Lax",
	}
	if secure {
		parts = append(parts, "Secure")
	}
	return strings.Join(parts, "; ")
}
