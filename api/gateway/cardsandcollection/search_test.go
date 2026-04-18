package cardsandcollection

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSearchParsesAPIResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/catalog/", r.URL.Path)
		require.Equal(t, http.MethodPost, r.Method)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"hits": {
				"total": 1,
				"hits": [
					{
						"_id": "abc123",
						"_source": {
							"name": "Counterspell",
							"img": "https://img.example/counterspell.png",
							"setName": "Revised",
							"quantityOnSale": "2",
							"minPriceSale": "3.75"
						}
					}
				]
			}
		}`))
	}))
	defer server.Close()

	store := Store{
		Name:      StoreName,
		BaseUrl:   server.URL,
		SearchUrl: StoreSearchURL,
	}

	result, err := store.Search(context.Background(), "counterspell")
	require.NoError(t, err)
	require.Len(t, result, 1)

	card := result[0]
	require.Equal(t, "Counterspell", card.Name)
	require.Equal(t, StoreName, card.Source)
	require.Equal(t, 3.75, card.Price)
	require.True(t, card.InStock)
	require.Equal(t, "https://img.example/counterspell.png", card.Img)
	require.Equal(t, []string{"[Revised]"}, card.ExtraInfo)
	require.Contains(t, card.Url, fmt.Sprintf("%s/product/abc123", StoreBaseURL))
	require.Contains(t, card.Url, "utm_source=gishathfetch")
}
