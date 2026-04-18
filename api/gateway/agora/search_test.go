package agora

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSearchParsesStoreResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/store/search", r.URL.Path)
		require.Equal(t, "mtg", r.URL.Query().Get("category"))
		require.Equal(t, "Abrade", r.URL.Query().Get("searchfield"))

		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`
<div id="store_listingcontainer">
  <div class="store-item">
    <div class="store-item-stock">Stock: 1</div>
    <div class="store-item-price">$2.50</div>
    <div class="store-item-cat">Single - NM [BRO]</div>
    <div class="store-item-title">Abrade</div>
    <div class="store-item-img" data-img="https://img.example/abrade.png"></div>
  </div>
  <div class="store-item">
    <div class="store-item-stock">Stock: 0</div>
    <div class="store-item-price">$0.00</div>
    <div class="store-item-cat">Single - LP</div>
    <div class="store-item-title">Abrade (Sold Out)</div>
    <div class="store-item-img" data-img="https://img.example/abrade-soldout.png"></div>
  </div>
</div>`))
	}))
	defer server.Close()

	baseURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	s := Store{
		Name:       StoreName,
		BaseUrl:    server.URL,
		SearchPath: "/store/search",
	}

	result, err := s.Search(context.Background(), "Abrade")
	require.NoError(t, err)
	require.Len(t, result, 1)

	card := result[0]
	require.True(t, card.InStock)
	require.Equal(t, "Abrade", card.Name)
	require.Equal(t, StoreName, card.Source)
	require.Equal(t, 2.5, card.Price)
	require.Equal(t, "https://img.example/abrade.png", card.Img)
	require.Contains(t, card.Url, baseURL.String()+"/store/search?category=mtg&searchfield=Abrade")
	require.Contains(t, card.Url, "utm_source=gishathfetch")
	require.Equal(t, []string{"Single - NM [BRO]"}, card.ExtraInfo)
}
