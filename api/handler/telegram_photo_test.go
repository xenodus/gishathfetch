package handler

import (
	"testing"

	"mtg-price-checker-sg/controller"

	"github.com/stretchr/testify/require"
)

func TestSelectTelegramPhotoURL_PrefersCheapestStoreImage(t *testing.T) {
	photoURL := selectTelegramPhotoURL([]controller.Card{
		{Name: "Opt", Img: "https://placehold.co/304x424?text=Opt"},
		{Name: "Opt", Img: "https://cdn.shopify.com/s/files/card.jpg"},
	})
	require.Equal(t, "https://cdn.shopify.com/s/files/card.jpg", photoURL)
}

func TestSelectTelegramPhotoURL_ReturnsEmptyWithoutCandidates(t *testing.T) {
	require.Empty(t, selectTelegramPhotoURL(nil))
}

func TestSelectTelegramPhotoURL_ReturnsEmptyWhenOnlyPlaceholders(t *testing.T) {
	require.Empty(t, selectTelegramPhotoURL([]controller.Card{
		{Name: "Opt", Img: "https://placehold.co/304x424?text=Opt"},
	}))
}
