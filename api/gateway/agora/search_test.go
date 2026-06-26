package agora

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

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
			require.Contains(t, card.Url, StoreBaseURL+"/store/search?category="+storeCategoryMTG+"&searchfield=")
		}
	}
}

func Test_Search_FiltersMTGCategory(t *testing.T) {
	s := NewLGS()
	result, err := s.Search(context.Background(), "Bulbasaur")
	require.NoError(t, err)

	for _, card := range result {
		require.Contains(t, card.Url, "category="+storeCategoryMTG,
			"Agora product links should stay scoped to the MTG category")
		lower := strings.ToLower(card.Name)
		require.NotContains(t, lower, "pokemon",
			"Pokemon inventory should not appear when category=mtg is set")
		require.NotContains(t, lower, "holofoil",
			"Pokemon condition labels should not appear when category=mtg is set")
	}
}
