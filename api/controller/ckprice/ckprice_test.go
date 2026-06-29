package ckprice

import (
	"context"
	"testing"
	"time"

	"mtg-price-checker-sg/gateway/cardkingdom"

	"github.com/stretchr/testify/require"
)

type mockStore struct {
	listing *cardkingdom.Listing
	getKey  string
}

func (m *mockStore) GetByNameKey(_ context.Context, nameKey string) (*cardkingdom.Listing, error) {
	m.getKey = nameKey
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
	require.Equal(t, "lightning bolt", store.getKey)
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
			UpdatedAt: time.Now().UTC().Add(-25 * time.Hour).Format(time.RFC3339),
		},
	}

	listing, err := GetLatestPrice(context.Background(), store, "Lightning Bolt")
	require.NoError(t, err)
	require.Nil(t, listing)
}

func TestListingIsFresh(t *testing.T) {
	now := time.Date(2026, 6, 26, 12, 0, 0, 0, time.UTC)

	require.True(t, listingIsFresh(&cardkingdom.Listing{
		UpdatedAt: "2026-06-26T00:00:00Z",
	}, now))
	require.False(t, listingIsFresh(&cardkingdom.Listing{
		UpdatedAt: "2026-06-25T11:59:59Z",
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
