package fyendalhobby

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/util"
	"mtg-price-checker-sg/pkg/config"
)

const StoreName = "Fyendal Hobby"
const StoreBaseURL = "https://fyendalhobby.com"
const StoreSearchPath = "/search/suggest.json"

// predictiveSearchLimit is the maximum number of products Shopify's predictive
// search endpoint will return per request (the platform caps this at 10).
const predictiveSearchLimit = "10"

// mtgSingleProductType scopes the predictive search to Magic: The Gathering
// single cards. Fyendal Hobby stocks several TCGs (Flesh and Blood, Grand
// Archive, etc.), so the search query is filtered with
// product_type:"MTG Single Cards" to keep results MTG-only and prevent the
// 10-result predictive limit from being consumed by non-MTG products.
const mtgSingleProductType = "MTG Single Cards"

// foilTitlePrefix is the prefix Fyendal Hobby uses on foil single listings, e.g.
// "[Foil] Cauldron of Essence". The non-foil counterpart has no prefix.
const foilTitlePrefix = "[foil]"

type Store struct {
	Name       string
	BaseUrl    string
	SearchPath string
}

func NewLGS() gateway.LGS {
	return Store{
		Name:       StoreName,
		BaseUrl:    StoreBaseURL,
		SearchPath: StoreSearchPath,
	}
}

type suggestResponse struct {
	Resources struct {
		Results struct {
			Products []suggestProduct `json:"products"`
		} `json:"results"`
	} `json:"resources"`
}

type suggestProduct struct {
	Title         string   `json:"title"`
	Handle        string   `json:"handle"`
	URL           string   `json:"url"`
	Price         string   `json:"price"`
	Available     bool     `json:"available"`
	Image         string   `json:"image"`
	ProductType   string   `json:"type"`
	Vendor        string   `json:"vendor"`
	Tags          []string `json:"tags"`
	FeaturedImage struct {
		URL string `json:"url"`
	} `json:"featured_image"`
}

func (s Store) Search(ctx context.Context, searchStr string) ([]gateway.Card, error) {
	var cards []gateway.Card

	searchQuery := fmt.Sprintf("product_type:%q AND %s", mtgSingleProductType, searchStr)

	apiURL := &url.URL{
		Scheme: "https",
		Host:   "fyendalhobby.com",
		Path:   s.SearchPath,
		RawQuery: url.Values{
			"q":                {searchQuery},
			"resources[type]":  {"product"},
			"resources[limit]": {predictiveSearchLimit},
		}.Encode(),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL.String(), nil)
	if err != nil {
		return cards, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", gateway.RandomBrowserUserAgent())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return cards, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return cards, err
	}

	var res suggestResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return cards, err
	}

	for _, product := range res.Resources.Results.Products {
		if !product.Available {
			continue
		}
		if !isMagicProduct(product.ProductType, product.Vendor, product.Tags) {
			continue
		}

		price, err := util.ParsePrice(product.Price)
		if err != nil {
			log.Printf("error parsing price for %s with value [%s]: %v", s.Name, product.Price, err)
			continue
		}
		if price <= 0 {
			continue
		}

		name, isFoil := parseNameAndFoil(product.Title)

		cleanPageURL, err := s.buildProductURL(product.Handle, product.URL)
		if err != nil {
			log.Printf("error parsing url for %s with handle [%s]: %v", s.Name, product.Handle, err)
			continue
		}

		var extraInfo []string
		if set := setFromTags(product.Tags); set != "" {
			extraInfo = append(extraInfo, fmt.Sprintf("[%s]", set))
		}

		cards = append(cards, gateway.Card{
			Name:      name,
			Url:       cleanPageURL,
			InStock:   true,
			IsFoil:    isFoil,
			Price:     price,
			Source:    s.Name,
			Img:       resolveImage(product),
			ExtraInfo: extraInfo,
		})
	}

	return cards, nil
}

// buildProductURL builds a clean product URL from the product handle, attaching
// the UTM source. It falls back to the suggest-provided relative URL when a
// handle is unavailable.
func (s Store) buildProductURL(handle, suggestURL string) (string, error) {
	path := strings.TrimSpace(suggestURL)
	if handle != "" {
		path = "/products/" + handle
	}

	cleanPageURL, err := url.Parse(strings.TrimSpace(s.BaseUrl + path))
	if err != nil {
		return "", err
	}
	cleanPageURL.RawQuery = url.Values{
		"utm_source": []string{config.UtmSource},
	}.Encode()

	return cleanPageURL.String(), nil
}

func resolveImage(product suggestProduct) string {
	if img := strings.TrimSpace(product.Image); img != "" {
		return img
	}
	return strings.TrimSpace(product.FeaturedImage.URL)
}

// isMagicProduct reports whether a storefront product belongs to Magic: The
// Gathering. The search query already scopes results to MTG single cards
// server-side; this acts as a defensive secondary guard so non-MTG products are
// never surfaced even if the upstream filter behaviour changes.
func isMagicProduct(productType, vendor string, tags []string) bool {
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

// parseNameAndFoil strips the "[Foil]" listing prefix from a product title and
// reports whether the listing is foil.
func parseNameAndFoil(title string) (string, bool) {
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

// setFromTags infers the card's set name from its tags. To preserve data
// integrity it only returns a set when exactly one candidate tag remains after
// removing rarity/status tags; otherwise it returns an empty string rather than
// risk emitting an incorrect set.
func setFromTags(tags []string) string {
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
