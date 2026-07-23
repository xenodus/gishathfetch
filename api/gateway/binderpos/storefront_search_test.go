package binderpos

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStorefrontStrategyOrder(t *testing.T) {
	t.Run("without graphql token", func(t *testing.T) {
		got := storefrontStrategyNames("", "example.myshopify.com")
		want := []string{
			"scrap-dedicated",
			"scrap-direct",
			"scrap-dynamic",
			"decklist-dedicated",
			"decklist-direct",
			"decklist-dynamic",
		}
		require.Equal(t, want, got)
	})

	t.Run("with graphql token", func(t *testing.T) {
		got := storefrontStrategyNames("token", "example.myshopify.com")
		want := []string{
			"graphql-dedicated",
			"graphql-direct",
			"scrap-dedicated",
			"scrap-direct",
			"scrap-dynamic",
			"decklist-dedicated",
			"decklist-direct",
			"decklist-dynamic",
		}
		require.Equal(t, want, got)
	})
}
