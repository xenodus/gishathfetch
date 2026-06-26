package hideyoshi

import (
	"context"
	"strings"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

func init() {
	_ = godotenv.Load("../../.env")
}

func Test_Search(t *testing.T) {
	s := NewLGS()
	result, err := s.Search(context.Background(), "Abrade")
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

func Test_Search_ExcludesPokemonInventory(t *testing.T) {
	s := NewLGS()
	result, err := s.Search(context.Background(), "Bulbasaur")
	require.NoError(t, err)

	for _, card := range result {
		lower := strings.ToLower(card.Name)
		require.NotContains(t, lower, "mega evolution: base set",
			"Pokemon Bulbasaur inventory should not appear in MTG search results")
		require.NotContains(t, lower, "near mint holofoil",
			"Pokemon condition labels should not appear in MTG search results")
	}
}
