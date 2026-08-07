package scryfall

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// FoldCardNameForMatch lowercases a card name and strips combining marks so
// "Kíli the Resourceful" and "Kili the Resourceful" compare equal.
func FoldCardNameForMatch(name string) string {
	name = norm.NFKD.String(strings.TrimSpace(name))
	var b strings.Builder
	b.Grow(len(name))
	for _, r := range name {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}

func cardNamesMatchForVerify(canonical, query string) bool {
	if strings.EqualFold(canonical, query) {
		return true
	}
	return FoldCardNameForMatch(canonical) == FoldCardNameForMatch(query)
}
