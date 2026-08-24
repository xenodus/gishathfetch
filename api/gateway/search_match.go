package gateway

import "strings"

// CardsMatchSearch reports whether at least one card name contains the trimmed
// search string (case-insensitive). This mirrors controller result filtering so
// gateways can reject irrelevant upstream matches before returning them.
func CardsMatchSearch(cards []Card, searchStr string) bool {
	lowerSearch := strings.ToLower(strings.TrimSpace(searchStr))
	if lowerSearch == "" {
		return len(cards) > 0
	}

	for _, card := range cards {
		if strings.Contains(strings.ToLower(strings.TrimSpace(card.Name)), lowerSearch) {
			return true
		}
	}
	return false
}
