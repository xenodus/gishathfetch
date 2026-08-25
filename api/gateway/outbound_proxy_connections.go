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

// CloseTrackedProxyIdleConnections closes idle keep-alive connections on outbound
// HTTP clients that routed through a proxy during the current search fan-out.
func CloseTrackedProxyIdleConnections() {
	trackedProxyOutboundClients.Range(func(key, _ any) bool {
		CloseHTTPClientIdleConnections(key.(*http.Client))
		trackedProxyOutboundClients.Delete(key)
		return true
	})
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
