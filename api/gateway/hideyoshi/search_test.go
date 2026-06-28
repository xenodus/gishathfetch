package hideyoshi

import (
	"context"
	"strings"
	"testing"

	"mtg-price-checker-sg/gateway/binderpos"
	"mtg-price-checker-sg/gateway/gatewaytest"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

func init() {
	_ = godotenv.Load("../../.env")
}

func Test_Search(t *testing.T) {
	s := NewLGS()
	result, err := s.Search(context.Background(), "Abrade")
	gatewaytest.RequireSearchOrProbe(t, err, result, gatewaytest.CardExpect{
		URLContains:    StoreBaseURL + "/products/",
		RequireInStock: true,
	}, func(t *testing.T, ctx context.Context) {
		binderpos.RequireStorefrontStructure(t, ctx, binderpos.StructureProbeConfig{
			ScrapVariant:  2,
			BaseURL:       StoreBaseURL,
			SearchURL:     StoreSearchURL,
			ShopifyDomain: StoreShopifyDomain,
			ScrapOnly:     ScrapOnly,
			Query:         "Abrade",
		})
	})
}

func Test_Search_ExcludesPokemonInventory(t *testing.T) {
	s := NewLGS()
	result, err := s.Search(context.Background(), "Bulbasaur")
	require.NoError(t, err)

	for _, card := range result {
		require.True(t, card.InStock, "gateway should only return in-stock variants")
		lower := strings.ToLower(card.Name)
		require.NotContains(t, lower, "mega evolution: base set",
			"Pokemon Bulbasaur inventory should not appear in MTG search results")
		require.NotContains(t, lower, "near mint holofoil",
			"Pokemon condition labels should not appear in MTG search results")
	}
}
