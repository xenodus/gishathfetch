package ckprice

import (
	"context"
	"testing"
	"time"

	"mtg-price-checker-sg/gateway/cardkingdom"

	"github.com/stretchr/testify/require"
)

type mockStore struct {
	listing  *cardkingdom.Listing
	listings map[string]*cardkingdom.Listing
	getKeys  []string
}

func (m *mockStore) GetByNameKey(_ context.Context, nameKey string) (*cardkingdom.Listing, error) {
	m.getKeys = append(m.getKeys, nameKey)
	if m.listings != nil {
		return m.listings[nameKey], nil
	}
	return m.listing, nil
}

func (m *mockStore) PutAll(_ context.Context, _ map[string]cardkingdom.Listing) (string, error) {
	return "", nil
}

func TestGetLatestPrice_UsesVerifiedName(t *testing.T) {
	originalVerify := verifyCardNameFunc
	defer func() { verifyCardNameFunc = originalVerify }()

	verifyCardNameFunc = func(_ context.Context, _ string) (string, error) {
		return "Lightning Bolt", nil
	}

	store := &mockStore{
		listing: &cardkingdom.Listing{
			CardName:  "Lightning Bolt",
			PriceUsd:  1.49,
			UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		},
	}

	listing, err := GetLatestPrice(context.Background(), store, "lightning bolt")
	require.NoError(t, err)
	require.Equal(t, []string{"lightning bolt"}, store.getKeys)
	require.Equal(t, 1.49, listing.PriceUsd)
}

func TestGetLatestPrice_InvalidCard(t *testing.T) {
	originalVerify := verifyCardNameFunc
	defer func() { verifyCardNameFunc = originalVerify }()

	verifyCardNameFunc = func(_ context.Context, _ string) (string, error) {
		return "", nil
	}

	listing, err := GetLatestPrice(context.Background(), &mockStore{}, "asdfasdf")
	require.NoError(t, err)
	require.Nil(t, listing)
}

func TestGetLatestPrice_StaleListing(t *testing.T) {
	originalVerify := verifyCardNameFunc
	defer func() { verifyCardNameFunc = originalVerify }()

	verifyCardNameFunc = func(_ context.Context, _ string) (string, error) {
		return "Lightning Bolt", nil
	}

	store := &mockStore{
		listing: &cardkingdom.Listing{
			CardName:  "Lightning Bolt",
			PriceUsd:  1.49,
			UpdatedAt: time.Now().UTC().Add(-49 * time.Hour).Format(time.RFC3339),
		},
	}

	listing, err := GetLatestPrice(context.Background(), store, "Lightning Bolt")
	require.NoError(t, err)
	require.Nil(t, listing)
}

func TestGetLatestPrice_PrefersCheapestDoubleFacedAlias(t *testing.T) {
	originalVerify := verifyCardNameFunc
	defer func() { verifyCardNameFunc = originalVerify }()

	verifyCardNameFunc = func(_ context.Context, _ string) (string, error) {
		return "Jennifer Walters // The Sensational She-Hulk", nil
	}

	store := &mockStore{
		listings: map[string]*cardkingdom.Listing{
			"jennifer walters // the sensational she-hulk": {
				CardName:  "Jennifer Walters // The Sensational She-Hulk",
				PriceUsd:  69.99,
				UpdatedAt: time.Now().UTC().Format(time.RFC3339),
			},
			"jennifer walters": {
				CardName:  "Jennifer Walters",
				PriceUsd:  10.99,
				UpdatedAt: time.Now().UTC().Format(time.RFC3339),
			},
		},
	}

	listing, err := GetLatestPrice(context.Background(), store, "Jennifer Walters")
	require.NoError(t, err)
	require.Equal(t, []string{
		"jennifer walters",
		"the sensational she-hulk",
	}, store.getKeys)
	require.Equal(t, 10.99, listing.PriceUsd)
	require.Equal(t, "Jennifer Walters", listing.CardName)
}

func TestGetLatestPrice_DoubleFacedSkipsFullNameFoilVariant(t *testing.T) {
	originalVerify := verifyCardNameFunc
	defer func() { verifyCardNameFunc = originalVerify }()

	verifyCardNameFunc = func(_ context.Context, _ string) (string, error) {
		return "Jennifer Walters // The Sensational She-Hulk", nil
	}

	store := &mockStore{
		listings: map[string]*cardkingdom.Listing{
			"jennifer walters // the sensational she-hulk": {
				CardName:  "Jennifer Walters // The Sensational She-Hulk",
				PriceUsd:  69.99,
				IsFoil:    true,
				UpdatedAt: time.Now().UTC().Format(time.RFC3339),
			},
			"jennifer walters": {
				CardName:  "Jennifer Walters",
				PriceUsd:  10.99,
				UpdatedAt: time.Now().UTC().Format(time.RFC3339),
			},
			"the sensational she-hulk": {
				CardName:  "The Sensational She-Hulk",
				PriceUsd:  14.99,
				UpdatedAt: time.Now().UTC().Format(time.RFC3339),
			},
		},
	}

	listing, err := GetLatestPrice(context.Background(), store, "Jennifer Walters // The Sensational She-Hulk")
	require.NoError(t, err)
	require.Equal(t, []string{
		"jennifer walters",
		"the sensational she-hulk",
	}, store.getKeys)
	require.Equal(t, 10.99, listing.PriceUsd)
	require.Equal(t, "Jennifer Walters", listing.CardName)
}

func TestGetLatestPrice_DoubleFacedFallsBackToFullName(t *testing.T) {
	originalVerify := verifyCardNameFunc
	defer func() { verifyCardNameFunc = originalVerify }()

	verifyCardNameFunc = func(_ context.Context, _ string) (string, error) {
		return "Tony Stark // The Invincible Iron Man", nil
	}

	store := &mockStore{
		listings: map[string]*cardkingdom.Listing{
			"tony stark // the invincible iron man": {
				CardName:  "Tony Stark // The Invincible Iron Man",
				PriceUsd:  7.49,
				UpdatedAt: time.Now().UTC().Format(time.RFC3339),
			},
		},
	}

	listing, err := GetLatestPrice(context.Background(), store, "Tony Stark")
	require.NoError(t, err)
	require.Equal(t, []string{
		"tony stark",
		"the invincible iron man",
		"tony stark // the invincible iron man",
	}, store.getKeys)
	require.Equal(t, 7.49, listing.PriceUsd)
}

func TestListingIsFresh(t *testing.T) {
	now := time.Date(2026, 6, 26, 12, 0, 0, 0, time.UTC)

	require.True(t, listingIsFresh(&cardkingdom.Listing{
		UpdatedAt: "2026-06-26T00:00:00Z",
	}, now))
	require.False(t, listingIsFresh(&cardkingdom.Listing{
		UpdatedAt: "2026-06-24T11:59:59Z",
	}, now))
	require.False(t, listingIsFresh(&cardkingdom.Listing{}, now))
}

func TestListingIsFresh_PrefersSyncedAt(t *testing.T) {
	now := time.Date(2026, 6, 29, 12, 0, 0, 0, time.UTC)

	require.True(t, listingIsFresh(&cardkingdom.Listing{
		UpdatedAt: "2026-06-27T00:00:00Z",
		SyncedAt:  "2026-06-29T03:00:00Z",
	}, now))
	require.False(t, listingIsFresh(&cardkingdom.Listing{
		UpdatedAt: "2026-06-29T03:00:00Z",
		SyncedAt:  "2026-06-27T03:00:00Z",
	}, now))
}
