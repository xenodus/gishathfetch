package handler

import (
	"context"
	"testing"

	"mtg-price-checker-sg/controller"

	"github.com/stretchr/testify/require"
)

func TestSelectTelegramPhotoURL_UsesStoreImageWhenSendable(t *testing.T) {
	photoURL := selectTelegramPhotoURL(context.Background(), []controller.Card{
		{
			Name: "Sol Ring",
			Img:  "https://product-images.tcgplayer.com/fit-in/437x437/12345.jpg",
		},
	})
	require.Equal(t, "https://product-images.tcgplayer.com/fit-in/437x437/12345.jpg", photoURL)
}

func TestSelectTelegramPhotoURL_ReturnsEmptyWithoutCandidates(t *testing.T) {
	require.Empty(t, selectTelegramPhotoURL(context.Background(), nil))
}

func TestSelectTelegramPhotoURL_ReturnsEmptyWhenStoreImageInvalid(t *testing.T) {
	require.Empty(t, selectTelegramPhotoURL(context.Background(), []controller.Card{
		{Name: "Opt", Img: "https://thetcgmarketplace.com:3500/uploads/products/opt.webp"},
	}))
	require.Empty(t, selectTelegramPhotoURL(context.Background(), []controller.Card{
		{Name: "Opt", Img: "https://placehold.co/304x424?text=Opt"},
	}))
}
