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

func TestTelegramAPI_SendForceReply(t *testing.T) {
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Contains(t, r.URL.Path, "/bottest-token/sendMessage")
		raw, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(raw, &gotBody))
		_, _ = w.Write([]byte(`{"ok":true,"result":{"message_id":100}}`))
	}))
	t.Cleanup(server.Close)

	api := NewTelegramAPI("test-token")
	api.httpClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		req.URL.Scheme = "http"
		req.URL.Host = server.Listener.Addr().String()
		return http.DefaultTransport.RoundTrip(req)
	})}

	messageID, err := api.SendForceReply(context.Background(), 42, pricePromptMessage, pricePromptPlaceholder)
	require.NoError(t, err)
	require.Equal(t, int64(100), messageID)

	require.Equal(t, float64(42), gotBody["chat_id"])
	require.Equal(t, pricePromptMessage, gotBody["text"])
	markup, ok := gotBody["reply_markup"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, true, markup["force_reply"])
	require.Equal(t, pricePromptPlaceholder, markup["input_field_placeholder"])
	_, hasSelective := markup["selective"]
	require.False(t, hasSelective)
}

func TestTelegramAPI_SendMessage_LinkPreviewOptions(t *testing.T) {
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Contains(t, r.URL.Path, "/bottest-token/sendMessage")
		raw, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(raw, &gotBody))
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(server.Close)

	api := NewTelegramAPI("test-token")
	api.httpClient = server.Client()
	// Point requests at the test server by temporarily replacing post via custom transport host rewrite.
	api.httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		req.URL.Scheme = "http"
		req.URL.Host = server.Listener.Addr().String()
		return http.DefaultTransport.RoundTrip(req)
	})

	preview := "https://gishathfetch.com/?s=Opt"
	err := api.SendMessage(context.Background(), 42, "buy https://shop.example\n"+preview, preview)
	require.NoError(t, err)

	require.Equal(t, float64(42), gotBody["chat_id"])
	opts, ok := gotBody["link_preview_options"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, preview, opts["url"])
	_, hasDisable := gotBody["disable_web_page_preview"]
	require.False(t, hasDisable)
}

func TestTelegramAPI_SendMessage_DefaultPreview(t *testing.T) {
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	require.NoError(t, api.SendMessage(context.Background(), 1, "Searching…", ""))
	require.Equal(t, true, gotBody["disable_web_page_preview"])
	_, hasOpts := gotBody["link_preview_options"]
	require.False(t, hasOpts)
}

func TestTelegramAPI_SendPhoto(t *testing.T) {
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Contains(t, r.URL.Path, "/bottest-token/sendPhoto")
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

	photoURL := "https://product-images.tcgplayer.com/fit-in/437x437/12345.jpg"
	caption := "Opt -> S$1.25 @ Hideout"
	require.NoError(t, api.SendPhoto(context.Background(), 42, photoURL, caption))

	require.Equal(t, float64(42), gotBody["chat_id"])
	require.Equal(t, photoURL, gotBody["photo"])
	require.Equal(t, caption, gotBody["caption"])
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
