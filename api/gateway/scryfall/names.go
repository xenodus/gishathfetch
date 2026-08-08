package scryfall

import (
	"strings"

	"mtg-price-checker-sg/gateway/util"
)

// FoldCardNameForMatch lowercases a card name and strips combining marks so
// "Kíli the Resourceful" and "Kili the Resourceful" compare equal.
func FoldCardNameForMatch(name string) string {
	return util.FoldForMatch(name)
}

func cardNamesMatchForVerify(canonical, query string) bool {
	if strings.EqualFold(canonical, query) {
		return true
	}
	return FoldCardNameForMatch(canonical) == FoldCardNameForMatch(query)
}
