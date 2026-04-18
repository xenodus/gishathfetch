package moxandlotus

import (
	"context"
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
			require.Contains(t, card.Url, StoreBaseURL+"/view/")
		}
	}
}

func Test_resolveCardImageURL(t *testing.T) {
	t.Run("uses image path from API when available", func(t *testing.T) {
		actual, ok := resolveCardImageURL("SOC", "55", "https://d3nmvyqkci0c2u.cloudfront.net/SOC/card-418530-325166.jpg")
		require.True(t, ok)
		require.Equal(t, "https://d3nmvyqkci0c2u.cloudfront.net/SOC/card-418530-325166.jpg", actual)
	})

	t.Run("uses fallback image path when image path is empty", func(t *testing.T) {
		actual, ok := resolveCardImageURL("SOC", "55", "")
		require.True(t, ok)
		require.Equal(t, "https://d3nmvyqkci0c2u.cloudfront.net/SOC/055.png", actual)
	})

	t.Run("returns empty image when fallback card number is invalid", func(t *testing.T) {
		actual, ok := resolveCardImageURL("SOC", "invalid", nil)
		require.False(t, ok)
		require.Equal(t, "", actual)
	})
}
