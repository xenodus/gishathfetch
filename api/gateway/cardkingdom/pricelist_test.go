package cardkingdom

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFetchCheapestByName_ReturnsErrorWhenPricelistBlocked(t *testing.T) {
	pricelistServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "blocked", http.StatusForbidden)
	}))
	defer pricelistServer.Close()

	t.Setenv("CK_PRICELIST_URL", pricelistServer.URL)

	_, err := FetchCheapestByName(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "ck price pricelist")
}
