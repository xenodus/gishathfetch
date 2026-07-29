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

func TestSession_TurnstileRequiredWhenConfigured(t *testing.T) {
	originalVerify := verifyTurnstileFunc
	verifyTurnstileFunc = func(_ context.Context, token, _ string) error {
		if token == "" {
			return apiauth.ErrTurnstileTokenMissing
		}
		return nil
	}
	t.Cleanup(func() { verifyTurnstileFunc = originalVerify })

	require.NoError(t, os.Setenv(config.APISessionSecretEnv, "test-session-secret"))
	require.NoError(t, os.Setenv(config.TurnstileSecretKeyEnv, "turnstile-secret"))
	t.Cleanup(func() {
		_ = os.Unsetenv(config.APISessionSecretEnv)
		_ = os.Unsetenv(config.TurnstileSecretKeyEnv)
	})

	req := events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodGet,
		Headers:    map[string]string{},
	}

	res, err := Session(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, http.StatusForbidden, res.StatusCode)

	req.Headers = map[string]string{
		apiauth.TurnstileResponseHeader: "valid-token",
	}
	res, err = Session(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, res.StatusCode)

	req = events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodGet,
		QueryStringParameters: map[string]string{
			apiauth.TurnstileResponseQueryParam: "valid-token",
		},
	}
	res, err = Session(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, res.StatusCode)
}
