package apiauth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"mtg-price-checker-sg/pkg/config"

	"github.com/stretchr/testify/require"
)

func TestVerifyTurnstile_SkipsWhenSecretUnset(t *testing.T) {
	t.Setenv(config.TurnstileSecretKeyEnv, "")
	require.NoError(t, VerifyTurnstile(context.Background(), "", ""))
}

func TestVerifyTurnstile_RequiresTokenWhenSecretSet(t *testing.T) {
	t.Setenv(config.TurnstileSecretKeyEnv, "test-secret")

	err := VerifyTurnstile(context.Background(), "", "203.0.113.1")
	require.ErrorIs(t, err, ErrTurnstileTokenMissing)
}

func TestVerifyTurnstile_AcceptsSuccessfulSiteverify(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.NoError(t, r.ParseForm())
		require.Equal(t, "test-secret", r.Form.Get("secret"))
		require.Equal(t, "good-token", r.Form.Get("response"))
		require.Equal(t, "203.0.113.1", r.Form.Get("remoteip"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true}`))
	}))
	t.Cleanup(server.Close)

	originalURL := turnstileSiteverifyURL
	turnstileSiteverifyURL = server.URL
	t.Cleanup(func() { turnstileSiteverifyURL = originalURL })

	t.Setenv(config.TurnstileSecretKeyEnv, "test-secret")

	err := VerifyTurnstile(context.Background(), "good-token", "203.0.113.1")
	require.NoError(t, err)
}

func TestVerifyTurnstile_RejectsFailedSiteverify(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"success":false,"error-codes":["invalid-input-response"]}`))
	}))
	t.Cleanup(server.Close)

	originalURL := turnstileSiteverifyURL
	turnstileSiteverifyURL = server.URL
	t.Cleanup(func() { turnstileSiteverifyURL = originalURL })

	t.Setenv(config.TurnstileSecretKeyEnv, "test-secret")

	err := VerifyTurnstile(context.Background(), "bad-token", "")
	require.ErrorIs(t, err, ErrTurnstileVerificationFailed)
}
