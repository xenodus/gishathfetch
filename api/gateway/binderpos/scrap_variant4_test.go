package binderpos

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFyendalSearchQuery(t *testing.T) {
	require.Equal(t, `product_type:"MTG Single Cards" AND Abrade`, fyendalSearchQuery("Abrade"))
}

func TestParseFyendalNameAndFoil(t *testing.T) {
	name, isFoil := parseFyendalNameAndFoil("[Foil] Abrade")
	require.Equal(t, "Abrade", name)
	require.True(t, isFoil)

	name, isFoil = parseFyendalNameAndFoil("Lightning Bolt")
	require.Equal(t, "Lightning Bolt", name)
	require.False(t, isFoil)
}
