package telegrambot

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTelegramAPI_SetMyCommands(t *testing.T) {
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Contains(t, r.URL.Path, "/bottest-token/setMyCommands")
		raw, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(raw, &gotBody))
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(server.Close)

	api := NewTelegramAPI("test-token")
	api.httpClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		req.URL.Scheme = "http"
		req.URL.Host = server.Listener.Addr().String()
		return http.DefaultTransport.RoundTrip(req)
	})}

	commands := DefaultBotCommands()
	require.NoError(t, api.SetMyCommands(context.Background(), commands))

	rawCommands, ok := gotBody["commands"].([]any)
	require.True(t, ok)
	require.Len(t, rawCommands, len(commands))
}

func TestRegisterBot_RegistersCommandsAndWebhook(t *testing.T) {
	var methods []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.URL.Path)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(server.Close)

	api := NewTelegramAPI("test-token")
	api.httpClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		req.URL.Scheme = "http"
		req.URL.Host = server.Listener.Addr().String()
		return http.DefaultTransport.RoundTrip(req)
	})}

	cfg := Config{
		WebhookPublicURL: "https://bot.example/telegram/webhook",
		WebhookSecret:    "secret",
	}
	require.NoError(t, RegisterBot(context.Background(), cfg, api))
	require.Len(t, methods, 2)
	require.Contains(t, methods[0], "/setMyCommands")
	require.Contains(t, methods[1], "/setWebhook")
}

func TestRegisterBot_SkipsWebhookWhenUnset(t *testing.T) {
	var methods []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.URL.Path)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(server.Close)

	api := NewTelegramAPI("test-token")
	api.httpClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		req.URL.Scheme = "http"
		req.URL.Host = server.Listener.Addr().String()
		return http.DefaultTransport.RoundTrip(req)
	})}

	require.NoError(t, RegisterBot(context.Background(), Config{}, api))
	require.Len(t, methods, 1)
	require.Contains(t, methods[0], "/setMyCommands")
}
