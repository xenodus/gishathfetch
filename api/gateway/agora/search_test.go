package agora

import (
	"context"
	"strings"
	"testing"

	"mtg-price-checker-sg/gateway/gatewaytest"

	"github.com/stretchr/testify/require"
)

func Test_Search(t *testing.T) {
	s := NewLGS()
	result, err := s.Search(context.Background(), "Abrade")
	gatewaytest.RequireSearchOrProbe(t, err, result, gatewaytest.CardExpect{
		URLContains:    StoreBaseURL + "/store/search?category=" + storeCategoryMTG + "&searchfield=",
		RequireInStock: true,
	}, func(t *testing.T, ctx context.Context) {
		gatewaytest.RequireAgoraSearchStructure(t, ctx, StoreBaseURL, StoreSearchPath, storeCategoryMTG, "Abrade")
	})
}

func Test_Search_FiltersMTGCategory(t *testing.T) {
	s := NewLGS()
	result, err := s.Search(context.Background(), "Bulbasaur")
	require.NoError(t, err)

	for _, card := range result {
		require.True(t, card.InStock, "gateway should only return in-stock listings")
		require.Contains(t, card.Url, "category="+storeCategoryMTG,
			"Agora product links should stay scoped to the MTG category")
		lower := strings.ToLower(card.Name)
		require.NotContains(t, lower, "pokemon",
			"Pokemon inventory should not appear when category=mtg is set")
		require.NotContains(t, lower, "holofoil",
			"Pokemon condition labels should not appear when category=mtg is set")
	}
}
