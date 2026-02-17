package util

import "testing"

func TestMapQuality(t *testing.T) {
	tests := []struct {
		name    string
		quality string
		want    string
	}{
		{"NM", "NM", "Near Mint"},
		{"nm", "nm", "Near Mint"},
		{"LP", "LP", "Lightly Played"},
		{"lp", "lp", "Lightly Played"},
		{"MP", "MP", "Moderately Played"},
		{"mp", "mp", "Moderately Played"},
		{"HP", "HP", "Heavily Played"},
		{"hp", "hp", "Heavily Played"},
		{"DM", "DM", "Damaged"},
		{"dm", "dm", "Damaged"},
		{"Unknown", "Unknown", "Unknown"},
		{"Mixed Case", "nM", "Near Mint"},
		{"With Spaces", " NM ", "Near Mint"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MapQuality(tt.quality); got != tt.want {
				t.Errorf("MapQuality() = %v, want %v", got, tt.want)
			}
		})
	}
}
