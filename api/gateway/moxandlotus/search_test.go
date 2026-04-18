package moxandlotus

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

func TestSearchParsesAPIResponse(t *testing.T) {
	oldTransport := http.DefaultTransport
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		require.Equal(t, "moxandlotus.sg", req.URL.Host)
		require.Equal(t, "/api/products", req.URL.Path)
		require.Equal(t, "Abrade", req.URL.Query().Get("search"))
		body := `{
			"data": [
				{
					"id": 111,
					"title": "Abrade",
					"card_number": "7",
					"expansion_code": "bro",
					"expansion": "The Brothers' War",
					"variation_code": "foil",
					"conditions": [
						{"code":"NM","stocks":2,"price":"9.90"}
					]
				}
			]
		}`
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
	require.Equal(t, 9.90, result[0].Price)
	require.True(t, result[0].InStock)
	require.True(t, result[0].IsFoil)
	require.Equal(t, "Near Mint", result[0].Quality)
	require.Contains(t, result[0].Url, StoreBaseURL+"/view/bro/111")
	require.Contains(t, result[0].Url, "utm_source=gishathfetch")
	require.Contains(t, result[0].Img, "/bro/007.png")
}
