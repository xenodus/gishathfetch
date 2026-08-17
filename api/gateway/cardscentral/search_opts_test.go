package cardscentral

import (
	"testing"

	"mtg-price-checker-sg/gateway"

	"github.com/stretchr/testify/require"
)

func TestCardsCentralOutboundOpts(t *testing.T) {
	opts := cardsCentralOutboundOpts()
	require.Equal(t, gateway.OutboundStyleJSON, opts.Style)
	require.Equal(t, StoreBaseURL, opts.StoreBase.String())
	require.True(t, opts.SkipWebBotAuth)
	require.True(t, opts.PreferResidentialProxy)
	require.False(t, opts.PreferDedicatedFirst)
	require.False(t, opts.SkipDirect)
}
