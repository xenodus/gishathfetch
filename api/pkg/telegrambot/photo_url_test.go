package telegrambot

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_normalizePhotoURL(t *testing.T) {
	require.Equal(t, "", normalizePhotoURL(""))
	require.Equal(t, "https://cdn.example/card.jpg", normalizePhotoURL("//cdn.example/card.jpg"))
	require.Equal(t, "https://cdn.example/card.jpg", normalizePhotoURL("https://cdn.example/card.jpg"))
}

func Test_isSendableTelegramPhotoURL(t *testing.T) {
	require.False(t, isSendableTelegramPhotoURL(""))
	require.False(t, isSendableTelegramPhotoURL("/local/path.jpg"))
	require.False(t, isSendableTelegramPhotoURL("https://placehold.co/304x424?text=Opt"))
	require.False(t, isSendableTelegramPhotoURL("https://placehold.co/304x424/png?text=Opt"))
	require.False(t, isSendableTelegramPhotoURL("https://cdn.example/card.svg"))
	require.True(t, isSendableTelegramPhotoURL("https://product-images.tcgplayer.com/fit-in/437x437/12345.jpg"))
	require.True(t, isSendableTelegramPhotoURL("https://cdn.shopify.com/s/files/card.png"))
}
