package mtgasia

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"unicode"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/util"
	"mtg-price-checker-sg/pkg/config"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

const (
	base44ServerURL = "https://base44.app"
	base44AppID     = "6a984c3786516ed6b7f9a090"
	inventoryLimit  = 25
)

type inventoryItem struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	SetID         string  `json:"set_id"`
	SubCategory   string  `json:"sub_category"`
	Condition     string  `json:"condition"`
	SellingPrice  float64 `json:"selling_price"`
	QtyAvailable  float64 `json:"qty_available"`
	Image         string  `json:"image"`
}

type cardSet struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Hidden bool   `json:"hidden"`
}

func searchBase44Inventory(ctx context.Context, searchStr string) ([]gateway.Card, error) {
	searchStr = strings.TrimSpace(searchStr)
	if searchStr == "" {
		return nil, fmt.Errorf("%s: empty search string", StoreName)
	}

	items, err := fetchInventoryItems(ctx, searchStr)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}

	sets, err := fetchSets(ctx, items)
	if err != nil {
		return nil, err
	}

	cards := make([]gateway.Card, 0, len(items))
	for _, item := range items {
		set, hasSet := sets[item.SetID]
		if !storefrontSellable(set, hasSet) {
			continue
		}
		setName := ""
		if hasSet {
			setName = set.Name
		}
		if card, ok := mapInventoryItem(item, setName); ok {
			cards = append(cards, card)
		}
	}
	return cards, nil
}

// storefrontSellable mirrors the MTG Asia Base44 SPA: CardSet.hidden shows "NOT FOR SALE".
func storefrontSellable(set cardSet, found bool) bool {
	return !found || !set.Hidden
}

func fetchInventoryItems(ctx context.Context, searchStr string) ([]inventoryItem, error) {
	filter := map[string]any{
		"name": map[string]any{
			"$regex":   regexp.QuoteMeta(searchStr),
			"$options": "i",
		},
		"qty_available": map[string]any{"$gt": 0},
	}
	filterJSON, err := json.Marshal(filter)
	if err != nil {
		return nil, err
	}

	apiURL := &url.URL{
		Scheme: "https",
		Host:   strings.TrimPrefix(base44ServerURL, "https://"),
		Path:   "/api/apps/" + base44AppID + "/entities/InventoryItem",
		RawQuery: url.Values{
			"q":     {string(filterJSON)},
			"limit": {fmt.Sprint(inventoryLimit)},
			"sort":  {"name"},
		}.Encode(),
	}

	resp, err := gateway.DoOutboundGET(ctx, apiURL.String(), gateway.OutboundRequestOptions{
		Style: gateway.OutboundStyleJSON,
	}, config.DedicatedSearchAttemptTimeout)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("%s", gateway.FormatUnexpectedHTTPStatus(StoreName, resp, body))
	}

	var items []inventoryItem
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, gateway.WrapJSONDecodeError(err, resp, body)
	}
	return items, nil
}

func fetchSets(ctx context.Context, items []inventoryItem) (map[string]cardSet, error) {
	unique := map[string]struct{}{}
	for _, item := range items {
		if id := strings.TrimSpace(item.SetID); id != "" {
			unique[id] = struct{}{}
		}
	}
	if len(unique) == 0 {
		return map[string]cardSet{}, nil
	}

	ids := make([]string, 0, len(unique))
	for id := range unique {
		ids = append(ids, id)
	}

	filterJSON, err := json.Marshal(map[string]any{
		"id": map[string]any{"$in": ids},
	})
	if err != nil {
		return nil, err
	}

	apiURL := &url.URL{
		Scheme: "https",
		Host:   strings.TrimPrefix(base44ServerURL, "https://"),
		Path:   "/api/apps/" + base44AppID + "/entities/CardSet",
		RawQuery: url.Values{
			"q":     {string(filterJSON)},
			"limit": {fmt.Sprint(len(ids))},
		}.Encode(),
	}

	resp, err := gateway.DoOutboundGET(ctx, apiURL.String(), gateway.OutboundRequestOptions{
		Style: gateway.OutboundStyleJSON,
	}, config.DedicatedSearchAttemptTimeout)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("%s", gateway.FormatUnexpectedHTTPStatus(StoreName, resp, body))
	}

	var sets []cardSet
	if err := json.Unmarshal(body, &sets); err != nil {
		return nil, gateway.WrapJSONDecodeError(err, resp, body)
	}

	byID := make(map[string]cardSet, len(sets))
	for _, set := range sets {
		if id := strings.TrimSpace(set.ID); id != "" {
			set.Name = strings.TrimSpace(set.Name)
			byID[id] = set
		}
	}
	return byID, nil
}

func mapInventoryItem(item inventoryItem, setName string) (gateway.Card, bool) {
	name := strings.TrimSpace(item.Name)
	itemID := strings.TrimSpace(item.ID)
	if name == "" || itemID == "" {
		return gateway.Card{}, false
	}

	price := item.SellingPrice
	if price <= 0 || item.QtyAvailable <= 0 {
		return gateway.Card{}, false
	}

	productTitle := name
	if setName != "" {
		productTitle = fmt.Sprintf("%s [%s]", name, setName)
	}
	qualityLabel := util.MapQuality(strings.TrimSpace(item.Condition))
	displayName := productTitle
	if qualityLabel != "" {
		displayName = strings.TrimSpace(productTitle + " - " + qualityLabel)
	}

	cardURL, err := buildCardSinglePageURL(StoreBaseURL, itemID, name, setName)
	if err != nil {
		return gateway.Card{}, false
	}

	subCategory := strings.ToLower(strings.TrimSpace(item.SubCategory))
	isFoil := strings.Contains(subCategory, "foil") && !strings.Contains(subCategory, "non-foil")

	img := strings.TrimSpace(item.Image)
	if img == "" {
		img = fmt.Sprintf("https://placehold.co/304x424?text=%s", url.QueryEscape(name))
	}

	return gateway.Card{
		Name:    displayName,
		Url:     cardURL,
		Img:     img,
		Price:   price,
		InStock: true,
		IsFoil:  isFoil,
		Source:  StoreName,
		Quality: qualityLabel,
	}, true
}

func buildCardSinglePageURL(baseURL, itemID, cardName, setName string) (string, error) {
	slug := cardPathSlug(cardName, setName)
	u, err := url.Parse(strings.TrimRight(baseURL, "/") + "/card/" + slug + "/" + itemID)
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("utm_source", config.UtmSource)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// cardPathSlug mirrors the Base44 storefront slug helper (set name + card name).
func cardPathSlug(cardName, setName string) string {
	parts := make([]string, 0, 2)
	if strings.TrimSpace(setName) != "" {
		parts = append(parts, strings.TrimSpace(setName))
	}
	if strings.TrimSpace(cardName) != "" {
		parts = append(parts, strings.TrimSpace(cardName))
	}
	raw := strings.Join(parts, " ")
	raw = strings.ToLower(raw)

	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	slug, _, _ := transform.String(t, raw)

	var b strings.Builder
	lastDash := false
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash && b.Len() > 0 {
			b.WriteByte('-')
			lastDash = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if len(out) > 80 {
		out = out[:80]
		out = strings.TrimRight(out, "-")
	}
	if out == "" {
		return "card"
	}
	return out
}
