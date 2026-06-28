package gateway

import (
	"context"
	"fmt"
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

func TestDoOutboundGET_DirectForbiddenWithoutProxy(t *testing.T) {
	clearProxyEnv(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	defer server.Close()

	_, err := DoOutboundGET(
		context.Background(),
		server.URL,
		OutboundRequestOptions{Style: OutboundStyleJSON},
		2*time.Second,
	)
	require.Error(t, err)
	require.Contains(t, err.Error(), "403")
}
