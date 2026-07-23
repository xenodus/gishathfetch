package gateway

import "testing"

func TestCardsMatchSearch(t *testing.T) {
	cards := []Card{
		{Name: "Lightning Bolt [Unlimited Edition]"},
		{Name: "Shock"},
	}

	if !CardsMatchSearch(cards, "lightning bolt") {
		t.Fatal("expected lightning bolt match")
	}
	if CardsMatchSearch(cards, "Teferi") {
		t.Fatal("expected no teferi match")
	}
	if !CardsMatchSearch(cards, "Lightning Bolt") {
		t.Fatal("expected case-insensitive match")
	}
	if CardsMatchSearch(nil, "opt") {
		t.Fatal("expected empty cards to not match")
	}
	if !CardsMatchSearch([]Card{{Name: "Opt"}}, "") {
		t.Fatal("expected any card to match empty search")
	}
}
