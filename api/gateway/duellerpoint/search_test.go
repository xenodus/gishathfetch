package duellerpoint

import (
	"context"
	"encoding/json"
	"testing"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/gatewaytest"

	"github.com/stretchr/testify/require"
)

func Test_parseSearchResult(t *testing.T) {
	card, ok := parseSearchResult(searchResult{
		Name:               "Lightning Bolt",
		Slug:               "lightning-bolt-a25",
		Price:              "1.75",
		FoilPrice:          "0.0",
		Quantity:           3,
		IsActive:           true,
		FoilType:           "non",
		GetNameWithFoil:    "Lightning Bolt (a25)",
		GetVariationName:   "The List",
		TCGPlayerProductID: "240810",
	})
	require.True(t, ok)
	require.Equal(t, "Lightning Bolt (a25)", card.Name)
	require.Equal(t, StoreBaseURL+"/products/lightning-bolt-a25", card.Url)
	require.Equal(t, "https://product-images.tcgplayer.com/fit-in/437x437/240810.jpg", card.Img)
	require.InDelta(t, 1.75, card.Price, 0.001)
	require.True(t, card.InStock)
	require.False(t, card.IsFoil)
	require.Equal(t, []string{"[The List]"}, card.ExtraInfo)
}

func Test_parseSearchResult_usesFoilPrice(t *testing.T) {
	card, ok := parseSearchResult(searchResult{
		Name:               "Lightning Bolt",
		Slug:               "lightning-bolt-foil",
		Price:              "17.75",
		FoilPrice:          "10.16",
		Quantity:           4,
		IsActive:           true,
		FoilType:           "foil",
		GetNameWithFoil:    "Lightning Bolt (Foil)",
		GetVariationName:   "Magic 2010 (M10)",
		TCGPlayerProductID: "32656-foil",
	})
	require.True(t, ok)
	require.True(t, card.IsFoil)
	require.InDelta(t, 10.16, card.Price, 0.001)
	require.Equal(t, "https://product-images.tcgplayer.com/fit-in/437x437/32656.jpg", card.Img)
}

func Test_parseSearchResult_skipsOutOfStock(t *testing.T) {
	_, ok := parseSearchResult(searchResult{
		Name:     "Lightning Bolt",
		Slug:     "lightning-bolt-a25",
		Price:    "1.75",
		Quantity: 0,
		IsActive: true,
	})
	require.False(t, ok)
}

func Test_Search(t *testing.T) {
	s := NewLGS()
	result, err := s.Search(context.Background(), "lightning bolt")
	gatewaytest.RequireSearchOrProbe(t, err, result, gatewaytest.CardExpect{
		URLContains:    StoreBaseURL + "/products/",
		Source:         StoreName,
		RequireInStock: true,
	}, func(t *testing.T, ctx context.Context) {
		gatewaytest.RequireDuellersPointSearchStructure(t, ctx, StoreBaseURL, StoreSearchPath, "lightning bolt")
	})
}

func Test_parseSearchResponse(t *testing.T) {
	var payload searchResponse
	require.NoError(t, jsonUnmarshal(searchResultsFixture, &payload))
	require.Len(t, payload.Results, 2)

	cards := make([]gateway.Card, 0, len(payload.Results))
	for _, item := range payload.Results {
		card, ok := parseSearchResult(item)
		require.True(t, ok)
		cards = append(cards, card)
	}
	require.Len(t, cards, 2)
}

const searchResultsFixture = `{
  "results": [
    {
      "name": "Lightning Bolt",
      "slug": "lightning-bolt-a25",
      "price": "1.75",
      "foil_price": "0.0",
      "quantity": 3,
      "is_active": true,
      "foil_type": "non",
      "get_name_with_foil": "Lightning Bolt (a25)",
      "get_variation_name": "The List",
      "tcgplayer_product_id_ext": "240810"
    },
    {
      "name": "Lightning Bolt",
      "slug": "lightning-bolt-foil",
      "price": "17.75",
      "foil_price": "10.16",
      "quantity": 4,
      "is_active": true,
      "foil_type": "foil",
      "get_name_with_foil": "Lightning Bolt (Foil)",
      "get_variation_name": "Magic 2010 (M10)",
      "tcgplayer_product_id_ext": "32656-foil"
    }
  ]
}`

func jsonUnmarshal(raw string, dst any) error {
	return json.Unmarshal([]byte(raw), dst)
}
