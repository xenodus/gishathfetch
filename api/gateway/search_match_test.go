package gateway

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCardsMatchSearch(t *testing.T) {
	cards := []Card{
		{Name: "Abrade [Foundations]"},
		{Name: "Abraded Bluffs [Outlaws of Thunder Junction]"},
	}
	require.True(t, CardsMatchSearch(cards, "Abrade"))
	require.False(t, CardsMatchSearch(cards, "Lightning Bolt"))
	require.False(t, CardsMatchSearch(nil, "Abrade"))
	require.False(t, CardsMatchSearch([]Card{{Name: "Electro's Bolt"}}, "Lightning Bolt"))
}
