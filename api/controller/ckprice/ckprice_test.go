package ckprice

import (
	"context"
	"testing"

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

func (m *mockStore) PutAll(_ context.Context, _ map[string]cardkingdom.Listing) error {
	return nil
}

func TestGetLatestPrice_UsesVerifiedName(t *testing.T) {
	originalVerify := verifyCardNameFunc
	defer func() { verifyCardNameFunc = originalVerify }()

	verifyCardNameFunc = func(_ context.Context, _ string) (string, error) {
		return "Lightning Bolt", nil
	}

	store := &mockStore{
		listing: &cardkingdom.Listing{CardName: "Lightning Bolt", PriceUsd: 1.49},
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
