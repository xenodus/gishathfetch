package cardscentral

import (
	"testing"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/pkg/config"

	"github.com/stretchr/testify/require"
)

func TestCardsCentralOutboundOpts_SetsAPIKeyHeader(t *testing.T) {
	t.Setenv(config.CardsCentralKeyEnv, "configured-key")
	opts := cardsCentralOutboundOpts()
	require.Equal(t, gateway.OutboundStyleJSON, opts.Style)
	require.Equal(t, map[string]string{
		config.CardsCentralKeyHeader: "configured-key",
	}, opts.ExtraHeaders)
}

func TestCardsCentralOutboundOpts_OmitsHeaderWhenUnset(t *testing.T) {
	t.Setenv(config.CardsCentralKeyEnv, "")
	opts := cardsCentralOutboundOpts()
	require.Equal(t, gateway.OutboundStyleJSON, opts.Style)
	require.Nil(t, opts.ExtraHeaders)
}
