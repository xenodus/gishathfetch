package cardscentral

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Search(t *testing.T) {
	s := NewLGS()
	result, err := s.Search(context.Background(), "lightning bolt")
	require.NoError(t, err)

	for _, card := range result {
		if card.InStock {
			require.NotEmpty(t, card.Name)
			require.Equal(t, StoreName, card.Source)
			require.Contains(t, card.Url, StoreBaseURL+"/card/")
			require.NotEmpty(t, card.Img)
			require.Greater(t, card.Price, float64(0))
			require.Equal(t, "Near Mint", card.Quality)
			require.NotEmpty(t, card.ExtraInfo)
		}
	}
}
