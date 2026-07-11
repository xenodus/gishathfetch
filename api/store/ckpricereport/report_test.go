package ckpricereport

import (
	"testing"
	"time"

	"mtg-price-checker-sg/gateway/cardkingdom"
	"mtg-price-checker-sg/store/ckprices"
)

func TestNewReportIncludesSyncedAtAndRankings(t *testing.T) {
	increase := 15
	decrease := -10
	changes := &ckprices.TopBottomPriceChanges{
		Top: []ckprices.PriceChangeListing{
			{
				NameKey: "lightning bolt",
				Listing: cardkingdom.Listing{
					CardName:           "Lightning Bolt",
					PriceUsd:           1.25,
					PriceChangePercent: &increase,
					SyncedAt:           "2026-07-11T00:00:00Z",
				},
			},
		},
		Bottom: []ckprices.PriceChangeListing{
			{
				NameKey: "counterspell",
				Listing: cardkingdom.Listing{
					CardName:           "Counterspell",
					PriceUsd:           0.75,
					PriceChangePercent: &decrease,
					SyncedAt:           "2026-07-11T00:00:00Z",
				},
			},
		},
	}

	report := NewReport(changes, time.Date(2026, 7, 11, 12, 0, 0, 0, time.UTC))
	if report.GeneratedAt != "2026-07-11T12:00:00Z" {
		t.Fatalf("unexpected generatedAt: %q", report.GeneratedAt)
	}
	if report.SyncedAt != "2026-07-11T00:00:00Z" {
		t.Fatalf("unexpected syncedAt: %q", report.SyncedAt)
	}
	if report.RankingLimit != ckprices.PriceChangeRankingLimit {
		t.Fatalf("unexpected ranking limit: %d", report.RankingLimit)
	}
	if len(report.Top) != 1 || len(report.Bottom) != 1 {
		t.Fatalf("unexpected rankings: top=%d bottom=%d", len(report.Top), len(report.Bottom))
	}
}
