package fivemana

import (
	"context"
	"testing"

	"mtg-price-checker-sg/gateway/gatewaytest"
)

func Test_Search(t *testing.T) {
	s := NewLGS()
	result, err := s.Search(context.Background(), "Abrade")
	gatewaytest.RequireSearchOrProbe(t, err, result, gatewaytest.CardExpect{
		URLContains: StoreBaseURL + "/products/",
	}, func(t *testing.T, ctx context.Context) {
		gatewaytest.RequireFiveManaSearchStructure(t, ctx, StoreBaseURL, StoreSearchPath, "Abrade")
	})
}
