package fivemana

import (
	"context"
	"strings"
	"testing"

	"mtg-price-checker-sg/gateway/gatewaytest"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/require"
)

func Test_ParseProductCardSkipsSoldOut(t *testing.T) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(soldOutProductHTML))
	require.NoError(t, err)

	card, ok := parseProductCard(doc.Find("ul.product-grid li").First(), StoreName)
	require.False(t, ok, "sold-out listing should be skipped")
	require.Empty(t, card.Name)
}

func Test_ParseProductCardKeepsInStock(t *testing.T) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(inStockProductHTML))
	require.NoError(t, err)

	card, ok := parseProductCard(doc.Find("ul.product-grid li").First(), StoreName)
	require.True(t, ok)
	require.Equal(t, "Abrade [Foundations]", card.Name)
	require.True(t, card.InStock)
	require.Equal(t, 0.40, card.Price)
}

func Test_Search(t *testing.T) {
	s := NewLGS()
	result, err := s.Search(context.Background(), "Abrade")
	gatewaytest.RequireSearchOrProbe(t, err, result, gatewaytest.CardExpect{
		URLContains: StoreBaseURL + "/products/",
	}, func(t *testing.T, ctx context.Context) {
		gatewaytest.RequireFiveManaSearchStructure(t, ctx, StoreBaseURL, StoreSearchPath, "Abrade")
	})
}
