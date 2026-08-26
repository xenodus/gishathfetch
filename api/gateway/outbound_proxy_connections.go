package gateway

import (
	"net/http"
	"sync"
)

var trackedProxyOutboundClients sync.Map

func trackProxyOutboundClient(client *http.Client) {
	if client == nil {
		return
	}
	trackedProxyOutboundClients.Store(client, client)
}

// TrackedProxyOutboundClientCount reports proxy-backed outbound clients awaiting
// idle connection cleanup. Intended for tests that verify cleanup after outbound work.
func TrackedProxyOutboundClientCount() int {
	count := 0
	trackedProxyOutboundClients.Range(func(_, _ any) bool {
		count++
		return true
	})
	return count
}

// CloseTrackedProxyIdleConnections closes idle keep-alive connections on outbound
// HTTP clients that routed through a proxy during outbound scraping work.
func CloseTrackedProxyIdleConnections() {
	trackedProxyOutboundClients.Range(func(key, _ any) bool {
		closeProxyOutboundClient(key.(*http.Client))
		return true
	})
}

func closeProxyOutboundClient(client *http.Client) {
	if client == nil {
		return
	}
	CloseHTTPClientIdleConnections(client)
	trackedProxyOutboundClients.Delete(client)
}

// CloseHTTPClientIdleConnections drops idle pooled connections for an HTTP client.
func CloseHTTPClientIdleConnections(client *http.Client) {
	if client == nil || client.Transport == nil {
		return
	}
	switch transport := client.Transport.(type) {
	case *http.Transport:
		transport.CloseIdleConnections()
	case *browserTLSRoundTripper:
		transport.inner.CloseIdleConnections()
	}
}
