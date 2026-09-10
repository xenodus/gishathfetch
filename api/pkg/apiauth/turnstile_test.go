package apiauth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"mtg-price-checker-sg/pkg/config"

	"github.com/stretchr/testify/require"
)

func TestVerifyTurnstileToken_SkipsWhenSecretUnset(t *testing.T) {
	require.NoError(t, VerifyTurnstileToken(context.Background(), "", ""))
}

func TestVerifyTurnstileToken_RejectsEmptyToken(t *testing.T) {
	t.Setenv(config.TurnstileSecretKeyEnv, "test-turnstile-secret")

	err := VerifyTurnstileToken(context.Background(), "", "")
	require.ErrorIs(t, err, ErrTurnstileVerificationFailed)
}

func TestVerifyTurnstileToken_AcceptsSuccessfulSiteVerify(t *testing.T) {
	t.Setenv(config.TurnstileSecretKeyEnv, "test-turnstile-secret")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
		require.NoError(t, r.ParseForm())
		require.Equal(t, "test-turnstile-secret", r.FormValue("secret"))
		require.Equal(t, "good-token", r.FormValue("response"))
		require.Equal(t, "203.0.113.1", r.FormValue("remoteip"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true}`))
	}))
	t.Cleanup(server.Close)

	originalClient := turnstileHTTPClient
	turnstileHTTPClient = server.Client()
	t.Cleanup(func() { turnstileHTTPClient = originalClient })

	originalURL := turnstileSiteVerifyURL
	turnstileSiteVerifyURL = server.URL
	t.Cleanup(func() { turnstileSiteVerifyURL = originalURL })

	err := VerifyTurnstileToken(context.Background(), "good-token", "203.0.113.1")
	require.NoError(t, err)
}

func TestVerifyTurnstileToken_RejectsFailedSiteVerify(t *testing.T) {
	t.Setenv(config.TurnstileSecretKeyEnv, "test-turnstile-secret")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":false}`))
	}))
	t.Cleanup(server.Close)

	originalClient := turnstileHTTPClient
	turnstileHTTPClient = server.Client()
	t.Cleanup(func() { turnstileHTTPClient = originalClient })

	originalURL := turnstileSiteVerifyURL
	turnstileSiteVerifyURL = server.URL
	t.Cleanup(func() { turnstileSiteVerifyURL = originalURL })

	err := VerifyTurnstileToken(context.Background(), "bad-token", "")
	require.ErrorIs(t, err, ErrTurnstileVerificationFailed)
}
