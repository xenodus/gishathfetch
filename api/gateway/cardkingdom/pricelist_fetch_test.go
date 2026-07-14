package cardkingdom

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const sampleCKPricelist = `{
  "meta": {
    "created_at": "2026-07-14 09:10:41",
    "base_url": "https://www.cardkingdom.com/"
  },
  "data": [
    {
      "url": "mtg/fourth-edition/lightning-bolt",
      "name": "Lightning Bolt",
      "variation": "",
      "edition": "4th Edition",
      "is_foil": "false",
      "price_retail": "1.49",
      "qty_retail": 3,
      "condition_values": {
        "nm_price": "1.49",
        "nm_qty": 2,
        "ex_price": "1.19",
        "ex_qty": 1
      }
    },
    {
      "url": "mtg/modern-masters-2015/lightning-bolt-foil",
      "name": "Lightning Bolt",
      "variation": "",
      "edition": "Modern Masters 2015",
      "is_foil": "true",
      "price_retail": "3.99",
      "qty_retail": 1,
      "condition_values": {
        "nm_price": "3.99",
        "nm_qty": 1
      }
    },
    {
      "url": "mtg/world-championships/lightning-bolt",
      "name": "Lightning Bolt",
      "variation": "",
      "edition": "World Championships",
      "is_foil": "false",
      "price_retail": "0.99",
      "qty_retail": 1,
      "condition_values": {
        "nm_price": "0.99",
        "nm_qty": 1
      }
    },
    {
      "url": "mtg/marvel-super-heroes-variants/spectacular-spider-man-0014-borderless",
      "name": "Spectacular Spider-Man",
      "variation": "0014 - Borderless",
      "edition": "Marvel's Spider-Man Variants",
      "is_foil": "false",
      "price_retail": "7.49",
      "qty_retail": 0,
      "condition_values": {
        "nm_price": "7.49",
        "nm_qty": 0
      }
    },
    {
      "url": "mtg/promotional/spectacular-spider-man-spotlight-series",
      "name": "Spectacular Spider-Man",
      "variation": "Spotlight Series Non-Foil",
      "edition": "Promotional",
      "is_foil": "false",
      "price_retail": "99.99",
      "qty_retail": 1,
      "condition_values": {
        "nm_price": "99.99",
        "nm_qty": 1
      }
    },
    {
      "url": "mtg/marvel-super-heroes/tony-stark",
      "name": "Tony Stark // The Invincible Iron Man",
      "variation": "",
      "edition": "Marvel Super Heroes",
      "is_foil": "false",
      "price_retail": "7.49",
      "qty_retail": 1,
      "condition_values": {
        "nm_price": "7.49",
        "nm_qty": 1
      }
    },
    {
      "url": "mtg/marvel-super-heroes-variants/tony-stark-0363-borderless-foil",
      "name": "Tony Stark // The Invincible Iron Man",
      "variation": "0363 - Borderless",
      "edition": "Marvel Super Heroes Variants",
      "is_foil": "true",
      "price_retail": "44.99",
      "qty_retail": 1,
      "condition_values": {
        "nm_price": "44.99",
        "nm_qty": 1
      }
    }
  ]
}`

func TestCheapestListedUSD_UsesLowestListedCondition(t *testing.T) {
	price, ok := cheapestListedUSD(ckConditionValues{
		NmPrice: 1.49,
		NmQty:   0,
		ExPrice: 1.19,
		ExQty:   2,
		VgPrice: 0.99,
		VgQty:   1,
	}, 1.49)
	require.True(t, ok)
	require.InDelta(t, 0.99, price, 0.001)
}

func TestCheapestListedUSD_IncludesOutOfStock(t *testing.T) {
	price, ok := cheapestListedUSD(ckConditionValues{
		NmPrice: 7.49,
		NmQty:   0,
	}, 7.49)
	require.True(t, ok)
	require.InDelta(t, 7.49, price, 0.001)
}

func TestCheapestListingsFromPricelist(t *testing.T) {
	var payload ckPricelistPayload
	require.NoError(t, json.Unmarshal([]byte(sampleCKPricelist), &payload))

	updatedAt := time.Date(2026, 7, 14, 9, 10, 41, 0, time.UTC)
	cheapest := cheapestListingsFromPricelist(&payload, updatedAt)

	listing := cheapest["lightning bolt"]
	require.Equal(t, "Lightning Bolt", listing.CardName)
	require.Equal(t, "4th Edition", listing.Edition)
	require.InDelta(t, 1.19, listing.PriceUsd, 0.001)
	require.False(t, listing.IsFoil)
	require.Equal(t, "https://www.cardkingdom.com/mtg/fourth-edition/lightning-bolt", listing.URL)
	require.Equal(t, updatedAt.Format(time.RFC3339), listing.UpdatedAt)

	spider := cheapest["spectacular spider-man"]
	require.InDelta(t, 7.49, spider.PriceUsd, 0.001)
	require.Equal(t, "0014 - Borderless", spider.Edition)

	tony := cheapest["tony stark // the invincible iron man"]
	require.InDelta(t, 7.49, tony.PriceUsd, 0.001)
	require.False(t, tony.IsFoil)

	require.InDelta(t, 7.49, cheapest["tony stark"].PriceUsd, 0.001)
	require.InDelta(t, 7.49, cheapest["the invincible iron man"].PriceUsd, 0.001)
}

func TestFetchCheapestFromCKPricelist_FromTestServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sampleCKPricelist))
	}))
	defer server.Close()

	t.Setenv("CK_PRICELIST_URL", server.URL)

	cheapest, err := fetchCheapestFromCKPricelist(context.Background())
	require.NoError(t, err)
	require.InDelta(t, 1.19, cheapest["lightning bolt"].PriceUsd, 0.001)
	require.InDelta(t, 7.49, cheapest["spectacular spider-man"].PriceUsd, 0.001)
}
