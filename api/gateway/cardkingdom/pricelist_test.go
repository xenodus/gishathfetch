package cardkingdom

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFetchCheapestByName_FallsBackToMTGJSONWhenPricelistBlocked(t *testing.T) {
	pricelistServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "blocked", http.StatusForbidden)
	}))
	defer pricelistServer.Close()

	pricesServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = writeBzip2(w, []byte(sampleAllPricesToday))
	}))
	defer pricesServer.Close()

	printingsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = writeBzip2(w, []byte(sampleAllPrintings))
	}))
	defer printingsServer.Close()

	t.Setenv("CK_PRICELIST_URL", pricelistServer.URL)
	t.Setenv("MTGJSON_ALL_PRICES_TODAY_URL", pricesServer.URL)
	t.Setenv("MTGJSON_ALL_PRINTINGS_URL", printingsServer.URL)

	cheapest, err := FetchCheapestByName(context.Background())
	require.NoError(t, err)
	require.InDelta(t, 1.49, cheapest["lightning bolt"].PriceUsd, 0.001)
}
