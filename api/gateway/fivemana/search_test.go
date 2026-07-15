package fivemana

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"testing"

	"mtg-price-checker-sg/gateway/gatewaytest"

	"github.com/stretchr/testify/require"
)

func TestMapSuggestProductToCardSkipsUnavailable(t *testing.T) {
	card, ok := mapSuggestProductToCard(suggestProduct{
		Available: false,
		Handle:    "the-ten-rings-marvel-super-heroes",
		Image:     "https://cdn.shopify.com/files/ten-rings.png",
		Price:     "9.90",
		Title:     "The Ten Rings [Marvel Super Heroes]",
	}, StoreName)
	require.False(t, ok)
	require.Empty(t, card.Name)
}

func TestMapSuggestProductToCardKeepsInStock(t *testing.T) {
	card, ok := mapSuggestProductToCard(suggestProduct{
		Available: true,
		Handle:    "abrade-foundations",
		Image:     "https://cdn.shopify.com/files/abrade.png",
		Price:     "0.40",
		Tags:      []string{"Foundations", "Foundations Non-Foil"},
		Title:     "Abrade [Foundations]",
	}, StoreName)
	require.True(t, ok)
	require.Equal(t, "Abrade [Foundations]", card.Name)
	require.False(t, card.IsFoil)
	require.True(t, card.InStock)
	require.Equal(t, 0.40, card.Price)
	require.Contains(t, card.Url, StoreBaseURL+"/products/abrade-foundations")
	require.Contains(t, card.Url, "utm_source=")
}

func TestMapSuggestProductToCardDetectsFoilFromTitleAndTags(t *testing.T) {
	fromTitle, ok := mapSuggestProductToCard(suggestProduct{
		Available: true,
		Handle:    "bolt-foil-title",
		Price:     "5.00",
		Title:     "Lightning Bolt [Foil]",
	}, StoreName)
	require.True(t, ok)
	require.True(t, fromTitle.IsFoil)
	require.Equal(t, "Lightning Bolt", fromTitle.Name)

	fromTag, ok := mapSuggestProductToCard(suggestProduct{
		Available: true,
		Handle:    "bolt-foil-tag",
		Price:     "5.00",
		Tags:      []string{"Foil"},
		Title:     "Lightning Bolt [Alpha]",
	}, StoreName)
	require.True(t, ok)
	require.True(t, fromTag.IsFoil)
}

func TestMapSuggestProductsToCardsFromFixture(t *testing.T) {
	var parsed suggestResponse
	require.NoError(t, json.Unmarshal([]byte(suggestResponseFixture), &parsed))

	cards := mapSuggestProductsToCards(parsed.Resources.Results.Products, StoreName)
	require.Len(t, cards, 2)
	require.Equal(t, "Abrade [Foundations]", cards[0].Name)
	require.Equal(t, "Lightning Bolt [Alpha]", cards[1].Name)
	require.True(t, cards[1].IsFoil)
}

func TestIsSuggestProductFoilIgnoresNonFoilTags(t *testing.T) {
	require.False(t, isSuggestProductFoil("Abrade [Foundations]", []string{"Foundations Non-Foil"}))
}

func Test_Search(t *testing.T) {
	s := NewLGS()
	result, err := s.Search(context.Background(), "Abrade")
	gatewaytest.RequireSearchOrProbe(t, err, result, gatewaytest.CardExpect{
		URLContains: StoreBaseURL + "/products/",
	}, func(t *testing.T, ctx context.Context) {
		host := strings.TrimPrefix(strings.TrimPrefix(StoreBaseURL, "https://"), "http://")
		probeURL := gatewaytest.BuildURL("https", host, StoreSuggestPath, url.Values{
			"q":                                      {"Abrade"},
			"resources[type]":                        {"product"},
			"resources[limit]":                       {suggestProductLimit},
			"resources[options][unavailable_products]": {"hide"},
		})
		gatewaytest.RequireJSONStructure(t, ctx, gatewaytest.JSONProbe{
			URL: probeURL,
			Validate: func(body []byte) error {
				var parsed suggestResponse
				if err := json.Unmarshal(body, &parsed); err != nil {
					return err
				}
				if parsed.Resources.Results.Products == nil {
					return gatewaytest.ValidateErrorf("expected resources.results.products")
				}
				return nil
			},
		})
	})
}
