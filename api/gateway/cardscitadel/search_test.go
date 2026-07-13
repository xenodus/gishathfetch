package cardscitadel

import (
	"context"
	"testing"

	"mtg-price-checker-sg/gateway/binderpos"
	"mtg-price-checker-sg/gateway/gatewaytest"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

func Test_NewLGS(t *testing.T) {
	require.Equal(t, Store{
		Name:         StoreName,
		BaseUrl:      StoreBaseURL,
		SearchUrl:    StoreSearchURL,
		BinderposGwy: binderpos.New(),
	}, NewLGS())
}

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
			ScrapVariant:  1,
			BaseURL:       StoreBaseURL,
			SearchURL:     StoreSearchURL,
			Query:         "Abrade",
		})
	})
}
