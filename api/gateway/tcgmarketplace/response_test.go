package tcgmarketplace

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseResponseBody_success(t *testing.T) {
	body := []byte(`{
		"status": 200,
		"data": {
			"message": "",
			"data": [{
				"name": " [DOM] Teferi, Hero of Dominaria",
				"setcode": "DOM",
				"setname": "Dominaria",
				"image": "https://thetcgmarketplace.com:3500/uploads/example.webp",
				"language": "en",
				"crd_foil_type": null,
				"rarity": "Mythic",
				"available": 4,
				"from": "7.49",
				"non_foil_reference_price": "SGD 8.00",
				"foil_reference_price": "USD 16.29",
				"url": "https://thetcgmarketplace.com/product/B/example/0"
			}]
		},
		"meta": {"total": 1}
	}`)

	items, err := parseResponseBody(body)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, " [DOM] Teferi, Hero of Dominaria", items[0].Name)
	require.Equal(t, "Dominaria", items[0].Setname)
}

func TestParseResponseBody_unauthorized(t *testing.T) {
	body := []byte(`{"status":500,"data":{"message":"Unathorised","data":""},"meta":{"total":0}}`)

	_, err := parseResponseBody(body)
	require.Error(t, err)
	require.Contains(t, err.Error(), "API error (status=500)")
	require.Contains(t, err.Error(), "Unathorised")
}

func TestParseResponseBody_databaseError(t *testing.T) {
	body := []byte(`{"status":500,"data":{"message":"","data":{"errno":-111,"code":"ECONNREFUSED","syscall":"connect","address":"::1","port":3306,"fatal":true}},"meta":{}}`)

	_, err := parseResponseBody(body)
	require.Error(t, err)
	require.Contains(t, err.Error(), "API error (status=500)")
	require.Contains(t, err.Error(), "ECONNREFUSED")
	require.Contains(t, err.Error(), "::1:3306")
}

func TestParseResponseBody_emptyResults(t *testing.T) {
	body := []byte(`{"status":200,"data":{"message":"","data":[]},"meta":{"total":0}}`)

	items, err := parseResponseBody(body)
	require.NoError(t, err)
	require.Empty(t, items)
}

func TestParseResponseBody_unexpectedDataShape(t *testing.T) {
	body := []byte(`{"status":200,"data":{"message":"","data":{"unexpected":"value"}},"meta":{}}`)

	_, err := parseResponseBody(body)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "unexpected data payload"))
}
