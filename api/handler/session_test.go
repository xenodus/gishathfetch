package handler

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"mtg-price-checker-sg/pkg/apiauth"
	"mtg-price-checker-sg/pkg/config"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/require"
)

func TestSession_SameOriginWithoutOriginHeader(t *testing.T) {
	original := sessionTokenFunc
	sessionTokenFunc = func(now time.Time) (string, error) {
		return apiauth.NewSessionToken(now)
	}
	t.Cleanup(func() { sessionTokenFunc = original })

	require.NoError(t, os.Setenv(config.APISessionSecretEnv, "test-session-secret"))
	require.NoError(t, os.Setenv("ENV", config.EnvProd))
	t.Cleanup(func() {
		_ = os.Unsetenv(config.APISessionSecretEnv)
		_ = os.Unsetenv("ENV")
	})

	req := events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodGet,
		Headers:    map[string]string{},
	}

	res, err := Session(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)
	require.NotEmpty(t, res.Headers["Set-Cookie"])
	require.Contains(t, res.Headers["Set-Cookie"], "gf_api_session=")
	require.Contains(t, res.Body, `"maintenanceMode":false`)
}
