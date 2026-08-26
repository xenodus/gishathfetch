package tefuda

import (
	"net/url"
	"testing"

	"mtg-price-checker-sg/gateway"

	"github.com/stretchr/testify/require"
)

func TestTefudaOutboundOptsDirectFirst(t *testing.T) {
	storeBase, err := url.Parse(StoreBaseURL)
	require.NoError(t, err)
	pageURL, err := url.Parse(StoreBaseURL + "/search")
	require.NoError(t, err)

	opts := tefudaOutboundOpts(storeBase, pageURL, gateway.OutboundStyleHTML)
	require.False(t, opts.PreferDedicatedFirst)
	require.False(t, opts.SkipDirect)
}
