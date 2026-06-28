package fyendalhobby

import (
	"context"
	"testing"

	"mtg-price-checker-sg/gateway/gatewaytest"
	"mtg-price-checker-sg/gateway/shopifysuggest"

	"github.com/joho/godotenv"
)

func init() {
	_ = godotenv.Load("../../.env")
}

func Test_Search(t *testing.T) {
	s := NewLGS()
	result, err := s.Search(context.Background(), "Abrade")
	gatewaytest.RequireSearchOrProbe(t, err, result, gatewaytest.CardExpect{
		URLContains: StoreBaseURL + "/products/",
	}, func(t *testing.T, ctx context.Context) {
		shopifysuggest.RequireSuggestStructure(t, ctx, shopifysuggest.Options{
			Config: shopifysuggest.Config{
				StoreName: StoreName,
				BaseURL:   StoreBaseURL,
			},
			SearchStr:   "Abrade",
			BuildQuery:  shopifysuggest.FyendalQuery,
			QueryValues: shopifysuggest.FyendalQueryValues,
			MapProduct:  shopifysuggest.MapFyendalProduct,
		})
	})
}
