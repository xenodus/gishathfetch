package duellerpoint

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

func TestSearchParsesHTMLTable(t *testing.T) {
	oldTransport := http.DefaultTransport
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		require.Equal(t, "www.duellerspoint.com", req.URL.Host)
		require.Equal(t, "/products/search", req.URL.Path)
		require.Equal(t, "Abrade", req.URL.Query().Get("search_text"))

		body := `
<div class="container">
  <table>
    <tbody>
      <tr>
        <td><a class="product-list-thumb" href="/products/abrade"><img src="/images/abrade.png" /></a></td>
        <td>Abrade</td>
        <td>Revised</td>
        <td><p><span>Condition</span><strong>NM</strong></p></td>
        <td>2 left</td>
        <td>$3.10</td>
      </tr>
    </tbody>
  </table>
</div>`
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
	require.Equal(t, 3.10, result[0].Price)
	require.True(t, result[0].InStock)
	require.Equal(t, "Near Mint", result[0].Quality)
	require.Equal(t, StoreBaseURL+"/images/abrade.png", result[0].Img)
	require.Equal(t, []string{"[Revised]"}, result[0].ExtraInfo)
	require.Contains(t, result[0].Url, StoreBaseURL+"/products/abrade")
	require.Contains(t, result[0].Url, "utm_source=gishathfetch")
}
