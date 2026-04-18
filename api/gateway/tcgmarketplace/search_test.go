package tcgmarketplace

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

func TestSearchParsesMarketplaceResponse(t *testing.T) {
	oldTransport := http.DefaultTransport
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		require.Equal(t, "thetcgmarketplace.com:3501", req.URL.Host)
		require.Equal(t, "/encoder/advancedsearch", req.URL.Path)
		require.Equal(t, http.MethodPost, req.Method)

		body := `{
			"status": 200,
			"data": {
				"message": "ok",
				"data": [
					{
						"name": "[BRC] Abrade",
						"setname": "The Brothers' War Commander",
						"image": "https://img.example/abrade.png alt",
						"available": 2,
						"from": 1.8,
						"url": "https://thetcgmarketplace.com/product/B/abrade"
					}
				]
			},
			"meta": {"total": 1}
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
	t.Setenv(accessTokenKey, "test-token")

	s := NewLGS()
	result, err := s.Search(context.Background(), "abrade")
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, "Abrade", result[0].Name)
	require.Equal(t, StoreName, result[0].Source)
	require.Equal(t, 1.8, result[0].Price)
	require.True(t, result[0].InStock)
	require.Equal(t, "https://img.example/abrade.png", result[0].Img)
	require.Equal(t, []string{"[The Brothers' War Commander]"}, result[0].ExtraInfo)
	require.Contains(t, result[0].Url, StoreBaseURL+"/product/B/abrade")
	require.Contains(t, result[0].Url, "utm_source=gishathfetch")
}
