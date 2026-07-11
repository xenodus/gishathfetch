package ckpricereport

import (
	"time"

	"mtg-price-checker-sg/store/ckprices"
)

// Report is the daily CK price change export written to S3.
type Report struct {
	GeneratedAt  string                       `json:"generatedAt"`
	SyncedAt     string                       `json:"syncedAt,omitempty"`
	RankingLimit int                          `json:"rankingLimit"`
	Top          []ckprices.PriceChangeListing `json:"top"`
	Bottom       []ckprices.PriceChangeListing `json:"bottom"`
}

// NewReport builds an export payload from the latest DynamoDB price change rankings.
func NewReport(changes *ckprices.TopBottomPriceChanges, generatedAt time.Time) *Report {
	if changes == nil {
		changes = &ckprices.TopBottomPriceChanges{}
	}

	return &Report{
		GeneratedAt:  generatedAt.UTC().Format(time.RFC3339),
		SyncedAt:     syncedAtFromChanges(changes),
		RankingLimit: ckprices.PriceChangeRankingLimit,
		Top:          changes.Top,
		Bottom:       changes.Bottom,
	}
}

func syncedAtFromChanges(changes *ckprices.TopBottomPriceChanges) string {
	for _, listing := range append(changes.Top, changes.Bottom...) {
		if listing.SyncedAt != "" {
			return listing.SyncedAt
		}
	}
	return ""
}
