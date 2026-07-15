package fivemana

import (
	"context"
	"encoding/json"
	"testing"

	"mtg-price-checker-sg/gateway/gatewaytest"

	"github.com/stretchr/testify/require"
)

func Test_ParseSuggestProductSkipsUnavailable(t *testing.T) {
	var payload suggestResponse
	require.NoError(t, json.Unmarshal([]byte(unavailableSuggestJSON), &payload))

	card, ok := parseSuggestProduct(payload.Resources.Results.Products[0], StoreName)
	require.False(t, ok, "unavailable listing should be skipped")
	require.Empty(t, card.Name)
}

func Test_ParseSuggestProductKeepsInStock(t *testing.T) {
	var payload suggestResponse
	require.NoError(t, json.Unmarshal([]byte(inStockSuggestJSON), &payload))

	card, ok := parseSuggestProduct(payload.Resources.Results.Products[0], StoreName)
	require.True(t, ok)
	require.Equal(t, "Abrade [Foundations]", card.Name)
	require.False(t, card.IsFoil)
	require.True(t, card.InStock)
	require.Equal(t, 0.40, card.Price)
	require.Contains(t, card.Url, StoreBaseURL+"/products/abrade-foundations")
}

func Test_ParseSuggestProductDetectsFoil(t *testing.T) {
	var payload suggestResponse
	require.NoError(t, json.Unmarshal([]byte(foilSuggestJSON), &payload))

	card, ok := parseSuggestProduct(payload.Resources.Results.Products[0], StoreName)
	require.True(t, ok)
	require.Equal(t, "Rhystic Study", card.Name)
	require.True(t, card.IsFoil)
}

func Test_Search(t *testing.T) {
	s := NewLGS()
	result, err := s.Search(context.Background(), "Abrade")
	gatewaytest.RequireSearchOrProbe(t, err, result, gatewaytest.CardExpect{
		URLContains: StoreBaseURL + "/products/",
	}, func(t *testing.T, ctx context.Context) {
		gatewaytest.RequireFiveManaSearchStructure(t, ctx, StoreBaseURL, StoreSearchPath, "Abrade")
	})
}
