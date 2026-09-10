package handler

import (
	"context"
	"encoding/json"
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

func TestSession_RequiresTurnstileWhenConfigured(t *testing.T) {
	t.Setenv(config.APISessionSecretEnv, "test-session-secret")
	t.Setenv(config.TurnstileSecretKeyEnv, "test-turnstile-secret")

	req := events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodGet,
		Headers: map[string]string{
			"origin": "http://localhost:5173",
		},
	}

	res, err := Session(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, res.StatusCode)

	var body ErrorResponse
	require.NoError(t, json.Unmarshal([]byte(res.Body), &body))
	require.Equal(t, "verification required", body.Error)
}

func TestSession_MintsWithVerifiedTurnstileToken(t *testing.T) {
	originalVerify := turnstileVerifyFunc
	turnstileVerifyFunc = func(ctx context.Context, token, remoteIP string) error {
		require.Equal(t, "good-token", token)
		require.Equal(t, "203.0.113.1", remoteIP)
		return nil
	}
	t.Cleanup(func() { turnstileVerifyFunc = originalVerify })

	t.Setenv(config.APISessionSecretEnv, "test-session-secret")
	t.Setenv(config.TurnstileSecretKeyEnv, "test-turnstile-secret")

	body, err := json.Marshal(sessionRequestBody{TurnstileToken: "good-token"})
	require.NoError(t, err)

	req := events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodPost,
		Headers: map[string]string{
			"origin": "http://localhost:5173",
		},
		Body: string(body),
		RequestContext: events.APIGatewayProxyRequestContext{
			Identity: events.APIGatewayRequestIdentity{
				SourceIP: "203.0.113.1",
			},
		},
	}

	res, err := Session(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Contains(t, res.Headers["Set-Cookie"], "gf_api_session=")
}

func TestSession_RejectsFailedTurnstileVerification(t *testing.T) {
	originalVerify := turnstileVerifyFunc
	turnstileVerifyFunc = func(ctx context.Context, token, remoteIP string) error {
		return apiauth.ErrTurnstileVerificationFailed
	}
	t.Cleanup(func() { turnstileVerifyFunc = originalVerify })

	t.Setenv(config.APISessionSecretEnv, "test-session-secret")
	t.Setenv(config.TurnstileSecretKeyEnv, "test-turnstile-secret")

	body, err := json.Marshal(sessionRequestBody{TurnstileToken: "bad-token"})
	require.NoError(t, err)

	req := events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodPost,
		Headers: map[string]string{
			"origin": "http://localhost:5173",
		},
		Body: string(body),
	}

	res, err := Session(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, http.StatusForbidden, res.StatusCode)

	var payload ErrorResponse
	require.NoError(t, json.Unmarshal([]byte(res.Body), &payload))
	require.Equal(t, "verification failed", payload.Error)
}
