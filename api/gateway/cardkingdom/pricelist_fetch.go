package cardkingdom

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/pkg/config"
)

const (
	defaultCKPricelistURL   = "https://api.cardkingdom.com/api/v2/pricelist"
	ckPricelistErrorPrefix  = "ck price pricelist"
	ckPricelistFetchTimeout = 3 * time.Minute
	ckPricelistHTTPTimeout  = 2 * time.Minute
)

type ckPricelistPayload struct {
	Meta ckPricelistMeta   `json:"meta"`
	Data []ckPricelistItem `json:"data"`
}

type ckPricelistMeta struct {
	CreatedAt string `json:"created_at"`
	BaseURL   string `json:"base_url"`
}

type ckPricelistItem struct {
	URL              string              `json:"url"`
	Name             string              `json:"name"`
	Variation        string              `json:"variation"`
	Edition          string              `json:"edition"`
	IsFoil           jsonStringBool      `json:"is_foil"`
	PriceRetail      jsonStringFloat     `json:"price_retail"`
	QtyRetail        jsonStringInt       `json:"qty_retail"`
	ConditionValues  ckConditionValues   `json:"condition_values"`
}

type ckConditionValues struct {
	NmPrice jsonStringFloat `json:"nm_price"`
	NmQty   jsonStringInt   `json:"nm_qty"`
	ExPrice jsonStringFloat `json:"ex_price"`
	ExQty   jsonStringInt   `json:"ex_qty"`
	VgPrice jsonStringFloat `json:"vg_price"`
	VgQty   jsonStringInt   `json:"vg_qty"`
	GPrice  jsonStringFloat `json:"g_price"`
	GQty    jsonStringInt   `json:"g_qty"`
}

type jsonStringFloat float64

func (v *jsonStringFloat) UnmarshalJSON(data []byte) error {
	data = bytesTrimSpace(data)
	if len(data) == 0 || string(data) == "null" {
		*v = 0
		return nil
	}
	if data[0] == '"' {
		var raw string
		if err := json.Unmarshal(data, &raw); err != nil {
			return err
		}
		raw = strings.TrimSpace(raw)
		if raw == "" {
			*v = 0
			return nil
		}
		parsed, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return err
		}
		*v = jsonStringFloat(parsed)
		return nil
	}
	var parsed float64
	if err := json.Unmarshal(data, &parsed); err != nil {
		return err
	}
	*v = jsonStringFloat(parsed)
	return nil
}

type jsonStringInt int

func (v *jsonStringInt) UnmarshalJSON(data []byte) error {
	data = bytesTrimSpace(data)
	if len(data) == 0 || string(data) == "null" {
		*v = 0
		return nil
	}
	if data[0] == '"' {
		var raw string
		if err := json.Unmarshal(data, &raw); err != nil {
			return err
		}
		raw = strings.TrimSpace(raw)
		if raw == "" {
			*v = 0
			return nil
		}
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return err
		}
		*v = jsonStringInt(parsed)
		return nil
	}
	var parsed int
	if err := json.Unmarshal(data, &parsed); err != nil {
		return err
	}
	*v = jsonStringInt(parsed)
	return nil
}

type jsonStringBool bool

func (v *jsonStringBool) UnmarshalJSON(data []byte) error {
	data = bytesTrimSpace(data)
	if len(data) == 0 || string(data) == "null" {
		*v = false
		return nil
	}
	if data[0] == '"' {
		var raw string
		if err := json.Unmarshal(data, &raw); err != nil {
			return err
		}
		parsed, err := strconv.ParseBool(strings.TrimSpace(raw))
		if err != nil {
			return err
		}
		*v = jsonStringBool(parsed)
		return nil
	}
	var parsed bool
	if err := json.Unmarshal(data, &parsed); err != nil {
		return err
	}
	*v = jsonStringBool(parsed)
	return nil
}

func bytesTrimSpace(data []byte) []byte {
	return []byte(strings.TrimSpace(string(data)))
}

func fetchCheapestFromCKPricelist(ctx context.Context) (map[string]Listing, error) {
	ctx, cancel := context.WithTimeout(ctx, ckPricelistFetchTimeout)
	defer cancel()

	downloadURL := ckPricelistURL()
	log.Printf("ck price refresh: downloading pricelist from %s", downloadURL)

	downloadStarted := time.Now()
	payload, err := downloadCKPricelist(ctx, downloadURL)
	if err != nil {
		return nil, err
	}
	log.Printf(
		"ck price refresh: downloaded pricelist products=%d created_at=%s duration=%s",
		len(payload.Data),
		payload.Meta.CreatedAt,
		time.Since(downloadStarted).Round(time.Millisecond),
	)

	updatedAt, err := pricelistUpdatedAt(payload.Meta.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("%s: meta created_at: %w", ckPricelistErrorPrefix, err)
	}

	return cheapestListingsFromPricelist(payload, updatedAt), nil
}

