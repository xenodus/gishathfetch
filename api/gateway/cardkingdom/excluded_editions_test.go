package cardkingdom

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExcludedCKPriceEdition(t *testing.T) {
	require.True(t, excludedCKPriceEdition("World Championship Decks 2004"))
	require.True(t, excludedCKPriceEdition("World Championship Decks 1997"))
	require.True(t, excludedCKPriceEdition("World Championships"))
	require.True(t, excludedCKPriceEdition("  World Championships  "))
	require.True(t, excludedCKPriceEdition("World Championship Promos"))

	require.False(t, excludedCKPriceEdition("Fourth Edition"))
	require.False(t, excludedCKPriceEdition(""))
}
