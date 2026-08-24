package tefuda

import (
	"net/url"
	"testing"

	"mtg-price-checker-sg/gateway"

	"github.com/stretchr/testify/require"
)

func TestTefudaOutboundOptsPreferDedicatedFirst(t *testing.T) {
	storeBase, err := url.Parse(StoreBaseURL)
	require.NoError(t, err)
	pageURL, err := url.Parse(StoreBaseURL + "/search")
	require.NoError(t, err)

	opts := tefudaOutboundOpts(storeBase, pageURL, gateway.OutboundStyleHTML)
	require.True(t, opts.PreferDedicatedFirst)
	require.False(t, opts.SkipDirect)
}
