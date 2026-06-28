package cardscentral

import (
	"context"
	"testing"

	"mtg-price-checker-sg/gateway/gatewaytest"

	"github.com/stretchr/testify/require"
)

func Test_Search(t *testing.T) {
	s := NewLGS()
	result, err := s.Search(context.Background(), "lightning bolt")
	require.NoError(t, err)

	if len(result) > 0 {
		for _, card := range result {
			if card.InStock {
				require.NotEmpty(t, card.Name)
				require.Equal(t, StoreName, card.Source)
				require.Contains(t, card.Url, StoreBaseURL+"/card/")
				require.NotEmpty(t, card.Img)
				require.Greater(t, card.Price, float64(0))
				require.NotEmpty(t, card.Quality)
				require.NotEmpty(t, card.ExtraInfo)
			}
		}
		return
	}

	gatewaytest.RequireCardsCentralAPIStructure(t, context.Background(), StoreBaseURL, "lightning bolt")
}
