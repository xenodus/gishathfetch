package gateway

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func clearProxyEnv(t *testing.T) {
	t.Helper()
	for i := 1; i <= 7; i++ {
		t.Setenv(fmt.Sprintf("DEDICATED_PROXY_%d", i), "")
	}
	t.Setenv("DYNAMIC_PROXY", "")
}

func TestNewOutboundHTTPClient(t *testing.T) {
	clearProxyEnv(t)

	t.Run("returns direct client when no proxy is configured", func(t *testing.T) {
		client, err := NewOutboundHTTPClient(2 * time.Second)
		require.NoError(t, err)
		require.NotNil(t, client)
		require.Nil(t, client.Transport)
	})

	t.Run("returns dedicated proxy client when configured", func(t *testing.T) {
		t.Setenv("DEDICATED_PROXY_1", "1.2.3.4|8080|user|pass")
		client, err := NewOutboundHTTPClient(2 * time.Second)
		require.NoError(t, err)
		require.NotNil(t, client)
		require.NotNil(t, client.Transport)
	})
}

func TestDoOutboundGET_DirectSuccess(t *testing.T) {
	clearProxyEnv(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	resp, err := DoOutboundGET(
		context.Background(),
		server.URL,
		OutboundRequestOptions{Style: OutboundStyleJSON},
		2*time.Second,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

func TestBuildOutboundGETAttempts_TriesEachDedicatedProxy(t *testing.T) {
	clearProxyEnv(t)
	t.Setenv("DEDICATED_PROXY_1", "1.2.3.4|8080|user|pass")
	t.Setenv("DEDICATED_PROXY_2", "5.6.7.8|8080|user|pass")

	attempts := buildOutboundGETAttempts(2*time.Second, false)
	require.GreaterOrEqual(t, len(attempts), 3)
	require.Equal(t, "direct", attempts[0].strategy)
	require.Equal(t, "dedicated-1", attempts[1].strategy)
	require.Equal(t, "dedicated-2", attempts[2].strategy)
}

func TestBuildOutboundGETAttempts_SkipDirect(t *testing.T) {
	clearProxyEnv(t)
	t.Setenv("DEDICATED_PROXY_1", "1.2.3.4|8080|user|pass")

	attempts := buildOutboundGETAttempts(2*time.Second, true)
	require.Len(t, attempts, 1)
	require.Equal(t, "dedicated-1", attempts[0].strategy)
}

func TestDoOutboundGET_DirectForbiddenWithoutProxy(t *testing.T) {
	clearProxyEnv(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	defer server.Close()

	_, err := DoOutboundGET(
		context.Background(),
		server.URL,
		OutboundRequestOptions{Accept: "application/json", SkipWebBotAuth: true},
		2*time.Second,
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "direct: status 403")
}

func TestDoOutboundGET_ContinuesAfterDirectTimeout(t *testing.T) {
	clearProxyEnv(t)
	t.Setenv("DEDICATED_PROXY_1", "127.0.0.1|9|u|p")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	_, err := DoOutboundGET(
		context.Background(),
		server.URL,
		OutboundRequestOptions{SkipWebBotAuth: true},
		50*time.Millisecond,
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "direct:")
	require.Contains(t, err.Error(), "dedicated-1:")
}

func TestDoOutboundRoundTrip_RebuildsPOSTBodyPerAttempt(t *testing.T) {
	clearProxyEnv(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.Equal(t, `{"q":"test"}`, string(body))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	payload := []byte(`{"q":"test"}`)
	resp, err := DoOutboundRoundTrip(
		context.Background(),
		OutboundRequestOptions{SkipWebBotAuth: true},
		2*time.Second,
		func() (*http.Request, error) {
			return http.NewRequestWithContext(
				context.Background(),
				http.MethodPost,
				server.URL,
				bytes.NewBuffer(payload),
			)
		},
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}
