package cardsandcollection

import (
	"context"
	"testing"

	"mtg-price-checker-sg/gateway/gatewaytest"
)

func Test_Search(t *testing.T) {
	s := NewLGS()
	result, err := s.Search(context.Background(), "counterspell")
	gatewaytest.RequireSearchOrProbe(t, err, result, gatewaytest.CardExpect{
		URLContains: StoreBaseURL + "/product/",
	}, func(t *testing.T, ctx context.Context) {
		gatewaytest.RequireCardsAndCollectionAPIStructure(t, ctx, StoreBaseURL, "counterspell")
	})
}
