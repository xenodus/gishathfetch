package cardkingdom

import "strings"

const doubleFacedNameSeparator = " // "

// NormalizeNameKey lowercases and trims a card name for DynamoDB lookup.
func NormalizeNameKey(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// NameLookupKeys returns normalized lookup keys for a card name, including each
// face of a double-faced card split on " // ".
func NameLookupKeys(cardName string) []string {
	trimmed := strings.TrimSpace(cardName)
	if trimmed == "" {
		return nil
	}

	keys := []string{NormalizeNameKey(trimmed)}
	if before, after, ok := strings.Cut(trimmed, doubleFacedNameSeparator); ok {
		if front := NormalizeNameKey(before); front != "" {
			keys = append(keys, front)
		}
		if back := NormalizeNameKey(after); back != "" {
			keys = append(keys, back)
		}
	}

	return uniqueNameKeys(keys)
}

// DoubleFacedFaceNames returns normalized front and back face keys for a
// double-faced card name split on " // ".
func DoubleFacedFaceNames(cardName string) (front string, back string, ok bool) {
	trimmed := strings.TrimSpace(cardName)
	before, after, found := strings.Cut(trimmed, doubleFacedNameSeparator)
	if !found {
		return "", "", false
	}
	front = NormalizeNameKey(before)
	back = NormalizeNameKey(after)
	return front, back, front != "" && back != ""
}

// PriceLookupKeys returns the lookup keys used when resolving CK search prices.
// For double-faced cards the combined name, front face, and back face are always
// checked together in a single pass; the cheapest fresh listing wins. The
// combined name is not a fallback.
func PriceLookupKeys(cardName string) []string {
	trimmed := strings.TrimSpace(cardName)
	if trimmed == "" {
		return nil
	}

	combined := NormalizeNameKey(trimmed)
	front, back, ok := DoubleFacedFaceNames(trimmed)
	if !ok {
		return []string{combined}
	}

	return uniqueNameKeys([]string{combined, front, back})
}

// ListingNameKeys returns the lookup keys that should receive a listing when it is
// indexed. Foil double-faced listings are stored only under the full combined
// name so a variant foil price does not overwrite cheaper face-only names.
func ListingNameKeys(listing Listing) []string {
	keys := NameLookupKeys(listing.CardName)
	if !listing.IsFoil || len(keys) <= 1 {
		return keys
	}
	return keys[:1]
}

func uniqueNameKeys(keys []string) []string {
	seen := make(map[string]struct{}, len(keys))
	unique := make([]string, 0, len(keys))
	for _, key := range keys {
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		unique = append(unique, key)
	}
	return unique
}
