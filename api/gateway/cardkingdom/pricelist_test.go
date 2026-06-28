package cardkingdom

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestUnmarshalPricelistPayload(t *testing.T) {
	const sample = `{
		"meta":{"created_at":"2026-06-28 07:07:57"},
		"data":[
			{
				"name":"Lightning Bolt",
				"edition":"4th Edition",
				"price_retail":0.35,
				"qty_retail":15,
				"url":"mtg/4th-edition/lightning-bolt",
				"is_foil":"false"
			}
		]
	}`

	var payload struct {
		Data []Product `json:"data"`
	}
	require.NoError(t, json.Unmarshal([]byte(sample), &payload))
	require.Len(t, payload.Data, 1)

	cheapest := BuildCheapestByName(payload.Data, mustParseTime(t, "2026-06-28T07:07:57Z"))
	listing := cheapest["lightning bolt"]
	require.InDelta(t, 0.35, listing.PriceUsd, 0.001)
	require.Equal(t, 15, listing.Quantity)
}

func mustParseTime(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, value)
	require.NoError(t, err)
	return parsed
}
