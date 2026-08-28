package telegramphoto

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalize(t *testing.T) {
	require.Equal(t, "", Normalize(""))
	require.Equal(t, "https://cdn.example/card.jpg", Normalize("//cdn.example/card.jpg"))
	require.Equal(t, "https://cdn.example/card.jpg", Normalize("https://cdn.example/card.jpg"))
}

func TestIsSendable(t *testing.T) {
	require.False(t, IsSendable(""))
	require.False(t, IsSendable("/local/path.jpg"))
	require.False(t, IsSendable("https://placehold.co/304x424?text=Opt"))
	require.False(t, IsSendable("https://placehold.co/304x424/png?text=Opt"))
	require.False(t, IsSendable("https://cdn.example/card.svg"))
	require.False(t, IsSendable("https://thetcgmarketplace.com:3500/uploads/products/sol-ring.webp"))
	require.False(t, IsSendable("https://cdn.example/card.webp?size=large"))
	require.False(t, IsSendable("https://cdn.example/card"))
	require.False(t, IsSendable("https://cdn.example/card.bin"))

	require.True(t, IsSendable("https://product-images.tcgplayer.com/fit-in/437x437/12345.jpg"))
	require.True(t, IsSendable("https://cdn.shopify.com/s/files/card.png"))
	require.True(t, IsSendable("https://cdn.shopify.com/s/files/card.jpeg"))
	require.True(t, IsSendable("https://cdn.shopify.com/s/files/card.gif"))
	require.True(t, IsSendable("https://cdn.shopify.com/s/files/card.JPG?v=123"))
}

func TestSelect_SkipsUnsupportedFormatForNextCandidate(t *testing.T) {
	webp := "https://thetcgmarketplace.com:3500/uploads/products/sol-ring.webp"
	jpg := "https://cdn.shopify.com/s/files/sol-ring.jpg"
	require.Equal(t, jpg, Select(webp, jpg))
}

func TestSelect(t *testing.T) {
	require.Equal(t, "", Select("", "https://placehold.co/304x424?text=Opt"))
	require.Equal(t, "https://cdn.example/card.jpg", Select("", "https://placehold.co/304x424?text=Opt", "https://cdn.example/card.jpg"))
	require.Equal(t, "https://cdn.example/cheapest.jpg", Select("https://cdn.example/cheapest.jpg", "https://cdn.example/other.jpg"))
}
