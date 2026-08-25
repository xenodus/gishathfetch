package gateway

import (
	"mtg-price-checker-sg/gateway/util"
	"mtg-price-checker-sg/pkg/config"
)

// DedicatedProxiesEnabled reports whether dedicated proxy env vars are configured
// and USE_DEDICATED_PROXY has not disabled them.
func DedicatedProxiesEnabled() bool {
	if !config.UseDedicatedProxy() {
		return false
	}
	return len(util.GetDedicatedProxyURLs()) > 0
}
