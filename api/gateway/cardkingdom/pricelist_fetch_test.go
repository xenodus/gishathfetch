package cardkingdom

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFetchCheapestFromCKPricelist_FromTestServer(t *testing.T) {
	originalFetch := fetchPricelistBody
	defer func() { fetchPricelistBody = originalFetch }()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"meta":{"created_at":"2026-07-02 12:10:11"},
			"data":[
				{
					"name":"Jennifer Walters",
					"edition":"Marvel Super Heroes",
					"price_retail":10.99,
					"qty_retail":8,
					"url":"mtg/marvel-super-heroes/jennifer-walters",
					"is_foil":"false"
				}
			]
		}`))
	}))
	defer server.Close()

	fetchPricelistBody = func(ctx context.Context) ([]byte, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
		if err != nil {
			return nil, err
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		return io.ReadAll(resp.Body)
	}

	cheapest, err := fetchCheapestFromCKPricelist(context.Background())
	require.NoError(t, err)
	require.InDelta(t, 10.99, cheapest["jennifer walters"].PriceUsd, 0.001)
}

func TestMergeCheapestListings_PrefersCheaperFaceNameListing(t *testing.T) {
	updatedAt := time.Now().UTC().Format(time.RFC3339)
	listings := make(map[string]Listing)
	considerCheapestListing(listings, Listing{
		CardName:  "Jennifer Walters // The Sensational She-Hulk",
		Edition:   "Marvel Super Heroes Variants",
		PriceUsd:  69.99,
		URL:       "https://www.cardkingdom.com/mtg/marvel-super-heroes-variants/jennifer-walters-0355-borderless-foil",
		IsFoil:    true,
		UpdatedAt: updatedAt,
	})

	mergeCheapestListings(listings, map[string]Listing{
		"jennifer walters": {
			CardName:  "Jennifer Walters",
			Edition:   "Marvel Super Heroes",
			PriceUsd:  10.99,
			URL:       "https://www.cardkingdom.com/mtg/marvel-super-heroes/jennifer-walters",
			UpdatedAt: updatedAt,
		},
	})

	require.InDelta(t, 10.99, listings["jennifer walters"].PriceUsd, 0.001)
	require.InDelta(t, 69.99, listings["jennifer walters // the sensational she-hulk"].PriceUsd, 0.001)
	_, hasBackFace := listings["the sensational she-hulk"]
	require.False(t, hasBackFace)
}

func TestFetchPricelistBodyFromURL_RejectsCloudflareChallenge(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html>Just a moment... cloudflare</html>"))
	}))
	defer server.Close()

	_, err := fetchPricelistBodyFromURL(context.Background(), server.URL)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cloudflare challenge")
}

func TestFetchPricelistBodyFromURL_ReportsOutboundFailure(t *testing.T) {
	for i := 1; i <= 7; i++ {
		t.Setenv(fmt.Sprintf("DEDICATED_PROXY_%d", i), "")
	}
	t.Setenv("DYNAMIC_PROXY", "")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	defer server.Close()

	_, err := fetchPricelistBodyFromURL(context.Background(), server.URL)
	require.Error(t, err)
	require.Contains(t, err.Error(), "outbound request failed")
	require.Contains(t, err.Error(), "status 403")
}

func TestConsiderCheapestListing_FoilDoubleFacedDoesNotPolluteFaceAlias(t *testing.T) {
	updatedAt := time.Now().UTC().Format(time.RFC3339)
	listings := make(map[string]Listing)
	considerCheapestListing(listings, Listing{
		CardName:  "Jennifer Walters // The Sensational She-Hulk",
		Edition:   "Marvel Super Heroes",
		PriceUsd:  69.99,
		URL:       "https://mtgjson.com/links/d8187fd1eef32412",
		IsFoil:    true,
		UpdatedAt: updatedAt,
	})

	require.InDelta(t, 69.99, listings["jennifer walters // the sensational she-hulk"].PriceUsd, 0.001)
	_, hasFrontFace := listings["jennifer walters"]
	require.False(t, hasFrontFace)
	_, hasBackFace := listings["the sensational she-hulk"]
	require.False(t, hasBackFace)
}
