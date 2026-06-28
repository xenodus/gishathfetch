package arcanesanctum

import (
	"context"
	"testing"

	"mtg-price-checker-sg/gateway/binderpos"
	"mtg-price-checker-sg/gateway/gatewaytest"

	"github.com/joho/godotenv"
)

func init() {
	_ = godotenv.Load("../../.env")
}

func Test_Search(t *testing.T) {
	t.Skip("Arcane Sanctum is temporarily disabled in controller; re-enable this test when the store is wired back in")
	s := NewLGS()
	result, err := s.Search(context.Background(), "signet")
	gatewaytest.RequireSearchOrProbe(t, err, result, gatewaytest.CardExpect{
		URLContains: StoreBaseURL + "/products/",
	}, func(t *testing.T, ctx context.Context) {
		binderpos.RequireStorefrontStructure(t, ctx, binderpos.StructureProbeConfig{
			ScrapVariant:  5,
			BaseURL:       StoreBaseURL,
			SearchURL:     StoreSearchURL,
			ShopifyDomain: StoreShopifyDomain,
			ScrapOnly:     ScrapOnly,
			Query:         "signet",
		})
	})
}
