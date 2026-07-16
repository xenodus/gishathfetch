package ckpricereport

import (
	"testing"

	"mtg-price-checker-sg/gateway/cardkingdom"
	"mtg-price-checker-sg/store/ckprices"
)

func TestHasMovers(t *testing.T) {
	increase := 0.5
	decrease := -0.25

	tests := []struct {
		name   string
		report *Report
		want   bool
	}{
		{
			name:   "nil report",
			report: nil,
			want:   false,
		},
		{
			name: "empty rankings",
			report: &Report{
				Top:    nil,
				Bottom: nil,
			},
			want: false,
		},
		{
			name: "top riser",
			report: &Report{
				Top: []ckprices.PriceChangeListing{{
					NameKey: "bolt",
					Listing: cardkingdom.Listing{PriceChangeUsd: &increase},
				}},
			},
			want: true,
		},
		{
			name: "bottom drop",
			report: &Report{
				Bottom: []ckprices.PriceChangeListing{{
					NameKey: "counterspell",
					Listing: cardkingdom.Listing{PriceChangeUsd: &decrease},
				}},
			},
			want: true,
		},
		{
			name: "zero change ignored",
			report: &Report{
				Top: []ckprices.PriceChangeListing{{
					NameKey: "bolt",
					Listing: cardkingdom.Listing{PriceChangeUsd: new(0.0)},
				}},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasMovers(tt.report); got != tt.want {
				t.Fatalf("HasMovers() = %v, want %v", got, tt.want)
			}
		})
	}
}

//go:fix inline
func ptr(v float64) *float64 {
	return new(v)
}
