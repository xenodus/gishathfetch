package gatewaytest

import (
	"context"
	"testing"

	"mtg-price-checker-sg/gateway"

	"github.com/stretchr/testify/require"
)

// CardExpect configures field assertions for cards returned by a live search.
type CardExpect struct {
	URLContains    string
	Source         string
	RequireInStock bool
}

// AssertCards validates gateway.Card fields for each in-stock listing.
func AssertCards(t *testing.T, cards []gateway.Card, exp CardExpect) {
	t.Helper()

	for _, card := range cards {
		if exp.RequireInStock {
			require.True(t, card.InStock, "gateway should only return in-stock listings")
		}
		if !card.InStock {
			continue
		}

		require.NotEmpty(t, card.Name)
		require.NotEmpty(t, card.Source)
		if exp.Source != "" {
			require.Equal(t, exp.Source, card.Source)
		}
		require.NotEmpty(t, card.Url)
		require.NotEmpty(t, card.Img)
		require.Greater(t, card.Price, float64(0))
		if exp.URLContains != "" {
			require.Contains(t, card.Url, exp.URLContains)
		}
	}
}

// RequireSearchOrProbe runs a live search and validates returned cards when
// inventory is present. When the search succeeds with no cards, probe verifies
// the upstream response or page structure is still scrappable.
func RequireSearchOrProbe(
	t *testing.T,
	err error,
	cards []gateway.Card,
	exp CardExpect,
	probe func(t *testing.T, ctx context.Context),
) {
	t.Helper()
	require.NoError(t, err)
	if len(cards) > 0 {
		AssertCards(t, cards, exp)
		return
	}
	require.NotNil(t, probe, "search returned no cards; configure a structure probe to verify upstream is still scrappable")
	probe(t, context.Background())
}
