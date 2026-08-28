package handler

import (
	"context"
	"testing"

	"mtg-price-checker-sg/controller"

	"github.com/stretchr/testify/require"
)

func TestSelectTelegramPhotoURL_PrefersCheapestStoreImage(t *testing.T) {
	photoURL := selectTelegramPhotoURL(context.Background(), []controller.Card{
		{Name: "Opt", Img: "https://placehold.co/304x424?text=Opt"},
		{Name: "Opt", Img: "https://cdn.shopify.com/s/files/card.jpg"},
	}, "Opt")
	require.Equal(t, "https://cdn.shopify.com/s/files/card.jpg", photoURL)
}

func TestSelectTelegramPhotoURL_ReturnsEmptyWithoutCandidates(t *testing.T) {
	photoURL := selectTelegramPhotoURL(context.Background(), nil, "")
	require.Empty(t, photoURL)
}
