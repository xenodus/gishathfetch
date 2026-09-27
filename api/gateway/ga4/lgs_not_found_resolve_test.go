package ga4

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveCanonicalCardName_SkipsScryfallWhenCKLookupPerformed(t *testing.T) {
	origVerify := verifyExactCardNameFunc
	t.Cleanup(func() { verifyExactCardNameFunc = origVerify })

	verifyExactCardNameFunc = func(_ context.Context, _ string) (string, error) {
		return "", errors.New("scryfall should not be called")
	}

	got, err := resolveCanonicalCardName(context.Background(), "Lightning Bolt", true, "Lightning Bolt")
	require.NoError(t, err)
	require.Equal(t, "Lightning Bolt", got)

	got, err = resolveCanonicalCardName(context.Background(), "nonsense", true, "")
	require.NoError(t, err)
	require.Equal(t, "", got)
}

func TestResolveCanonicalCardName_UsesScryfallWhenCKDisabled(t *testing.T) {
	origVerify := verifyExactCardNameFunc
	t.Cleanup(func() { verifyExactCardNameFunc = origVerify })

	verifyExactCardNameFunc = func(_ context.Context, query string) (string, error) {
		require.Equal(t, "Opt", query)
		return "Opt", nil
	}

	got, err := resolveCanonicalCardName(context.Background(), "Opt", false, "")
	require.NoError(t, err)
	require.Equal(t, "Opt", got)
}
