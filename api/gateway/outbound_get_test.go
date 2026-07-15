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
	t.Setenv("RESIDENTIAL_PROXY_1", "")
}

func TestNewOutboundHTTPClient(t *testing.T) {
	clearProxyEnv(t)

	t.Run("returns direct client when no proxy is configured", func(t *testing.T) {
		t.Setenv("BROWSER_TLS_EMULATION_ENABLED", "false")
		client, err := NewOutboundHTTPClient(2 * time.Second)
		require.NoError(t, err)
		require.NotNil(t, client)
		require.Nil(t, client.Transport)
	})

	t.Run("returns dedicated proxy client when configured", func(t *testing.T) {
		t.Setenv("BROWSER_TLS_EMULATION_ENABLED", "false")
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

	attempts := buildOutboundGETAttempts(context.Background(), 2*time.Second, OutboundRequestOptions{})
	require.GreaterOrEqual(t, len(attempts), 2)
	require.Equal(t, "direct", attempts[0].strategy)
	require.True(t, strings.HasPrefix(attempts[1].strategy, "dedicated-"))
	require.Len(t, attempts, 2)
}

func TestBuildOutboundGETAttempts_DoesNotTryAllSevenDedicatedProxies(t *testing.T) {
	clearProxyEnv(t)
	for i := 1; i <= 7; i++ {
		t.Setenv(fmt.Sprintf("DEDICATED_PROXY_%d", i), fmt.Sprintf("10.0.0.%d|8080|user|pass", i))
	}

	attempts := buildOutboundGETAttempts(context.Background(), 2*time.Second, OutboundRequestOptions{})
	require.Len(t, attempts, 2, "expected direct + one dedicated attempt, not all seven proxies")
	require.Equal(t, "direct", attempts[0].strategy)
	require.True(t, strings.HasPrefix(attempts[1].strategy, "dedicated-"))
}

func TestBuildOutboundGETAttempts_UsesRequestDedicatedProxy(t *testing.T) {
	clearProxyEnv(t)
	t.Setenv("DEDICATED_PROXY_1", "1.2.3.4|8080|user|pass")
	t.Setenv("DEDICATED_PROXY_2", "5.6.7.8|8080|user|pass")

	ctx := WithRequestDedicatedProxy(context.Background(), "http://user:pass@1.2.3.4:8080")
	attempts := buildOutboundGETAttempts(ctx, 2*time.Second, OutboundRequestOptions{})
	require.Len(t, attempts, 2)
	require.Equal(t, "direct", attempts[0].strategy)
	require.Equal(t, "dedicated-1", attempts[1].strategy)
	require.Equal(t, "http://user:pass@1.2.3.4:8080", attempts[1].proxyURL)
}

func TestBuildOutboundGETAttempts_SkipDirect(t *testing.T) {
	clearProxyEnv(t)
	t.Setenv("DEDICATED_PROXY_1", "1.2.3.4|8080|user|pass")

	attempts := buildOutboundGETAttempts(context.Background(), 2*time.Second, OutboundRequestOptions{SkipDirect: true})
	require.Len(t, attempts, 1)
	require.True(t, strings.HasPrefix(attempts[0].strategy, "dedicated-"))
}

func TestBuildOutboundGETAttempts_OnlyProxyURL(t *testing.T) {
	clearProxyEnv(t)
	t.Setenv("DEDICATED_PROXY_1", "1.2.3.4|8080|user|pass")

	attempts := buildOutboundGETAttempts(context.Background(), 2*time.Second, OutboundRequestOptions{
		OnlyProxyURL: "http://user:pass@res.proxy:8080",
	})
	require.Len(t, attempts, 1)
	require.Equal(t, "ck-pricelist-proxy", attempts[0].strategy)
	require.Equal(t, "http://user:pass@res.proxy:8080", attempts[0].proxyURL)
}

func TestBuildOutboundGETAttempts_PreferResidentialProxy(t *testing.T) {
	clearProxyEnv(t)
	t.Setenv("RESIDENTIAL_PROXY_1", "res.proxy|8080|res-user|res-pass")
	t.Setenv("DEDICATED_PROXY_1", "1.2.3.4|8080|user|pass")

	attempts := buildOutboundGETAttempts(context.Background(), 2*time.Second, OutboundRequestOptions{
		SkipDirect:             true,
		PreferResidentialProxy: true,
	})
	require.Len(t, attempts, 2)
	require.Equal(t, "residential-1", attempts[0].strategy)
	require.Equal(t, "http://res-user:res-pass@res.proxy:8080", attempts[0].proxyURL)
	require.True(t, strings.HasPrefix(attempts[1].strategy, "dedicated-"))
}

func TestBuildOutboundGETAttempts_PreferResidentialProxyFallsBackWithoutResidential(t *testing.T) {
	clearProxyEnv(t)
	t.Setenv("RESIDENTIAL_PROXY_1", "")
	t.Setenv("DEDICATED_PROXY_1", "1.2.3.4|8080|user|pass")

	attempts := buildOutboundGETAttempts(context.Background(), 2*time.Second, OutboundRequestOptions{
		SkipDirect:             true,
		PreferResidentialProxy: true,
	})
	require.Len(t, attempts, 1)
	require.True(t, strings.HasPrefix(attempts[0].strategy, "dedicated-"))
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

	attempts := buildOutboundGETAttempts(context.Background(), 2*time.Second, OutboundRequestOptions{})
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

func TestDoOutboundGET_429FailsOverWithoutRetry(t *testing.T) {
	clearProxyEnv(t)

	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
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
	require.Equal(t, int32(1), calls.Load())
	require.Contains(t, err.Error(), "direct: status 429")
}

func TestDoOutboundGET_FailsOverOn429ToNextTransport(t *testing.T) {
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

func TestDoOutboundGET_FailsOverOn4xxWithoutRetry(t *testing.T) {
	clearProxyEnv(t)

	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer server.Close()

	_, err := DoOutboundGET(
		context.Background(),
		server.URL,
		OutboundRequestOptions{SkipWebBotAuth: true},
		2*time.Second,
	)
	require.Error(t, err)
	require.Equal(t, int32(1), calls.Load())
	require.Contains(t, err.Error(), "direct: status 404")
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
