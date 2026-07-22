package arcanesanctum

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
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
	result, err := s.Search(context.Background(), "Reanimate")
	gatewaytest.RequireSearchOrProbe(t, err, result, gatewaytest.CardExpect{
		URLContains: StoreBaseURL + "/products/",
		Source:      StoreName,
	}, func(t *testing.T, ctx context.Context) {
		requireArcaneSanctumStorefrontGraphQL(t, ctx)
	})
}

// requireArcaneSanctumStorefrontGraphQL verifies the Shopify Storefront GraphQL
// search endpoint still accepts the configured token. Arcane Sanctum's custom
// theme does not expose BinderPOS HTML scrape markers, so GraphQL is the
// reliable upstream probe when live inventory is empty.
func requireArcaneSanctumStorefrontGraphQL(t *testing.T, ctx context.Context) {
	t.Helper()

	payload := []byte(`{"query":"query ($q: String!) { search(query: $q, first: 1, types: PRODUCT) { edges { node { ... on Product { title } } } } }","variables":{"q":"mtg"}}`)
	resp, err := gateway.DoOutboundRoundTrip(ctx, gateway.OutboundRequestOptions{
		Style: gateway.OutboundStyleJSON,
	}, 20*time.Second, func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, StoreBaseURL+"/api/2024-10/graphql.json", bytes.NewReader(payload))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Shopify-Storefront-Access-Token", StoreStorefrontAccessToken)
		return req, nil
	})
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := gateway.ReadResponseBody(resp)
	require.NoError(t, err)

	var parsed struct {
		Data *struct {
			Search *struct {
				Edges []struct{} `json:"edges"`
			} `json:"search"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	require.NoError(t, json.Unmarshal(body, &parsed))
	require.Empty(t, parsed.Errors, "storefront graphql errors: %v", parsed.Errors)
	require.NotNil(t, parsed.Data)
	require.NotNil(t, parsed.Data.Search)
}
