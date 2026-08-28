package telegrambot

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFormatPricePrompt_WithUsername(t *testing.T) {
	text, parseMode := formatPricePrompt(&User{ID: 1, Username: "alice"})
	require.Equal(t, "@alice, Enter a card name to search", text)
	require.Empty(t, parseMode)
}

func TestFormatPricePrompt_WithoutUsername(t *testing.T) {
	text, parseMode := formatPricePrompt(&User{ID: 42, FirstName: "Bob"})
	require.Equal(t, `<a href="tg://user?id=42">Bob</a>, Enter a card name to search`, text)
	require.Equal(t, "HTML", parseMode)
}

func TestIsPricePrompt(t *testing.T) {
	require.True(t, isPricePrompt("@alice, Enter a card name to search"))
	require.True(t, isPricePrompt(`<a href="tg://user?id=42">Bob</a>, Enter a card name to search`))
	require.False(t, isPricePrompt("Searching for Opt…"))
}
