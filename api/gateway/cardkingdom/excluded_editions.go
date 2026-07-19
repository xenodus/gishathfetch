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
	if strings.Contains(lower, "world championship") {
		return true
	}
	// CK "Variants" sets are alternate-art or special-finish SKUs. They share
	// card names with the main set but should not represent the default cheapest
	// price (for example borderless surge foils).
	return strings.HasSuffix(lower, " variants")
}
