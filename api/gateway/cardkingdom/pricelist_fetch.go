package cardkingdom

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/util"
	"mtg-price-checker-sg/pkg/config"
)

const (
	defaultCKPricelistURL   = "https://api.cardkingdom.com/api/v2/pricelist"
	ckPricelistErrorPrefix  = "ck price pricelist"
	// The CK pricelist JSON is ~65MB (~150k products). When the residential proxy
	// fallback is used, the body can take several minutes after headers return;
	// http.Client.Timeout covers the full round trip including body read.
	ckPricelistFetchTimeout = 14 * time.Minute
	ckPricelistHTTPTimeout  = 13 * time.Minute
	// Emit body-read progress while large pricelist downloads stream through proxy.
	ckPricelistBodyReadLogInterval = 15 * time.Second
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

func ckPricelistOutboundOptions() gateway.OutboundRequestOptions {
	return gateway.OutboundRequestOptions{
		Accept:         "application/json",
		SkipWebBotAuth: true,
	}
}

func ckPricelistResidentialProxyURL() (string, bool) {
	if proxyURL, ok := util.GetResidentialProxyURL(); ok {
		return proxyURL, true
	}
	return util.GetCKPricelistProxyURL()
}

var downloadCKPricelistOnceFunc = downloadCKPricelistOnce

func downloadCKPricelist(ctx context.Context, downloadURL string) (*ckPricelistPayload, error) {
	payload, err := downloadCKPricelistOnceFunc(ctx, downloadURL, ckPricelistOutboundOptions())
	if err == nil {
		return payload, nil
	}

	proxyURL, ok := ckPricelistResidentialProxyURL()
	if !ok {
		return nil, err
	}

	log.Printf("ck price refresh: direct pricelist download failed, retrying via residential proxy")
	proxyOpts := ckPricelistOutboundOptions()
	proxyOpts.OnlyProxyURL = proxyURL
	return downloadCKPricelistOnceFunc(ctx, downloadURL, proxyOpts)
}

func downloadCKPricelistOnce(ctx context.Context, downloadURL string, opts gateway.OutboundRequestOptions) (*ckPricelistPayload, error) {
	resp, err := gateway.DoOutboundGET(ctx, downloadURL, opts, ckPricelistHTTPTimeout)
	if err != nil {
		return nil, fmt.Errorf("%s: download: %w", ckPricelistErrorPrefix, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("%s: status %d: %s", ckPricelistErrorPrefix, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	log.Printf("ck price refresh: reading pricelist body")
	readStarted := time.Now()
	bodyReader := &progressLogReader{
		reader:     resp.Body,
		interval:   ckPricelistBodyReadLogInterval,
		label:      "ck price refresh: reading pricelist body",
		totalBytes: resp.ContentLength,
	}
	raw, err := io.ReadAll(bodyReader)
	if err != nil {
		return nil, fmt.Errorf("%s: read body: %w", ckPricelistErrorPrefix, err)
	}
	log.Printf(
		"ck price refresh: read pricelist body bytes=%d duration=%s",
		len(raw),
		time.Since(readStarted).Round(time.Millisecond),
	)

	var payload ckPricelistPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("%s: decode json: %w", ckPricelistErrorPrefix, err)
	}
	if len(payload.Data) == 0 {
		return nil, fmt.Errorf("%s: empty product list", ckPricelistErrorPrefix)
	}

	return &payload, nil
}

type progressLogReader struct {
	reader     io.Reader
	interval   time.Duration
	label      string
	totalBytes int64

	readStarted time.Time
	readBytes   int64
	lastLogged  time.Time
}

func (r *progressLogReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	if n <= 0 {
		return n, err
	}

	now := time.Now()
	if r.readStarted.IsZero() {
		r.readStarted = now
		r.lastLogged = now
	}
	r.readBytes += int64(n)

	if now.Sub(r.lastLogged) >= r.interval {
		r.logProgress(now)
		r.lastLogged = now
	}

	return n, err
}

func (r *progressLogReader) logProgress(now time.Time) {
	elapsed := now.Sub(r.readStarted).Round(time.Millisecond)
	if r.totalBytes > 0 {
		pct := (float64(r.readBytes) * 100) / float64(r.totalBytes)
		log.Printf(
			"%s progress bytes=%d total=%d pct=%.1f%% elapsed=%s",
			r.label,
			r.readBytes,
			r.totalBytes,
			pct,
			elapsed,
		)
		return
	}

	log.Printf("%s progress bytes=%d elapsed=%s", r.label, r.readBytes, elapsed)
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

	priceUsd, ok := cheapestListedUSD(item.ConditionValues, float64(item.PriceRetail))
	if !ok {
		return Listing{}, false
	}

	listingURL := pricelistItemURL(baseURL, item.URL)
	if listingURL == "" {
		return Listing{}, false
	}

	inStock := listingIsInStock(item)
	return Listing{
		CardName:  cardName,
		Edition:   pricelistEditionLabel(item.Edition, item.Variation),
		PriceUsd:  priceUsd,
		URL:       listingURL,
		IsFoil:    bool(item.IsFoil),
		InStock:   &inStock,
		UpdatedAt: updatedAt,
	}, true
}

func listingIsInStock(item ckPricelistItem) bool {
	return int(item.QtyRetail) > 0
}

func cheapestListedUSD(values ckConditionValues, priceRetail float64) (float64, bool) {
	if values.NmPrice > 0 {
		return float64(values.NmPrice), true
	}
	if priceRetail > 0 {
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
