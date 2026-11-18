package mtgasia

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_cardPathSlug_matchesStorefront(t *testing.T) {
	require.Equal(t, "innistrad-remastered-abrade", cardPathSlug("Abrade", "Innistrad Remastered"))
	require.Equal(t, "abrade", cardPathSlug("Abrade", ""))
	require.Equal(t, "reality-fracture-aerid-konstrari", cardPathSlug("Aerid Konstrari", "Reality Fracture"))
}

func Test_storefrontSellable(t *testing.T) {
	require.True(t, storefrontSellable(cardSet{}, false))
	require.True(t, storefrontSellable(cardSet{Name: "Foundations", Hidden: false}, true))
	require.False(t, storefrontSellable(cardSet{Name: "Commander Masters", Hidden: true}, true))
}

func Test_subCategoryExtraInfoLabel(t *testing.T) {
	require.Equal(t, "[FDN Commander]", subCategoryExtraInfoLabel("FDN Commander"))
	require.Equal(t, "[FRC Non-Foil]", subCategoryExtraInfoLabel("FRC Non-Foil"))
	require.Empty(t, subCategoryExtraInfoLabel(""))
}

func Test_mapInventoryItem_includesSubCategoryInExtraInfo(t *testing.T) {
	card, ok := mapInventoryItem(inventoryItem{
		ID:           "6ab21f0c8a7dddcb3960518d",
		Name:         "Swords to Plowshares",
		SubCategory:  "FDN Commander",
		Condition:    "NM",
		SellingPrice: 2.9,
		QtyAvailable: 2,
	}, "Reality Fracture")
	require.True(t, ok)
	require.Equal(t, []string{"[FDN Commander]"}, card.ExtraInfo)
	require.Contains(t, card.Url, "6ab21f0c8a7dddcb3960518d")
}

func Test_buildCardSinglePageURL(t *testing.T) {
	got, err := buildCardSinglePageURL(
		"https://mtg-asia.com",
		"6ac33f1368fe2a219f23d9d0",
		"Abrade",
		"Innistrad Remastered",
	)
	require.NoError(t, err)
	require.Contains(t, got, "https://mtg-asia.com/card/innistrad-remastered-abrade/6ac33f1368fe2a219f23d9d0")
	require.Contains(t, got, "utm_source=")
}
