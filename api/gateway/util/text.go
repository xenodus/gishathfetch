package util

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// StripDiacritics removes combining marks from a string while preserving case.
// Store inventory search APIs typically only match ASCII names, so "Kíli" becomes "Kili".
func StripDiacritics(value string) string {
	value = norm.NFKD.String(strings.TrimSpace(value))
	var b strings.Builder
	b.Grow(len(value))
	for _, r := range value {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// FoldForMatch lowercases a string and strips combining marks so accented and
// ASCII-only variants compare equal during search filtering.
func FoldForMatch(value string) string {
	value = norm.NFKD.String(strings.TrimSpace(value))
	var b strings.Builder
	b.Grow(len(value))
	for _, r := range value {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}

// StoreSearchQueries returns the query variants to send to a store. When the
// search string contains diacritics, both the original and ASCII-stripped forms
// are searched so shops that index either spelling can return results.
func StoreSearchQueries(searchString string) []string {
	trimmed := strings.TrimSpace(searchString)
	if trimmed == "" {
		return nil
	}

	stripped := StripDiacritics(trimmed)
	if stripped == trimmed {
		return []string{trimmed}
	}
	if stripped == "" {
		return []string{trimmed}
	}
	return []string{trimmed, stripped}
}
