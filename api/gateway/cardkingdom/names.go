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
