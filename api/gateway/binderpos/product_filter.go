package binderpos

import (
	"encoding/json"
	"strings"

	"mtg-price-checker-sg/gateway/shopifysuggest"
)

// shouldIncludeBinderposProduct reports whether a scraped BinderPOS product
// should be surfaced. Multi-game storefronts (e.g. Hideyoshi) tag each product
// with data-product-type; when present we only keep Magic listings so Pokemon
// and other TCG inventory is excluded. Stores that omit the attribute are left
// unchanged for backward compatibility.
func shouldIncludeBinderposProduct(productType, productTagsJSON string) bool {
	productType = strings.TrimSpace(productType)
	if productType == "" {
		return true
	}
	return shopifysuggest.IsMagicProduct(productType, "", parseProductTags(productTagsJSON))
}

func parseProductTags(productTagsJSON string) []string {
	productTagsJSON = strings.TrimSpace(productTagsJSON)
	if productTagsJSON == "" {
		return nil
	}

	var tags []string
	if err := json.Unmarshal([]byte(productTagsJSON), &tags); err != nil {
		return nil
	}
	return tags
}
