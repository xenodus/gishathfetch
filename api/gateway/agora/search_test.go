package agora

import (
	"context"
	"net/url"
	"strings"
	"testing"

	"mtg-price-checker-sg/gateway/gatewaytest"
	"mtg-price-checker-sg/gateway/util"
	"mtg-price-checker-sg/pkg/config"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/require"
)

func Test_ParseStoreItemSkipsOutOfStock(t *testing.T) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(outOfStockStoreItemHTML))
	require.NoError(t, err)

	apiURL, err := url.Parse(StoreBaseURL + StoreSearchPath + "?category=mtg&searchfield=Abrade")
	require.NoError(t, err)

	card, ok := parseStoreItem(doc.Find("div.store-item").First(), StoreName, apiURL, "Abrade")
	require.False(t, ok)
	require.Empty(t, card.Name)
}

func Test_ParseStoreItemKeepsInStock(t *testing.T) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(inStockStoreItemHTML))
	require.NoError(t, err)

	apiURL, err := url.Parse(StoreBaseURL + StoreSearchPath + "?category=mtg&searchfield=Abrade")
	require.NoError(t, err)

	card, ok := parseStoreItem(doc.Find("div.store-item").First(), StoreName, apiURL, "Abrade")
	require.True(t, ok)
	require.Equal(t, "Abrade FOIL", card.Name)
	require.True(t, card.InStock)
	require.True(t, card.IsFoil)
	require.Equal(t, 2.0, card.Price)
	require.Equal(t, "[DMU]", card.ExtraInfo[0])
	require.Contains(t, card.Url, "utm_source=")
}

func Test_Search(t *testing.T) {
	skipLiveAgoraSearchUnlessResidential(t)

	s := NewLGS()
	result, err := s.Search(context.Background(), "Abrade")
	gatewaytest.RequireSearchOrProbe(t, err, result, gatewaytest.CardExpect{
		URLContains:    StoreBaseURL + "/store/search?category=" + storeCategoryMTG + "&searchfield=",
		RequireInStock: true,
	}, func(t *testing.T, ctx context.Context) {
		gatewaytest.RequireAgoraSearchStructure(t, ctx, StoreBaseURL, StoreSearchPath, storeCategoryMTG, "Abrade")
	})
}

func Test_Search_FiltersMTGCategory(t *testing.T) {
	skipLiveAgoraSearchUnlessResidential(t)

	s := NewLGS()
	result, err := s.Search(context.Background(), "Bulbasaur")
	require.NoError(t, err)

	for _, card := range result {
		require.True(t, card.InStock, "gateway should only return in-stock listings")
		require.Contains(t, card.Url, "category="+storeCategoryMTG,
			"Agora product links should stay scoped to the MTG category")
		lower := strings.ToLower(card.Name)
		require.NotContains(t, lower, "pokemon",
			"Pokemon inventory should not appear when category=mtg is set")
		require.NotContains(t, lower, "holofoil",
			"Pokemon condition labels should not appear when category=mtg is set")
	}
}

func skipLiveAgoraSearchUnlessResidential(t *testing.T) {
	t.Helper()
	if _, ok := util.GetResidentialProxyURL(); ok {
		return
	}
	t.Skipf("set %s to run live Agora search checks (datacenter proxies are blocked by Cloudflare)", config.ResidentialProxyEnv)
}
