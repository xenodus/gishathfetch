package handler

import (
	"context"
	"testing"

	"mtg-price-checker-sg/controller"

	"github.com/stretchr/testify/require"
)

func TestSelectTelegramPhotoURL_UsesScryfallForCardName(t *testing.T) {
	originalLookup := lookupScryfallImageURLFunc
	defer func() {
		lookupScryfallImageURLFunc = originalLookup
	}()
	lookupScryfallImageURLFunc = func(_ context.Context, cardName string) (string, error) {
		require.Equal(t, "Sol Ring", cardName)
		return "https://cards.scryfall.io/normal/front/a/b/sol-ring.jpg", nil
	}

	photoURL := selectTelegramPhotoURL(context.Background(), []controller.Card{
		{Name: "Sol Ring", Img: "https://thetcgmarketplace.com:3500/uploads/products/sol-ring.webp"},
	})
	require.Equal(t, "https://cards.scryfall.io/normal/front/a/b/sol-ring.jpg", photoURL)
}

func TestSelectTelegramPhotoURL_ReturnsEmptyWithoutCandidates(t *testing.T) {
	require.Empty(t, selectTelegramPhotoURL(context.Background(), nil))
}

func TestSelectTelegramPhotoURL_ReturnsEmptyWhenScryfallHasNoImage(t *testing.T) {
	originalLookup := lookupScryfallImageURLFunc
	defer func() {
		lookupScryfallImageURLFunc = originalLookup
	}()
	lookupScryfallImageURLFunc = func(context.Context, string) (string, error) {
		return "", nil
	}

	require.Empty(t, selectTelegramPhotoURL(context.Background(), []controller.Card{
		{Name: "Opt", Img: "https://cdn.shopify.com/s/files/card.jpg"},
	}))
}
