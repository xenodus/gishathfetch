package handler

import (
	"context"
	"encoding/json"
	"testing"

	"mtg-price-checker-sg/gateway/cardkingdom"
	"mtg-price-checker-sg/store/ckprices"
)

type mockCKRefreshStore struct{}

func (m *mockCKRefreshStore) GetByNameKey(_ context.Context, _ string) (*cardkingdom.Listing, error) {
	return nil, nil
}

func (m *mockCKRefreshStore) PutAll(_ context.Context, _ map[string]cardkingdom.Listing) (string, error) {
	return "", nil
}

func TestHandle_RoutesCKPriceRefreshRun(t *testing.T) {
	originalStoreFunc := newCKRefreshStoreFunc
	originalRefreshFunc := refreshCKPricesFunc
	defer func() {
		newCKRefreshStoreFunc = originalStoreFunc
		refreshCKPricesFunc = originalRefreshFunc
	}()

	newCKRefreshStoreFunc = func(_ context.Context) (ckprices.Store, error) {
		return &mockCKRefreshStore{}, nil
	}
	refreshCKPricesFunc = func(_ context.Context, _ ckprices.Store) (int, error) {
		return 1, nil
	}

	event, err := json.Marshal(map[string]string{"action": ckPriceRefreshRunAction})
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}

	if _, err = Handle(context.Background(), event); err != nil {
		t.Fatalf("handle event: %v", err)
	}
}
