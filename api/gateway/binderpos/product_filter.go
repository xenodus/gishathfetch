package binderpos

import (
	"encoding/json"
	"strings"
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
	return isMagicProductType(productType, parseProductTags(productTagsJSON))
}

func isMagicProductType(productType string, tags []string) bool {
	pt := strings.ToLower(productType)
	if strings.Contains(pt, "mtg") || strings.Contains(pt, "magic the gathering") {
		return true
	}
	for _, tag := range tags {
		if strings.EqualFold(strings.TrimSpace(tag), "MTG") {
			return true
		}
	}
	return false
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
