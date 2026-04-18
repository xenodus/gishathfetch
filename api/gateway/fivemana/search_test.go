package fivemana

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestSearchParsesProductGrid(t *testing.T) {
	oldTransport := http.DefaultTransport
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		require.Equal(t, "5-mana.sg", req.URL.Host)
		require.Equal(t, "/search", req.URL.Path)
		require.Equal(t, "Abrade", req.URL.Query().Get("q"))
		require.Equal(t, "1", req.URL.Query().Get("filter.v.availability"))

		body := `
<ul class="product-grid">
	<li>
		<h3 class="card__heading h5"><a href="/products/abrade">Abrade [Foil]</a></h3>
		<div class="card__media"><img src="https://img.example/abrade.png"></div>
		<span class="price-item price-item--sale price-item--last">$2.10</span>
	</li>
</ul>`
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    req,
		}, nil
	})
	t.Cleanup(func() {
		http.DefaultTransport = oldTransport
	})

	s := NewLGS()
	result, err := s.Search(context.Background(), "Abrade")
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, "Abrade", result[0].Name)
	require.Equal(t, StoreName, result[0].Source)
	require.Equal(t, 2.10, result[0].Price)
	require.True(t, result[0].InStock)
	require.True(t, result[0].IsFoil)
	require.Equal(t, "https://img.example/abrade.png", result[0].Img)
	require.Contains(t, result[0].Url, StoreBaseURL+"/products/abrade")
	require.Contains(t, result[0].Url, "utm_source=gishathfetch")
}
