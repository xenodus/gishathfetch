package shopifysuggest

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseNameAndFoil(t *testing.T) {
	name, isFoil := ParseNameAndFoil("[Foil] Cauldron of Essence")
	require.Equal(t, "Cauldron of Essence", name)
	require.True(t, isFoil)

	name, isFoil = ParseNameAndFoil("Cauldron of Essence")
	require.Equal(t, "Cauldron of Essence", name)
	require.False(t, isFoil)
}

func TestIsMagicProduct(t *testing.T) {
	require.True(t, IsMagicProduct("MTG Single Cards", "Magic the Gathering", []string{"MTG"}))
	require.True(t, IsMagicProduct("MTG Single", "Magic: The Gathering", nil))
	require.True(t, IsMagicProduct("Magic the Gathering Booster Box", "Magic the Gathering Sealed Products", nil))
	require.True(t, IsMagicProduct("", "", []string{"New Arrival", "MTG"}))
	require.False(t, IsMagicProduct("Flesh And Blood Single Cards", "Flesh and Blood", []string{"Common"}))
	require.False(t, IsMagicProduct("Grand Archive Single Cards", "Grand Archive", nil))
	require.False(t, IsMagicProduct("Lorcana Single", "Disney Lorcana", nil))
}

func TestSetFromTags(t *testing.T) {
	require.Equal(t, "Foundations", SetFromTags([]string{"Foundations", "MTG", "New Arrival", "Rare"}))
	require.Equal(t, "Dominaria United Commander", SetFromTags([]string{"Dominaria United Commander", "MTG", "Mythic", "New Arrival"}))

	// Ambiguous (more than one candidate) returns empty to protect data integrity.
	require.Equal(t, "", SetFromTags([]string{"Mystical Archive", "Secrets of Strixhaven", "MTG", "Rare"}))

	// No candidate set tag.
	require.Equal(t, "", SetFromTags([]string{"MTG", "New Arrival", "Rare"}))
}

func TestStripTrailingSetAndExtractSetName(t *testing.T) {
	title := "Abrade [Secrets of Strixhaven: Mystical Archive]"
	require.Equal(t, "Abrade", stripTrailingSet(title))
	require.Equal(t, "Secrets of Strixhaven: Mystical Archive", extractSetName(title))
	require.Equal(t, "Lightning Bolt", stripTrailingSet("Lightning Bolt"))
}

func TestBinderposQueryValues(t *testing.T) {
	values := BinderposQueryValues("donald")
	require.Equal(t, "donald", values.Get("q"))
	require.Equal(t, "product", values.Get("resources[type]"))
	require.Equal(t, "last", values.Get("resources[options][unavailable_products]"))
	require.Equal(t, "title,variants.title,product_type", values.Get("resources[options][fields]"))
}
