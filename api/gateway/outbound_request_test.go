package gateway

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrepareOutboundRequest_CKMinimalPricelistHeaders(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "https://api.cardkingdom.com/api/v2/pricelist", nil)
	require.NoError(t, err)

	err = PrepareOutboundRequest(context.Background(), req, OutboundRequestOptions{
		Accept:         "application/json",
		SkipWebBotAuth: true,
	})
	require.NoError(t, err)
	require.Equal(t, "application/json", req.Header.Get("Accept"))
	require.NotEmpty(t, req.Header.Get("User-Agent"))
	require.Empty(t, req.Header.Get("Origin"))
	require.Empty(t, req.Header.Get("Referer"))
	require.Empty(t, req.Header.Get("Signature"))
	require.Empty(t, req.Header.Get("Signature-Input"))
}

func TestPrepareOutboundRequest_JSONStyleSetsOrigin(t *testing.T) {
	storeBase, err := url.Parse("https://www.cardkingdom.com/")
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, "https://api.cardkingdom.com/api/v2/pricelist", nil)
	require.NoError(t, err)

	err = PrepareOutboundRequest(context.Background(), req, OutboundRequestOptions{
		Style:          OutboundStyleJSON,
		StoreBase:      storeBase,
		SkipWebBotAuth: true,
	})
	require.NoError(t, err)
	require.Equal(t, "https://www.cardkingdom.com/", req.Header.Get("Referer"))
	require.Equal(t, "https://www.cardkingdom.com", req.Header.Get("Origin"))
}
