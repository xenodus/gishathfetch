package handler

import (
	"context"
	"testing"
	"time"

	"mtg-price-checker-sg/gateway/cardkingdom"
	"mtg-price-checker-sg/store/ckpricereport"
	"mtg-price-checker-sg/store/ckprices"
)

func TestSelectPriceChangesPrefersPostRefresh(t *testing.T) {
	increase := 1.0
	pre := &ckprices.TopBottomPriceChanges{
		Top: []ckprices.PriceChangeListing{{NameKey: "pre", Listing: cardkingdom.Listing{PriceChangeUsd: &increase}}},
	}
	post := &ckprices.TopBottomPriceChanges{
		Top: []ckprices.PriceChangeListing{{NameKey: "post", Listing: cardkingdom.Listing{PriceChangeUsd: &increase}}},
	}

	selected := selectPriceChanges(pre, post)
	if len(selected.Top) != 1 || selected.Top[0].NameKey != "post" {
		t.Fatalf("selected = %+v, want post refresh rankings", selected)
	}
}

func TestSelectPriceChangesFallsBackToPreRefresh(t *testing.T) {
	increase := 1.0
	pre := &ckprices.TopBottomPriceChanges{
		Top: []ckprices.PriceChangeListing{{NameKey: "pre", Listing: cardkingdom.Listing{PriceChangeUsd: &increase}}},
	}
	post := &ckprices.TopBottomPriceChanges{}

	selected := selectPriceChanges(pre, post)
	if len(selected.Top) != 1 || selected.Top[0].NameKey != "pre" {
		t.Fatalf("selected = %+v, want pre refresh rankings", selected)
	}
}

func TestWritePriceChangeReportPreservesExistingWhenEmpty(t *testing.T) {
	increase := 1.0
	existing := &ckpricereport.Report{
		GeneratedAt: "2026-07-16T17:35:00Z",
		Top: []ckprices.PriceChangeListing{{
			NameKey: "bolt",
			Listing: cardkingdom.Listing{PriceChangeUsd: &increase},
		}},
	}
	writer := &mockCKPriceReportWriter{existing: existing}

	preserved, gotExisting, err := writePriceChangeReport(context.Background(), writer, &ckpricereport.Report{})
	if err != nil {
		t.Fatalf("writePriceChangeReport: %v", err)
	}
	if !preserved {
		t.Fatal("expected existing export to be preserved")
	}
	if gotExisting != existing {
		t.Fatalf("existing report = %p, want %p", gotExisting, existing)
	}
	if writer.written {
		t.Fatal("expected empty export not to overwrite existing report")
	}
}

func TestWritePriceChangeReportWritesWhenExistingEmpty(t *testing.T) {
	writer := &mockCKPriceReportWriter{}

	preserved, _, err := writePriceChangeReport(context.Background(), writer, &ckpricereport.Report{})
	if err != nil {
		t.Fatalf("writePriceChangeReport: %v", err)
	}
	if preserved {
		t.Fatal("expected empty export to be written when no existing report")
	}
	if !writer.written {
		t.Fatal("expected write when no existing report")
	}
}

func TestRunCKPriceRefresh_UsesPreRefreshRankingsOnSameDayReRun(t *testing.T) {
	originalStoreFunc := newCKRefreshStoreFunc
	originalRefreshFunc := refreshCKPricesFunc
	originalWriterFunc := newCKPriceReportWriterFunc
	originalNowFunc := ckPriceReportNowFunc
	originalAlertFunc := sendJobDiscordAlert
	defer func() {
		newCKRefreshStoreFunc = originalStoreFunc
		refreshCKPricesFunc = originalRefreshFunc
		newCKPriceReportWriterFunc = originalWriterFunc
		ckPriceReportNowFunc = originalNowFunc
		sendJobDiscordAlert = originalAlertFunc
	}()

	sendJobDiscordAlert = func(string) {}

	increase := 1.0
	decrease := -0.5
	pre := &ckprices.TopBottomPriceChanges{
		Top:    []ckprices.PriceChangeListing{{NameKey: "bolt", Listing: cardkingdom.Listing{PriceChangeUsd: &increase}}},
		Bottom: []ckprices.PriceChangeListing{{NameKey: "counterspell", Listing: cardkingdom.Listing{PriceChangeUsd: &decrease}}},
	}

	writer := &mockCKPriceReportWriter{}
	newCKRefreshStoreFunc = func(_ context.Context) (ckprices.Store, error) {
		return &sequentialCKRefreshStore{pre: pre, post: &ckprices.TopBottomPriceChanges{}}, nil
	}
	refreshCKPricesFunc = func(_ context.Context, _ ckprices.Store) (int, error) {
		return 42, nil
	}
	newCKPriceReportWriterFunc = func(_ context.Context) (ckpricereport.Writer, error) {
		return writer, nil
	}
	ckPriceReportNowFunc = func() time.Time {
		return time.Date(2026, 7, 16, 17, 57, 47, 0, time.UTC)
	}

	if err := runCKPriceRefresh(context.Background()); err != nil {
		t.Fatalf("runCKPriceRefresh: %v", err)
	}
	if !writer.written {
		t.Fatal("expected ck price change report to be written")
	}
	if writer.report == nil || len(writer.report.Top) != 1 || writer.report.Top[0].NameKey != "bolt" {
		t.Fatalf("report = %+v, want pre-refresh rankings", writer.report)
	}
	if len(writer.report.Bottom) != 1 || writer.report.Bottom[0].NameKey != "counterspell" {
		t.Fatalf("report bottom = %+v, want counterspell", writer.report.Bottom)
	}
}

type sequentialCKRefreshStore struct {
	pre   *ckprices.TopBottomPriceChanges
	post  *ckprices.TopBottomPriceChanges
	calls int
}

func (m *sequentialCKRefreshStore) GetByNameKey(_ context.Context, _ string) (*cardkingdom.Listing, error) {
	return nil, nil
}

func (m *sequentialCKRefreshStore) GetPriceChangesByUsd(_ context.Context, _ bool, _ int) ([]ckprices.PriceChangeListing, error) {
	return nil, nil
}

func (m *sequentialCKRefreshStore) GetTopBottomPriceChanges(_ context.Context) (*ckprices.TopBottomPriceChanges, error) {
	m.calls++
	if m.calls == 1 {
		return m.pre, nil
	}
	return m.post, nil
}

func (m *sequentialCKRefreshStore) PutAll(_ context.Context, _ map[string]cardkingdom.Listing) (string, error) {
	return "", nil
}

type mockCKPriceReportWriter struct {
	existing *ckpricereport.Report
	written  bool
	report   *ckpricereport.Report
}

func (m *mockCKPriceReportWriter) Write(_ context.Context, report *ckpricereport.Report) error {
	m.written = true
	m.report = report
	return nil
}

func (m *mockCKPriceReportWriter) ReadLatest(_ context.Context) (*ckpricereport.Report, error) {
	if m.existing == nil {
		return nil, ckpricereport.ErrLatestReportNotFound
	}
	return m.existing, nil
}
