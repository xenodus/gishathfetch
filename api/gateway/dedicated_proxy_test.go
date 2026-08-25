package gateway

import (
	"net/http"
	"testing"
	"time"
)

func TestDedicatedProxiesEnabled(t *testing.T) {
	clearProxyEnv(t)

	t.Run("defaults to enabled when unset and proxies configured", func(t *testing.T) {
		t.Setenv("DEDICATED_PROXY_1", "1.2.3.4|8080|user|pass")
		if !DedicatedProxiesEnabled() {
			t.Fatal("expected dedicated proxies enabled by default")
		}
	})

	t.Run("disabled when USE_DEDICATED_PROXY is false", func(t *testing.T) {
		t.Setenv("USE_DEDICATED_PROXY", "false")
		if DedicatedProxiesEnabled() {
			t.Fatal("expected dedicated proxies disabled")
		}
	})

	t.Run("enabled when configured and toggle is true", func(t *testing.T) {
		t.Setenv("USE_DEDICATED_PROXY", "true")
		t.Setenv("DEDICATED_PROXY_1", "1.2.3.4|8080|user|pass")
		if !DedicatedProxiesEnabled() {
			t.Fatal("expected dedicated proxies enabled")
		}
	})
}

func TestCloseHTTPClientIdleConnections(t *testing.T) {
	t.Run("stdlib transport", func(t *testing.T) {
		client := &http.Client{
			Transport: &http.Transport{},
		}
		CloseHTTPClientIdleConnections(client)
	})

	t.Run("browser emulated transport", func(t *testing.T) {
		t.Setenv("BROWSER_TLS_EMULATION_ENABLED", "true")
		client, err := newOutboundHTTPClient("http://user:pass@1.2.3.4:8080", 2*time.Second, PickBrowserProfile())
		if err != nil {
			t.Fatalf("newOutboundHTTPClient: %v", err)
		}
		CloseHTTPClientIdleConnections(client)
	})
}

func TestCloseTrackedProxyIdleConnections(t *testing.T) {
	clearProxyEnv(t)
	t.Setenv("BROWSER_TLS_EMULATION_ENABLED", "false")
	t.Setenv("DEDICATED_PROXY_1", "1.2.3.4|8080|user|pass")

	client, err := newOutboundHTTPClient("http://user:pass@1.2.3.4:8080", 2*time.Second, BrowserEmulationProfile{})
	if err != nil {
		t.Fatalf("newOutboundHTTPClient: %v", err)
	}
	trackProxyOutboundClient(client)

	CloseTrackedProxyIdleConnections()

	trackedProxyOutboundClients.Range(func(_, _ any) bool {
		t.Fatal("expected tracked proxy clients to be cleared")
		return false
	})
}
