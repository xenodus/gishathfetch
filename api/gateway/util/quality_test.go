package util

import "testing"

func TestIsKnownQuality(t *testing.T) {
	tests := []struct {
		quality string
		want    bool
	}{
		{"NM", true},
		{"nm", true},
		{"LP", true},
		{"Gift Bundle", false},
		{"Near Mint", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.quality, func(t *testing.T) {
			if got := IsKnownQuality(tt.quality); got != tt.want {
				t.Errorf("IsKnownQuality(%q) = %v, want %v", tt.quality, got, tt.want)
			}
		})
	}
}

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
