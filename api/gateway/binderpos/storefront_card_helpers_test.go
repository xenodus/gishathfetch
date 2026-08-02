package binderpos

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQualityFromProductTitle(t *testing.T) {
	tests := []struct {
		title string
		want  string
	}{
		{"Reanimate — NM [Breaking News]", "NM"},
		{"Intuition — LP [Tempest]", "LP"},
		{"Subtlety — NM Foil [Secret Lair Drop]", "NM"},
		{"Opt [Ixalan]", ""},
		{"The Hobbit — Gift Bundle", ""},
	}
	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			require.Equal(t, tt.want, qualityFromProductTitle(tt.title))
		})
	}
}

func TestStripEmbeddedQualityFromProductTitle(t *testing.T) {
	tests := []struct {
		title string
		want  string
	}{
		{"Reanimate — NM [Breaking News]", "Reanimate [Breaking News]"},
		{"Subtlety — NM Foil [Secret Lair Drop]", "Subtlety [Secret Lair Drop]"},
		{"Opt [Ixalan]", "Opt [Ixalan]"},
		{"The Hobbit — Gift Bundle", "The Hobbit — Gift Bundle"},
	}
	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			require.Equal(t, tt.want, stripEmbeddedQualityFromProductTitle(tt.title))
		})
	}
}

func TestResolveCardQualityPrefersVariantTitle(t *testing.T) {
	require.Equal(t, "Near Mint", resolveCardQuality("Reanimate — NM [Breaking News]", "Near Mint"))
	require.Equal(t, "NM", resolveCardQuality("Reanimate — NM [Breaking News]", "Default Title"))
}
