package mtgasia

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
	"time"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/gatewaytest"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

func init() {
	_ = godotenv.Load("../../.env")
}

func Test_Search(t *testing.T) {
	s := NewLGS()
	result, err := s.Search(context.Background(), "Abrade")
	gatewaytest.RequireSearchOrProbe(t, err, result, gatewaytest.CardExpect{
		URLContains: StoreBaseURL + "/card/",
		Source:      StoreName,
	}, func(t *testing.T, ctx context.Context) {
		requireBase44InventoryAPI(t, ctx)
	})
}

func requireBase44InventoryAPI(t *testing.T, ctx context.Context) {
	t.Helper()

	filter := url.Values{
		"q":     {`{"qty_available":{"$gt":0}}`},
		"limit": {"1"},
	}
	apiURL := base44ServerURL + "/api/apps/" + base44AppID + "/entities/InventoryItem?" + filter.Encode()

	resp, err := gateway.DoOutboundGET(ctx, apiURL, gateway.OutboundRequestOptions{
		Style: gateway.OutboundStyleJSON,
	}, 20*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := gateway.ReadResponseBody(resp)
	require.NoError(t, err)

	var items []inventoryItem
	require.NoError(t, json.Unmarshal(body, &items))
	require.NotEmpty(t, items)
}
