package binderpos

import (
	"testing"
)

func TestSelectStorefrontStrategiesScrapFirstThenDecklist(t *testing.T) {
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

	got := make([]string, 0, 6)
	for _, strategy := range append(scrap[:], decklist[:]...) {
		got = append(got, strategy.name)
	}

	want := []string{
		"scrap-dedicated",
		"scrap-direct",
		"scrap-dynamic",
		"decklist-dedicated",
		"decklist-direct",
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
