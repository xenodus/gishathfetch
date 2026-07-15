package binderpos

import (
	"testing"
)

func TestStorefrontStrategyOrder(t *testing.T) {
	scrap := [3]storefrontStrategy{
		{name: "scrap-dedicated"},
		{name: "scrap-direct"},
		{name: "scrap-dynamic"},
	}
	decklist := [3]storefrontStrategy{
		{name: "decklist-dedicated"},
		{name: "decklist-direct"},
		{name: "decklist-dynamic"},
	}

	got := []string{
		scrap[0].name,
		scrap[1].name,
		decklist[0].name,
		decklist[1].name,
		scrap[2].name,
		decklist[2].name,
	}

	want := []string{
		"scrap-dedicated",
		"scrap-direct",
		"decklist-dedicated",
		"decklist-direct",
		"scrap-dynamic",
		"decklist-dynamic",
	}
	if len(got) != len(want) {
		t.Fatalf("expected %d strategies, got %d", len(want), len(got))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("strategy %d: expected %q, got %q", i, want[i], got[i])
		}
	}
}
