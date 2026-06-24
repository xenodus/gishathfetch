package shopifysuggest

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/util"
	"mtg-price-checker-sg/pkg/config"
)

// BuildProductURL builds a clean product URL from the product handle, attaching
// the UTM source. It falls back to the suggest-provided relative URL when a
// handle is unavailable.
func BuildProductURL(baseURL, handle, suggestURL string) (string, error) {
	path := strings.TrimSpace(suggestURL)
	if handle != "" {
		path = "/products/" + handle
	}

	cleanPageURL, err := url.Parse(strings.TrimSpace(baseURL + path))
	if err != nil {
		return "", err
	}
	cleanPageURL.RawQuery = url.Values{
		"utm_source": []string{config.UtmSource},
	}.Encode()

	return cleanPageURL.String(), nil
}

// ResolveImage returns the best available image URL from a suggest product. When
// the store provides no image (some listings, e.g. "The List" reprints, have
// none), it falls back to a titled placeholder so a card never carries an empty
// image, mirroring the BinderPOS scrape/decklist behavior.
func ResolveImage(product Product) string {
	if img := strings.TrimSpace(product.Image); img != "" {
		return img
	}
	if img := strings.TrimSpace(product.FeaturedImage.URL); img != "" {
		return img
	}
	return fmt.Sprintf("https://placehold.co/304x424?text=%s", url.QueryEscape(strings.TrimSpace(product.Title)))
}

// IsMagicProduct reports whether a storefront product belongs to Magic: The
// Gathering. It acts as a defensive guard so non-MTG products are never
// surfaced even if upstream search behaviour changes.
func IsMagicProduct(productType, vendor string, tags []string) bool {
	pt := strings.ToLower(productType)
	if strings.Contains(pt, "mtg") || strings.Contains(pt, "magic the gathering") {
		return true
	}
	if strings.Contains(strings.ToLower(vendor), "magic the gathering") {
		return true
	}
	for _, tag := range tags {
		if strings.EqualFold(strings.TrimSpace(tag), "MTG") {
			return true
		}
	}
	return false
}

func parsePositivePrice(storeName, rawPrice string) (float64, error) {
	price, err := util.ParsePrice(rawPrice)
	if err != nil {
		log.Printf("error parsing price for %s with value [%s]: %v", storeName, rawPrice, err)
		return 0, err
	}
	return price, nil
}

// foilTitlePrefix is the prefix some Shopify stores use on foil single listings,
// e.g. "[Foil] Cauldron of Essence".
const foilTitlePrefix = "[foil]"

// ParseNameAndFoil strips the "[Foil]" listing prefix from a product title and
// reports whether the listing is foil.
func ParseNameAndFoil(title string) (string, bool) {
	trimmed := strings.TrimSpace(title)
	if strings.HasPrefix(strings.ToLower(trimmed), foilTitlePrefix) {
		return strings.TrimSpace(trimmed[len(foilTitlePrefix):]), true
	}
	return trimmed, false
}

// nonSetTags are tags that describe rarity, status, or category rather than the
// card's set. They are excluded when inferring a card's set from its tags.
var nonSetTags = map[string]struct{}{
	"mtg":         {},
	"new arrival": {},
	"promo":       {},
	"common":      {},
	"uncommon":    {},
	"uncommond":   {},
	"rare":        {},
	"mythic":      {},
	"mythic rare": {},
	"special":     {},
	"land":        {},
	"basic land":  {},
	"foil":        {},
	"preorder":    {},
	"pre-order":   {},
	"restock":     {},
}

// SetFromTags infers the card's set name from its tags. To preserve data
// integrity it only returns a set when exactly one candidate tag remains after
// removing rarity/status tags; otherwise it returns an empty string rather than
// risk emitting an incorrect set.
func SetFromTags(tags []string) string {
	var candidates []string
	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed == "" {
			continue
		}
		if _, skip := nonSetTags[strings.ToLower(trimmed)]; skip {
			continue
		}
		candidates = append(candidates, trimmed)
	}

	if len(candidates) == 1 {
		return candidates[0]
	}
	return ""
}

func isFoilFromTags(tags []string) bool {
	for _, tag := range tags {
		if strings.EqualFold(strings.TrimSpace(tag), "foil") {
			return true
		}
	}
	return false
}

func stripTrailingSet(productTitle string) string {
	title := strings.TrimSpace(productTitle)
	open := strings.LastIndex(title, "[")
	close := strings.LastIndex(title, "]")
	if open >= 0 && close > open && close == len(title)-1 {
		return strings.TrimSpace(title[:open])
	}
	return title
}

func extractSetName(productTitle string) string {
	title := strings.TrimSpace(productTitle)
	open := strings.LastIndex(title, "[")
	close := strings.LastIndex(title, "]")
	if open >= 0 && close > open && close == len(title)-1 {
		return strings.TrimSpace(title[open+1 : close])
	}
	return ""
}

func cardFromProduct(cfg Config, product Product, name string, isFoil bool, extraInfo []string) (gateway.Card, bool) {
	price, err := parsePositivePrice(cfg.StoreName, product.Price)
	if err != nil || price <= 0 {
		return gateway.Card{}, false
	}

	cleanPageURL, err := BuildProductURL(cfg.BaseURL, product.Handle, product.URL)
	if err != nil {
		log.Printf("error parsing url for %s with handle [%s]: %v", cfg.StoreName, product.Handle, err)
		return gateway.Card{}, false
	}

	return gateway.Card{
		Name:      name,
		Url:       cleanPageURL,
		InStock:   true,
		IsFoil:    isFoil,
		Price:     price,
		Source:    cfg.StoreName,
		Img:       ResolveImage(product),
		ExtraInfo: extraInfo,
	}, true
}
