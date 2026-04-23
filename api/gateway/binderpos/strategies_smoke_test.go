package binderpos

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/util"
)

// Test_EachSearchStrategyReturnsResults exercises the three real network paths
// used by Search (dedicated API → shared API → shared-proxy scraper) without
// fallback masking which layer produced cards.
func Test_EachSearchStrategyReturnsResults(t *testing.T) {
	_ = godotenv.Load("../../.env")

	prev := shouldUseDecklistEndpoint
	shouldUseDecklistEndpoint = func() bool { return false }
	t.Cleanup(func() { shouldUseDecklistEndpoint = prev })

	var tc *binderposStoreSearchCase
	for i := range binderposStoreSearchCases() {
		c := binderposStoreSearchCases()[i]
		if c.storeName == "OneMtg" {
			tc = &c
			break
		}
	}
	require.NotNil(t, tc, "expected OneMtg in binderposStoreSearchCases()")

	if len(util.GetDedicatedProxyURLs()) == 0 {
		t.Skip("DEDICATED_PROXY_* not configured; cannot test api-dedicated")
	}
	if strings.TrimSpace(os.Getenv("PROXY_URL")) == "" {
		t.Skip("PROXY_URL not configured; cannot test api-shared or scrap-shared")
	}

	ctx := context.Background()
	name := "OneMtg (smoke: dedicated storefront API over dedicated proxy)"
	t.Run(name, func(t *testing.T) {
		cards, err := searchByStorefrontAPI(ctx, tc.scrapVariant, tc.storeName, tc.baseURL, tc.query)
		assertStrategyCards(t, cards, err, name)
	})

	name = "OneMtg (smoke: shared storefront API over PROXY_URL)"
	t.Run(name, func(t *testing.T) {
		cards, err := searchByStorefrontAPISharedProxy(ctx, tc.scrapVariant, tc.storeName, tc.baseURL, tc.query)
		assertStrategyCards(t, cards, err, name)
	})

	name = "OneMtg (smoke: shared proxy scraper)"
	t.Run(name, func(t *testing.T) {
		impl, ok := New().(*impl)
		require.True(t, ok, "New() should return *impl")
		// colly uses binderposAttemptTimeout (2s), which is occasionally tight for full HTML
		// through a shared proxy under parallel load; retry a few times on transport timeouts.
		var cards []gateway.Card
		var err error
		for attempt := 1; attempt <= 3; attempt++ {
			cards, err = impl.scrapSharedProxy(ctx, tc.scrapVariant, tc.storeName, tc.baseURL, tc.searchURL, tc.query)
			if err == nil {
				break
			}
			if attempt < 3 && (strings.Contains(err.Error(), "deadline exceeded") || strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "i/o timeout")) {
				time.Sleep(2 * time.Second)
				continue
			}
			break
		}
		assertStrategyCards(t, cards, err, name)
	})
}

func assertStrategyCards(t *testing.T, cards []gateway.Card, err error, strategyName string) {
	t.Helper()
	require.NoError(t, err, strategyName)
	require.NotEmpty(t, cards, strategyName)

	for _, c := range cards {
		require.NotEmpty(t, c.Name, strategyName)
		require.NotEmpty(t, c.Url, strategyName)
		require.NotEmpty(t, c.Img, strategyName)
		require.NotEmpty(t, c.Source, strategyName)
		require.Greater(t, c.Price, float64(0), strategyName)
	}
}
