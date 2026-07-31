package apiauth

import (
	"os"
	"testing"
	"time"

	"mtg-price-checker-sg/pkg/config"

	"github.com/stretchr/testify/require"
)

func TestSessionTokenRoundTrip(t *testing.T) {
	require.NoError(t, os.Setenv(config.APISessionSecretEnv, "test-session-secret"))
	t.Cleanup(func() { _ = os.Unsetenv(config.APISessionSecretEnv) })

	now := time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC)
	token, err := NewSessionToken(now)
	require.NoError(t, err)
	require.NoError(t, ValidateSessionToken(token, now.Add(time.Minute)))
}

func TestSessionTokenExpired(t *testing.T) {
	require.NoError(t, os.Setenv(config.APISessionSecretEnv, "test-session-secret"))
	t.Cleanup(func() { _ = os.Unsetenv(config.APISessionSecretEnv) })

	now := time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC)
	token, err := NewSessionToken(now)
	require.NoError(t, err)

	err = ValidateSessionToken(token, now.Add(config.APISessionTTL()+time.Second))
	require.ErrorIs(t, err, ErrSessionExpired)
}

func TestVerifyOriginHeader(t *testing.T) {
	require.NoError(t, os.Setenv(config.APIOriginVerifySecretEnv, "cloudfront-secret"))
	t.Cleanup(func() { _ = os.Unsetenv(config.APIOriginVerifySecretEnv) })

	err := VerifyOriginHeader(map[string]string{"x-origin-verify": "cloudfront-secret"}, "")
	require.NoError(t, err)

	err = VerifyOriginHeader(map[string]string{"x-origin-verify": "wrong"}, "")
	require.ErrorIs(t, err, errOriginVerifyFailed)

	err = VerifyOriginHeader(
		map[string]string{"origin": "https://gishathfetch.com"},
		"api.gishathfetch.com",
	)
	require.NoError(t, err)

	err = VerifyOriginHeader(
		map[string]string{"origin": "https://gishathfetch.com"},
		"abc123.execute-api.ap-southeast-1.amazonaws.com",
	)
	require.ErrorIs(t, err, errOriginVerifyFailed)

	err = VerifyOriginHeader(map[string]string{}, "api.gishathfetch.com")
	require.ErrorIs(t, err, errOriginVerifyFailed)
}
