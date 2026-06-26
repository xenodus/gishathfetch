package binderpos

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShouldIncludeBinderposProduct(t *testing.T) {
	require.True(t, shouldIncludeBinderposProduct("MTG Single", ""))
	require.True(t, shouldIncludeBinderposProduct("MTG Single Cards", `["MTG","Rare"]`))
	require.False(t, shouldIncludeBinderposProduct("Pokemon Single", ""))
	require.False(t, shouldIncludeBinderposProduct("Grand Archive Single", ""))
	require.False(t, shouldIncludeBinderposProduct("Lorcana Single", ""))

	// Missing product type: do not filter (legacy storefronts).
	require.True(t, shouldIncludeBinderposProduct("", ""))
}

func TestParseProductTags(t *testing.T) {
	require.Equal(t, []string{"MTG", "Rare"}, parseProductTags(`["MTG","Rare"]`))
	require.Nil(t, parseProductTags(""))
	require.Nil(t, parseProductTags("not-json"))
}
