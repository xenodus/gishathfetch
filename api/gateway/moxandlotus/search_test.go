package moxandlotus

import (
	"context"
	"testing"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/gatewaytest"

	"github.com/stretchr/testify/require"
)

func TestMoxOutboundOpts(t *testing.T) {
	opts := moxOutboundOpts()
	require.Equal(t, gateway.OutboundStyleJSON, opts.Style)
	require.Equal(t, StoreBaseURL, opts.StoreBase.String())
	require.True(t, opts.SkipWebBotAuth)
	require.False(t, opts.PreferDedicatedFirst)
	require.False(t, opts.SkipDirect)
	require.False(t, opts.PreferResidentialProxy)
}

func TestMoxSearchLimit(t *testing.T) {
	require.Equal(t, "24", storeSearchLimit)
}

func Test_Search(t *testing.T) {
	skipLiveMoxSearchUnlessDedicatedProxy(t)

	s := NewLGS()
	result, err := s.Search(context.Background(), "Abrade")
	gatewaytest.RequireSearchOrProbe(t, err, result, gatewaytest.CardExpect{
		URLContains: StoreBaseURL + "/view/",
	}, func(t *testing.T, ctx context.Context) {
		gatewaytest.RequireMoxAndLotusAPIStructure(t, ctx, "Abrade")
	})
}

func skipLiveMoxSearchUnlessDedicatedProxy(t *testing.T) {
	t.Helper()
	if gateway.DedicatedProxiesEnabled() {
		return
	}
	t.Skip("set USE_DEDICATED_PROXY=true and DEDICATED_PROXY_* to run live Mox & Lotus search checks")
}

func Test_resolveCardImageURL(t *testing.T) {
	t.Run("uses image path from API when available", func(t *testing.T) {
		actual, ok := resolveCardImageURL("SOC", "55", "https://d3nmvyqkci0c2u.cloudfront.net/SOC/card-418530-325166.jpg")
		require.True(t, ok)
		require.Equal(t, "https://d3nmvyqkci0c2u.cloudfront.net/SOC/card-418530-325166.jpg", actual)
	})

	t.Run("uses fallback image path when image path is empty", func(t *testing.T) {
		actual, ok := resolveCardImageURL("SOC", "55", "")
		require.True(t, ok)
		require.Equal(t, "https://d3nmvyqkci0c2u.cloudfront.net/SOC/055.png", actual)
	})

	t.Run("returns empty image when fallback card number is invalid", func(t *testing.T) {
		actual, ok := resolveCardImageURL("SOC", "invalid", nil)
		require.False(t, ok)
		require.Equal(t, "", actual)
	})
}
