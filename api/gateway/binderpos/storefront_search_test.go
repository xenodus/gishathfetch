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
			"decklist-dedicated",
			"decklist-direct",
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
			"decklist-dedicated",
			"decklist-direct",
		}
		require.Equal(t, want, got)
	})

	t.Run("without dedicated proxies", func(t *testing.T) {
		t.Setenv("USE_DEDICATED_PROXY", "false")
		got := storefrontStrategyNames("token", "example.myshopify.com")
		want := []string{
			"graphql-direct",
			"scrap-direct",
			"decklist-direct",
		}
		require.Equal(t, want, got)
	})
}
