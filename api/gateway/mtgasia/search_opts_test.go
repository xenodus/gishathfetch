package mtgasia

import (
	"testing"

	"mtg-price-checker-sg/gateway"

	"github.com/stretchr/testify/require"
)

func Test_base44OutboundOpts_pinsDedicatedProxy1Fallback(t *testing.T) {
	t.Setenv("DEDICATED_PROXY_1", "1.2.3.4|8080|user|pass")
	t.Setenv("DEDICATED_PROXY_2", "5.6.7.8|8080|user|pass")

	opts := base44OutboundOpts()
	require.Equal(t, gateway.OutboundStyleJSON, opts.Style)
	require.Equal(t, "http://user:pass@1.2.3.4:8080", opts.DedicatedFallbackProxyURL)
}

func Test_base44OutboundOpts_withoutDedicatedProxy1(t *testing.T) {
	t.Setenv("DEDICATED_PROXY_1", "")

	opts := base44OutboundOpts()
	require.Empty(t, opts.DedicatedFallbackProxyURL)
}
