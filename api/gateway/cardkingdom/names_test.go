package cardkingdom

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNameLookupKeys_DoubleFacedCard(t *testing.T) {
	keys := NameLookupKeys("Jennifer Walters // The Sensational She-Hulk")
	require.Equal(t, []string{
		"jennifer walters // the sensational she-hulk",
		"jennifer walters",
		"the sensational she-hulk",
	}, keys)
}

func TestNameLookupKeys_SingleFacedCard(t *testing.T) {
	keys := NameLookupKeys("Lightning Bolt")
	require.Equal(t, []string{"lightning bolt"}, keys)
}

func TestNameLookupKeys_Empty(t *testing.T) {
	require.Nil(t, NameLookupKeys(""))
}

func TestListingNameKeys_FoilDoubleFacedCardOnlyUsesFullName(t *testing.T) {
	keys := ListingNameKeys(Listing{
		CardName: "Jennifer Walters // The Sensational She-Hulk",
		IsFoil:   true,
	})
	require.Equal(t, []string{"jennifer walters // the sensational she-hulk"}, keys)
}

func TestListingNameKeys_NonFoilDoubleFacedCardUsesAllAliases(t *testing.T) {
	keys := ListingNameKeys(Listing{
		CardName: "Jennifer Walters // The Sensational She-Hulk",
		IsFoil:   false,
	})
	require.Equal(t, []string{
		"jennifer walters // the sensational she-hulk",
		"jennifer walters",
		"the sensational she-hulk",
	}, keys)
}

func TestListingNameKeys_SingleFacedFoilUsesNameKey(t *testing.T) {
	keys := ListingNameKeys(Listing{
		CardName: "Lightning Bolt",
		IsFoil:   true,
	})
	require.Equal(t, []string{"lightning bolt"}, keys)
}

func TestDoubleFacedFaceNames(t *testing.T) {
	front, back, ok := DoubleFacedFaceNames("Jennifer Walters // The Sensational She-Hulk")
	require.True(t, ok)
	require.Equal(t, "jennifer walters", front)
	require.Equal(t, "the sensational she-hulk", back)
}

func TestPriceLookupKeys_DoubleFacedCardUsesAllAliases(t *testing.T) {
	keys := PriceLookupKeys("Jennifer Walters // The Sensational She-Hulk")
	require.Equal(t, []string{
		"jennifer walters // the sensational she-hulk",
		"jennifer walters",
		"the sensational she-hulk",
	}, keys)
}

func TestPriceLookupKeys_SingleFacedCard(t *testing.T) {
	require.Equal(t, []string{"lightning bolt"}, PriceLookupKeys("Lightning Bolt"))
}
