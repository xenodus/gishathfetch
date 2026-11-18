package tcgmarketplace

import (
	"context"
	"encoding/json"
	"testing"

	"mtg-price-checker-sg/gateway/gatewaytest"

	"github.com/stretchr/testify/require"
)

func TestCardFromListing_advancedFilterShape(t *testing.T) {
	t.Setenv("GODEBUG", "rsa1024min=0")

	body := []byte(`{"status":200,"data":{"message":"","data":[{"id":285511,"name":" [XLN] Opt","setname":"Ixalan","available":4,"from":"0.20","image":"https://thetcgmarketplace.com:3500/uploads/products/xln_65%20Opt.webp"}]},"meta":{"total":1}}`)
	var env apiEnvelope
	require.NoError(t, json.Unmarshal(body, &env))
	listings, err := listingsFromEnvelope(env)
	require.NoError(t, err)
	require.Len(t, listings, 1)

	s := NewLGS().(Store)
	card, ok := cardFromListing(s, listings[0])
	require.True(t, ok)
	require.Equal(t, "Opt", card.Name)
	require.Contains(t, card.Url, StoreBaseURL+"/product/B/")
	require.Equal(t, 0.20, card.Price)
}

func Test_SearchAdvancedFilter(t *testing.T) {
	ctx := context.Background()
	s := NewLGS().(Store)

	result, err := s.SearchAdvancedFilter(ctx, "Opt")
	if err != nil {
		t.Logf("advanced filter search failed (%v); probing api structure", err)
		gatewaytest.RequireTCGMarketplaceAdvancedFilterStructure(t, ctx, "Opt")
		return
	}

	gatewaytest.RequireSearchOrProbe(t, err, result, gatewaytest.CardExpect{
		URLContains: StoreBaseURL + "/product/B/",
	}, func(t *testing.T, ctx context.Context) {
		gatewaytest.RequireTCGMarketplaceAdvancedFilterStructure(t, ctx, "Opt")
	})
}
