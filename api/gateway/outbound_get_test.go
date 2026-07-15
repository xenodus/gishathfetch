package gateway

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
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

func TestBuildOutboundGETAttempts_TriesOneRandomDedicatedProxy(t *testing.T) {
	clearProxyEnv(t)
	t.Setenv("DEDICATED_PROXY_1", "1.2.3.4|8080|user|pass")
	t.Setenv("DEDICATED_PROXY_2", "5.6.7.8|8080|user|pass")

	attempts := buildOutboundGETAttempts(2*time.Second, false, "")
	require.GreaterOrEqual(t, len(attempts), 2)
	require.Equal(t, "direct", attempts[0].strategy)
	require.True(t, strings.HasPrefix(attempts[1].strategy, "dedicated-"))
	require.Len(t, attempts, 2)
}

func TestBuildOutboundGETAttempts_SkipDirect(t *testing.T) {
	clearProxyEnv(t)
	t.Setenv("DEDICATED_PROXY_1", "1.2.3.4|8080|user|pass")

	attempts := buildOutboundGETAttempts(2*time.Second, true, "")
	require.Len(t, attempts, 1)
	require.True(t, strings.HasPrefix(attempts[0].strategy, "dedicated-"))
}

func TestBuildOutboundGETAttempts_OnlyProxyURL(t *testing.T) {
	clearProxyEnv(t)
	t.Setenv("DEDICATED_PROXY_1", "1.2.3.4|8080|user|pass")

	attempts := buildOutboundGETAttempts(2*time.Second, false, "http://user:pass@res.proxy:8080")
	require.Len(t, attempts, 1)
	require.Equal(t, "ck-pricelist-proxy", attempts[0].strategy)
	require.Equal(t, "http://user:pass@res.proxy:8080", attempts[0].proxyURL)
}

func TestDoOutboundGET_OnlyProxyURLSkipsDirect(t *testing.T) {
	clearProxyEnv(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	_, err := DoOutboundGET(
		context.Background(),
		server.URL,
		OutboundRequestOptions{
			Accept:         "application/json",
			SkipWebBotAuth: true,
			OnlyProxyURL:   "http://user:pass@127.0.0.1:9",
		},
		50*time.Millisecond,
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "ck-pricelist-proxy:")
	require.NotContains(t, err.Error(), "direct:")
}

func TestOutboundProxyDescription(t *testing.T) {
	clearProxyEnv(t)
	t.Setenv("DEDICATED_PROXY_1", "1.2.3.4|8080|user|pass")

	attempts := buildOutboundGETAttempts(2*time.Second, false, "")
	require.GreaterOrEqual(t, len(attempts), 2)

	require.Equal(t, "proxy_mode=direct proxy=none", outboundProxyDescription(attempts[0]))
	require.Contains(t, outboundProxyDescription(attempts[1]), "proxy_mode=dedicated proxy=DEDICATED_PROXY_")
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

func TestDoOutboundGET_Retries429BeforeFailingOver(t *testing.T) {
	clearProxyEnv(t)

	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.Header().Set("Retry-After", "0")
			http.Error(w, "rate limited", http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	resp, err := DoOutboundGET(
		context.Background(),
		server.URL,
		OutboundRequestOptions{SkipWebBotAuth: true},
		2*time.Second,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.GreaterOrEqual(t, int(calls.Load()), 2)
	require.NoError(t, resp.Body.Close())
}

func TestDoOutboundGET_FailsOverAfter429RetriesExhausted(t *testing.T) {
	clearProxyEnv(t)
	t.Setenv("DEDICATED_PROXY_1", "127.0.0.1|9|u|p")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "rate limited", http.StatusTooManyRequests)
	}))
	defer server.Close()

	_, err := DoOutboundGET(
		context.Background(),
		server.URL,
		OutboundRequestOptions{SkipWebBotAuth: true},
		2*time.Second,
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "direct: status 429")
	require.Contains(t, err.Error(), "dedicated-1:")
	require.NotContains(t, err.Error(), "dedicated-2:")
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
