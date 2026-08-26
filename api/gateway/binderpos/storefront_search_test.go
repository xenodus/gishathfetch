package binderpos

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStorefrontStrategyOrder(t *testing.T) {
	t.Run("without graphql token", func(t *testing.T) {
		t.Setenv("USE_DEDICATED_PROXY", "true")
		got := storefrontStrategyNames("")
		want := []string{
			"scrap-direct",
			"scrap-dedicated",
		}
		require.Equal(t, want, got)
	})

	t.Run("with graphql token", func(t *testing.T) {
		t.Setenv("USE_DEDICATED_PROXY", "true")
		got := storefrontStrategyNames("token")
		want := []string{
			"graphql-direct",
			"graphql-dedicated",
			"scrap-direct",
			"scrap-dedicated",
		}
		require.Equal(t, want, got)
	})

	t.Run("without dedicated proxies", func(t *testing.T) {
		t.Setenv("USE_DEDICATED_PROXY", "false")
		got := storefrontStrategyNames("token")
		want := []string{
			"graphql-direct",
			"scrap-direct",
		}
		require.Equal(t, want, got)
	})
}
