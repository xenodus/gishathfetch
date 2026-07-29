package handler

import (
	"context"
	"net/http"
	"os"
	"testing"

	"mtg-price-checker-sg/pkg/apiauth"
	"mtg-price-checker-sg/pkg/config"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/require"
)

func TestTurnstileTokenFromRequest_PrefersHeader(t *testing.T) {
	req := events.APIGatewayProxyRequest{
		Headers: map[string]string{
			apiauth.TurnstileResponseHeader: "from-header",
		},
		QueryStringParameters: map[string]string{
			apiauth.TurnstileResponseQueryParam: "from-query",
		},
	}
	require.Equal(t, "from-header", turnstileTokenFromRequest(req))
}

func TestTurnstileTokenFromRequest_QueryParam(t *testing.T) {
	req := events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodGet,
		QueryStringParameters: map[string]string{
			apiauth.TurnstileResponseQueryParam: "from-query",
		},
	}
	require.Equal(t, "from-query", turnstileTokenFromRequest(req))
}

func TestSession_TurnstileQueryParamAccepted(t *testing.T) {
	originalVerify := verifyTurnstileFunc
	verifyTurnstileFunc = func(_ context.Context, token string) error {
		if token != "query-token" {
			return apiauth.ErrTurnstileTokenMissing
		}
		return nil
	}
	t.Cleanup(func() { verifyTurnstileFunc = originalVerify })

	require.NoError(t, os.Setenv(config.APISessionSecretEnv, "test-session-secret"))
	t.Cleanup(func() { _ = os.Unsetenv(config.APISessionSecretEnv) })

	req := events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodGet,
		QueryStringParameters: map[string]string{
			apiauth.TurnstileResponseQueryParam: "query-token",
		},
	}
	res, err := Session(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, res.StatusCode)
}
