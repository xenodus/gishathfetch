package manapro

import (
	"context"
	"testing"

	"mtg-price-checker-sg/gateway/binderpos"
	"mtg-price-checker-sg/gateway/gatewaytest"
	"mtg-price-checker-sg/pkg/config"

	"github.com/joho/godotenv"
)

func init() {
	_ = godotenv.Load("../../.env")
}

func Test_Search(t *testing.T) {
	if !config.ManaproSearchEnabled {
		t.Skip("Mana Pro search is disabled while the storefront is password-protected")
	}

	s := NewLGS()
	result, err := s.Search(context.Background(), "Abrade")
	gatewaytest.RequireSearchOrProbe(t, err, result, gatewaytest.CardExpect{
		URLContains: StoreBaseURL + "/products/",
	}, func(t *testing.T, ctx context.Context) {
		binderpos.RequireStorefrontStructure(t, ctx, binderpos.StructureProbeConfig{
			ScrapVariant:  2,
			BaseURL:       StoreBaseURL,
			SearchURL:     StoreSearchURL,
			Query:         "Abrade",
		})
	})
}