func ckPricelistURL() string {
	if url := strings.TrimSpace(os.Getenv(config.CKPricelistURLEnv)); url != "" {
		return url
	}
	return defaultCKPricelistURL
}

func downloadCKPricelist(ctx context.Context, downloadURL string) (*ckPricelistPayload, error) {
	storeBase, err := url.Parse("https://www.cardkingdom.com/")
	if err != nil {
		return nil, fmt.Errorf("%s: store base url: %w", ckPricelistErrorPrefix, err)
	}

	resp, err := gateway.DoOutboundGET(ctx, downloadURL, gateway.OutboundRequestOptions{
		Style:          gateway.OutboundStyleJSON,
		StoreBase:      storeBase,
		Accept:         "application/json",
		SkipWebBotAuth: true,
	}, ckPricelistHTTPTimeout)
	if err != nil {
		return nil, fmt.Errorf("%s: download: %w", ckPricelistErrorPrefix, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("%s: status %d: %s", ckPricelistErrorPrefix, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s: read body: %w", ckPricelistErrorPrefix, err)
	}

	var payload ckPricelistPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("%s: decode json: %w", ckPricelistErrorPrefix, err)
	}
	if len(payload.Data) == 0 {
		return nil, fmt.Errorf("%s: empty product list", ckPricelistErrorPrefix)
	}

	return &payload, nil
}

func cheapestListingsFromPricelist(payload *ckPricelistPayload, updatedAt time.Time) map[string]Listing {
	cheapest := make(map[string]Listing)
	updatedAtValue := updatedAt.UTC().Format(time.RFC3339)
	baseURL := strings.TrimRight(strings.TrimSpace(payload.Meta.BaseURL), "/")

	for _, item := range payload.Data {
		listing, ok := listingFromPricelistItem(item, baseURL, updatedAtValue)
		if !ok {
			continue
		}
		considerCheapestListing(cheapest, listing)
	}

	return cheapest
}

func listingFromPricelistItem(item ckPricelistItem, baseURL string, updatedAt string) (Listing, bool) {
	cardName := strings.TrimSpace(item.Name)
	if cardName == "" {
		return Listing{}, false
	}
	if excludedCKPriceEdition(item.Edition) {
		return Listing{}, false
	}

	priceUsd, ok := cheapestInStockUSD(item.ConditionValues, float64(item.PriceRetail), int(item.QtyRetail))
	if !ok {
		return Listing{}, false
	}

	listingURL := pricelistItemURL(baseURL, item.URL)
	if listingURL == "" {
		return Listing{}, false
	}

	return Listing{
		CardName:  cardName,
		Edition:   pricelistEditionLabel(item.Edition, item.Variation),
		PriceUsd:  priceUsd,
		URL:       listingURL,
		IsFoil:    bool(item.IsFoil),
		UpdatedAt: updatedAt,
	}, true
}

func cheapestInStockUSD(values ckConditionValues, priceRetail float64, qtyRetail int) (float64, bool) {
	best := 0.0
	found := false

	for _, option := range []struct {
		price float64
		qty   int
	}{
		{float64(values.NmPrice), int(values.NmQty)},
		{float64(values.ExPrice), int(values.ExQty)},
		{float64(values.VgPrice), int(values.VgQty)},
		{float64(values.GPrice), int(values.GQty)},
	} {
		if option.qty <= 0 || option.price <= 0 {
			continue
		}
		if !found || option.price < best {
			best = option.price
			found = true
		}
	}

	if found {
		return best, true
	}
	if qtyRetail > 0 && priceRetail > 0 {
		return priceRetail, true
	}
	return 0, false
}

func pricelistEditionLabel(edition string, variation string) string {
	variation = strings.TrimSpace(variation)
	if variation != "" {
		return variation
	}
	return strings.TrimSpace(edition)
}

func pricelistItemURL(baseURL string, itemURL string) string {
	itemURL = strings.TrimSpace(itemURL)
	if itemURL == "" {
		return ""
	}
	if strings.HasPrefix(itemURL, "http://") || strings.HasPrefix(itemURL, "https://") {
		return itemURL
	}
	if baseURL == "" {
		baseURL = "https://www.cardkingdom.com"
	}
	return baseURL + "/" + strings.TrimLeft(itemURL, "/")
}

func pricelistUpdatedAt(createdAt string) (time.Time, error) {
	createdAt = strings.TrimSpace(createdAt)
	if createdAt == "" {
		return time.Time{}, fmt.Errorf("missing created_at")
	}

	layouts := []string{
		"2006-01-02 15:04:05",
		time.RFC3339,
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, createdAt); err == nil {
			return parsed.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported format %q", createdAt)
}
