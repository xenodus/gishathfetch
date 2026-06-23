package fyendalhobby

import (
	"context"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

func init() {
	_ = godotenv.Load("../../.env")
}

func Test_Search(t *testing.T) {
	s := NewLGS()
	result, err := s.Search(context.Background(), "Cauldron of Essence")
	require.NoError(t, err)
	require.True(t, len(result) > 0)

	for _, card := range result {
		if card.InStock {
			require.NotEmpty(t, card.Name)
			require.NotEmpty(t, card.Source)
			require.NotEmpty(t, card.Url)
			require.NotEmpty(t, card.Img)
			require.NotEmpty(t, card.Price)
			require.Contains(t, card.Url, StoreBaseURL+"/products/")
		}
	}
}

func TestParseNameAndFoil(t *testing.T) {
	name, isFoil := parseNameAndFoil("[Foil] Cauldron of Essence")
	require.Equal(t, "Cauldron of Essence", name)
	require.True(t, isFoil)

	name, isFoil = parseNameAndFoil("Cauldron of Essence")
	require.Equal(t, "Cauldron of Essence", name)
	require.False(t, isFoil)
}

func TestIsMagicProduct(t *testing.T) {
	require.True(t, isMagicProduct("MTG Single Cards", "Magic the Gathering", []string{"MTG"}))
	require.True(t, isMagicProduct("Magic the Gathering Booster Box", "Magic the Gathering Sealed Products", nil))
	require.True(t, isMagicProduct("", "", []string{"New Arrival", "MTG"}))
	require.False(t, isMagicProduct("Flesh And Blood Single Cards", "Flesh and Blood", []string{"Common"}))
	require.False(t, isMagicProduct("Grand Archive Single Cards", "Grand Archive", nil))
}

func TestSetFromTags(t *testing.T) {
	require.Equal(t, "Foundations", setFromTags([]string{"Foundations", "MTG", "New Arrival", "Rare"}))
	require.Equal(t, "Dominaria United Commander", setFromTags([]string{"Dominaria United Commander", "MTG", "Mythic", "New Arrival"}))

	// Ambiguous (more than one candidate) returns empty to protect data integrity.
	require.Equal(t, "", setFromTags([]string{"Mystical Archive", "Secrets of Strixhaven", "MTG", "Rare"}))

	// No candidate set tag.
	require.Equal(t, "", setFromTags([]string{"MTG", "New Arrival", "Rare"}))
}
