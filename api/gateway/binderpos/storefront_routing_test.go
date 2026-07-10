package binderpos

import (
	"testing"
)

func TestUseDecklistForRoute(t *testing.T) {
	tests := []struct {
		name string
		seq  uint32
		want bool
	}{
		{name: "even seq 0 leads with decklist", seq: 0, want: true},
		{name: "odd seq 1 leads with scrap", seq: 1, want: false},
		{name: "even seq 2 leads with decklist", seq: 2, want: true},
		{name: "odd seq 3 leads with scrap", seq: 3, want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := useDecklistForRoute(test.seq)
			if got != test.want {
				t.Fatalf("expected useDecklistForRoute(%d)=%t, got %t", test.seq, test.want, got)
			}
		})
	}
}

// TestShouldStartWithDecklist_SplitsConsecutiveCallsEvenly verifies that the
// round-robin selector lets half of the stores lead with the decklist portal
// and half lead with their own storefront scrape, which is what reduces the
// concurrent first-attempt load on the shared portal host.
func TestShouldStartWithDecklist_SplitsConsecutiveCallsEvenly(t *testing.T) {
	previousSeq := binderposDecklistRouteSeq.Load()
	binderposDecklistRouteSeq.Store(0)
	t.Cleanup(func() { binderposDecklistRouteSeq.Store(previousSeq) })

	const calls = 10
	decklistCount := 0
	for range calls {
		if shouldStartWithDecklist() {
			decklistCount++
		}
	}

	if decklistCount != calls/2 {
		t.Fatalf("expected %d of %d stores leading with decklist, got %d", calls/2, calls, decklistCount)
	}
}

func TestOrderDecklistAndScrap(t *testing.T) {
	decklist := [3]storefrontStrategy{
		{name: "decklist-dedicated"},
		{name: "decklist-direct"},
		{name: "decklist-dynamic"},
	}
	scrap := [3]storefrontStrategy{
		{name: "scrap-dedicated"},
		{name: "scrap-direct"},
		{name: "scrap-dynamic"},
	}

	t.Run("decklist lead runs both dedicated/direct before any dynamic", func(t *testing.T) {
		previousSelector := shouldStartWithDecklist
		shouldStartWithDecklist = func() bool { return true }
		t.Cleanup(func() { shouldStartWithDecklist = previousSelector })

		got := strategyNames(orderDecklistAndScrap(decklist, scrap))
		want := []string{
			"decklist-dedicated", "decklist-direct",
			"scrap-dedicated", "scrap-direct",
			"decklist-dynamic", "scrap-dynamic",
		}
		assertStrategyOrder(t, want, got)
	})

	t.Run("scrap lead runs both dedicated/direct before any dynamic", func(t *testing.T) {
		previousSelector := shouldStartWithDecklist
		shouldStartWithDecklist = func() bool { return false }
		t.Cleanup(func() { shouldStartWithDecklist = previousSelector })

		got := strategyNames(orderDecklistAndScrap(decklist, scrap))
		want := []string{
			"scrap-dedicated", "scrap-direct",
			"decklist-dedicated", "decklist-direct",
			"scrap-dynamic", "decklist-dynamic",
		}
		assertStrategyOrder(t, want, got)
	})
}

func strategyNames(strategies []storefrontStrategy) []string {
	names := make([]string, len(strategies))
	for idx, strategy := range strategies {
		names[idx] = strategy.name
	}
	return names
}

func assertStrategyOrder(t *testing.T, want, got []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("expected %d strategies, got %d (%v)", len(want), len(got), got)
	}
	for idx := range want {
		if got[idx] != want[idx] {
			t.Fatalf("strategy %d: expected %q, got %q (full order %v)", idx+1, want[idx], got[idx], got)
		}
	}
}
