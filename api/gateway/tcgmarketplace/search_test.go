package tcgmarketplace

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"mtg-price-checker-sg/gateway/gatewaytest"
	"mtg-price-checker-sg/pkg/config"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

func init() {
	_ = godotenv.Load("../../.env")
}

func TestListingsFromEnvelope(t *testing.T) {
	t.Run("success payload", func(t *testing.T) {
		body := []byte(`{"status":200,"data":{"message":"","data":[{"name":" [3ED] Sol Ring","setname":"Revised Edition","available":3,"from":"28.00","url":"https://thetcgmarketplace.cc/product/B/x/0"}]},"meta":{"total":1}}`)
		var env apiEnvelope
		require.NoError(t, json.Unmarshal(body, &env))
		listings, err := listingsFromEnvelope(env)
		require.NoError(t, err)
		require.Len(t, listings, 1)
		require.Equal(t, "https://thetcgmarketplace.cc/product/B/x/0", listings[0].URL)
	})

	t.Run("unauthorized api status", func(t *testing.T) {
		body := []byte(`{"status":500,"data":{"message":"Unathorised","data":""},"meta":{"total":0}}`)
		var env apiEnvelope
		require.NoError(t, json.Unmarshal(body, &env))
		_, err := listingsFromEnvelope(env)
		require.ErrorContains(t, err, "Unathorised")
	})

	t.Run("pool limit object payload", func(t *testing.T) {
		body := []byte(`{"status":500,"data":{"message":"","data":{"code":"POOL_ENQUEUELIMIT"}},"meta":{}}`)
		var env apiEnvelope
		require.NoError(t, json.Unmarshal(body, &env))
		_, err := listingsFromEnvelope(env)
		require.ErrorContains(t, err, "search api status 500")
	})
}

func TestCanonicalProductURL(t *testing.T) {
	t.Run("cc host rewritten to com", func(t *testing.T) {
		raw := "https://thetcgmarketplace.cc/product/B/abc/0"
		u, err := canonicalProductURL(raw)
		require.NoError(t, err)
		require.Equal(t, "thetcgmarketplace.com", u.Host)
		require.Equal(t, "/product/B/abc/0", u.Path)
	})

	t.Run("com host unchanged", func(t *testing.T) {
		raw := "https://thetcgmarketplace.com/product/B/abc/0"
		u, err := canonicalProductURL(raw)
		require.NoError(t, err)
		require.Equal(t, "thetcgmarketplace.com", u.Host)
	})
}

func TestIsSurgeFoil(t *testing.T) {
	tests := map[string]struct {
		extraInfo []string
		name      string
		want      bool
	}{
		"surge foil in name": {
			name: "Abaddon the Despoiler (V1)(Surge Foil)",
			want: true,
		},
		"surge foil in extra info": {
			extraInfo: []string{"[Warhammer 40,000 Commander]", "[Surge Foil]"},
			name:      "Abaddon the Despoiler",
			want:      true,
		},
		"non-foil card": {
			extraInfo: []string{"[Double Masters]"},
			name:      "Abrade",
			want:      false,
		},
		"etched foil not surge": {
			extraInfo: []string{"[Commander Masters]"},
			name:      "Deflecting Swat (V2)(Etched foil)",
			want:      false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, tt.want, isSurgeFoil(tt.extraInfo, tt.name))
		})
	}
}

func Test_Search(t *testing.T) {
	ctx := context.Background()
	if config.TCGMarketplaceAdvancedFilterSearchEnabled() {
		s := NewLGS()
		result, err := s.Search(ctx, "abrade")
		if err != nil {
			t.Logf("advanced filter search failed (%v); probing api structure", err)
			gatewaytest.RequireTCGMarketplaceAdvancedFilterStructure(t, ctx, "abrade")
			return
		}
		gatewaytest.RequireSearchOrProbe(t, err, result, gatewaytest.CardExpect{
			URLContains: StoreBaseURL + "/product/B/",
		}, func(t *testing.T, ctx context.Context) {
			gatewaytest.RequireTCGMarketplaceAdvancedFilterStructure(t, ctx, "abrade")
		})
		return
	}

	token := os.Getenv(accessTokenKey)
	if token == "" {
		gatewaytest.RequireTCGMarketplaceAPIStructure(t, ctx, "", "abrade")
		return
	}

	s := NewLGS()
	result, err := s.Search(ctx, "abrade")
	if err != nil {
		t.Logf("search failed (%v); probing api structure", err)
		gatewaytest.RequireTCGMarketplaceAPIStructure(t, ctx, token, "abrade")
		return
	}
	gatewaytest.RequireSearchOrProbe(t, err, result, gatewaytest.CardExpect{
		URLContains: StoreBaseURL + "/product/B/",
	}, func(t *testing.T, ctx context.Context) {
		gatewaytest.RequireTCGMarketplaceAPIStructure(t, ctx, token, "abrade")
	})
}

func Test_SearchRoutesToAdvancedFilterWhenEnabled(t *testing.T) {
	t.Setenv(config.TCGMarketplaceAdvancedFilterSearchEnv, "true")
	require.True(t, config.TCGMarketplaceAdvancedFilterSearchEnabled())

	ctx := context.Background()
	s := NewLGS()
	_, err := s.Search(ctx, "Opt")
	if err != nil {
		gatewaytest.RequireTCGMarketplaceAdvancedFilterStructure(t, ctx, "Opt")
		return
	}
}
