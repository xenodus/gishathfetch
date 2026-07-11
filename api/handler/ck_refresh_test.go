package handler

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"mtg-price-checker-sg/gateway/cardkingdom"
	"mtg-price-checker-sg/store/ckpricereport"
	"mtg-price-checker-sg/store/ckprices"
)

type mockCKRefreshStore struct {
	changes *ckprices.TopBottomPriceChanges
}

func (m *mockCKRefreshStore) GetByNameKey(_ context.Context, _ string) (*cardkingdom.Listing, error) {
	return nil, nil
}

func (m *mockCKRefreshStore) GetPriceChangesByPercent(_ context.Context, _ bool, _ int) ([]ckprices.PriceChangeListing, error) {
	return nil, nil
}

func (m *mockCKRefreshStore) GetTopBottomPriceChanges(_ context.Context) (*ckprices.TopBottomPriceChanges, error) {
	if m.changes != nil {
		return m.changes, nil
	}
	return &ckprices.TopBottomPriceChanges{}, nil
}

func (m *mockCKRefreshStore) PutAll(_ context.Context, _ map[string]cardkingdom.Listing) (string, error) {
	return "", nil
}

type mockCKPriceReportWriter struct {
	written bool
}

func (m *mockCKPriceReportWriter) Write(_ context.Context, _ *ckpricereport.Report) error {
	m.written = true
	return nil
}

func TestHandle_RoutesCKPriceRefreshRun(t *testing.T) {
	originalStoreFunc := newCKRefreshStoreFunc
	originalRefreshFunc := refreshCKPricesFunc
	originalWriterFunc := newCKPriceReportWriterFunc
	originalNowFunc := ckPriceReportNowFunc
	defer func() {
		newCKRefreshStoreFunc = originalStoreFunc
		refreshCKPricesFunc = originalRefreshFunc
		newCKPriceReportWriterFunc = originalWriterFunc
		ckPriceReportNowFunc = originalNowFunc
	}()

	writer := &mockCKPriceReportWriter{}
	newCKRefreshStoreFunc = func(_ context.Context) (ckprices.Store, error) {
		return &mockCKRefreshStore{}, nil
	}
	refreshCKPricesFunc = func(_ context.Context, _ ckprices.Store) (int, error) {
		return 1, nil
	}
	newCKPriceReportWriterFunc = func(_ context.Context) (ckpricereport.Writer, error) {
		return writer, nil
	}
	ckPriceReportNowFunc = func() time.Time {
		return time.Date(2026, 7, 11, 12, 0, 0, 0, time.UTC)
	}

	event, err := json.Marshal(map[string]string{"action": ckPriceRefreshRunAction})
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}

	if _, err = Handle(context.Background(), event); err != nil {
		t.Fatalf("handle event: %v", err)
	}
	if !writer.written {
		t.Fatalf("expected ck price change report to be written")
	}
}
