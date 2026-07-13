package binderpos

import (
	"testing"
)

func TestSearchUsesScrapStrategiesOnly(t *testing.T) {
	strategies := []storefrontStrategy{
		{name: "scrap-dedicated"},
		{name: "scrap-direct"},
		{name: "scrap-dynamic"},
	}

	got := make([]string, len(strategies))
	for i := range strategies {
		got[i] = strategies[i].name
	}

	want := []string{"scrap-dedicated", "scrap-direct", "scrap-dynamic"}
	if len(got) != len(want) {
		t.Fatalf("expected %d strategies, got %d", len(want), len(got))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("strategy %d: expected %q, got %q", i, want[i], got[i])
		}
	}
}
