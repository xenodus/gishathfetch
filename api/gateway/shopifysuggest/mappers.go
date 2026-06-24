package shopifysuggest

import (
	"fmt"
	"strings"

	"mtg-price-checker-sg/gateway"
)

// MapFyendalProduct maps a suggest product using Fyendal Hobby naming rules:
// strip a leading [Foil] prefix and infer set from tags when unambiguous.
func MapFyendalProduct(cfg Config, product Product) (gateway.Card, bool) {
	name, isFoil := ParseNameAndFoil(product.Title)

	var extraInfo []string
	if set := SetFromTags(product.Tags); set != "" {
		extraInfo = append(extraInfo, fmt.Sprintf("[%s]", set))
	}

	return cardFromProduct(cfg, product, name, isFoil, extraInfo)
}

// MapBinderposSetExtraProduct maps a suggest product using BinderPOS scrap
// variant-3 conventions: card name without trailing [Set], set name in ExtraInfo.
func MapBinderposSetExtraProduct(cfg Config, product Product) (gateway.Card, bool) {
	title := product.Title
	name := stripTrailingSet(title)
	isFoil := isFoilFromTags(product.Tags) || strings.Contains(strings.ToLower(title), "foil")

	var extraInfo []string
	if setName := extractSetName(title); setName != "" {
		extraInfo = append(extraInfo, setName)
	}

	return cardFromProduct(cfg, product, name, isFoil, extraInfo)
}

// MapBinderposFullTitleProduct maps a suggest product using BinderPOS scrap
// variant-2 conventions: keep the full listing title including set brackets.
func MapBinderposFullTitleProduct(cfg Config, product Product) (gateway.Card, bool) {
	title := product.Title
	isFoil := isFoilFromTags(product.Tags) || strings.Contains(strings.ToLower(title), "foil")
	return cardFromProduct(cfg, product, title, isFoil, nil)
}
