package cardkingdom

import "strings"

// excludedCKPriceEdition reports whether a set/edition should be omitted from CK
// cheapest-price indexing. World Championship Deck printings use the same card
// names as regular sets but are separate memorabilia products with distinct
// pricing.
func excludedCKPriceEdition(edition string) bool {
	edition = strings.TrimSpace(edition)
	if edition == "" {
		return false
	}

	lower := strings.ToLower(edition)
	if lower == "world championships" {
		return true
	}
	return strings.HasPrefix(lower, "world championship decks")
}
