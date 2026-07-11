package fyendalhobby

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
	s := NewLGS()
	result, err := s.Search(context.Background(), "Abrade")
	gatewaytest.RequireSearchOrProbe(t, err, result, gatewaytest.CardExpect{
		URLContains: StoreBaseURL + "/products/",
	}, func(t *testing.T, ctx context.Context) {
		binderpos.RequireStorefrontStructure(t, ctx, binderpos.StructureProbeConfig{
			ScrapVariant:  4,
			BaseURL:       StoreBaseURL,
			SearchURL:     StoreSearchURL,
			ShopifyDomain: StoreShopifyDomain,
			ScrapOnly:     ScrapOnly,
			Query:         "Abrade",
		})
	})
}
