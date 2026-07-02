package cardkingdom

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNameLookupKeys_DoubleFacedCard(t *testing.T) {
	keys := NameLookupKeys("Jennifer Walters // The Sensational She-Hulk")
	require.Equal(t, []string{
		"jennifer walters // the sensational she-hulk",
		"jennifer walters",
		"the sensational she-hulk",
	}, keys)
}

func TestNameLookupKeys_SingleFacedCard(t *testing.T) {
	keys := NameLookupKeys("Lightning Bolt")
	require.Equal(t, []string{"lightning bolt"}, keys)
}

func TestNameLookupKeys_Empty(t *testing.T) {
	require.Nil(t, NameLookupKeys(""))
}
