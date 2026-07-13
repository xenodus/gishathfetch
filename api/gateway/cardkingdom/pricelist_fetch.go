package cardkingdom

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"mtg-price-checker-sg/gateway"
)

const (
	pricelistURL        = "https://api.cardkingdom.com/api/v2/pricelist"
	pricelistTimeout    = 3 * time.Minute
	ckPricelistAttempts = 3
)

var fetchPricelistBody = func(ctx context.Context) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, pricelistTimeout)
	defer cancel()
	return fetchPricelistBodyFromURL(ctx, pricelistURL)
}

func fetchPricelistBodyFromURL(ctx context.Context, requestURL string) ([]byte, error) {
	storeBase, err := url.Parse(listingBaseURL)
	if err != nil {
		return nil, err
	}

	resp, err := gateway.DoOutboundGET(ctx, requestURL, gateway.OutboundRequestOptions{
		Style:          gateway.OutboundStyleJSON,
		StoreBase:      storeBase,
		SkipWebBotAuth: true,
	}, pricelistTimeout)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	if looksLikeCloudflareChallenge(body) {
		return nil, fmt.Errorf("cloudflare challenge")
	}
	return body, nil
}

func fetchCheapestFromCKPricelist(ctx context.Context) (map[string]Listing, error) {
	var lastErr error
	for attempt := 1; attempt <= ckPricelistAttempts; attempt++ {
		listings, err := fetchCheapestFromCKPricelistOnce(ctx)
		if err == nil {
			return listings, nil
		}
		lastErr = err
		if attempt < ckPricelistAttempts {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
	}
	return nil, lastErr
}

func fetchCheapestFromCKPricelistOnce(ctx context.Context) (map[string]Listing, error) {
	body, err := fetchPricelistBody(ctx)
	if err != nil {
		return nil, err
	}

	var products []Product
	if err := json.Unmarshal(body, &products); err == nil {
		return BuildCheapestByName(products, time.Now().UTC()), nil
	}

	var payload struct {
		Data []Product `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	return BuildCheapestByName(payload.Data, time.Now().UTC()), nil
}

func mergeCheapestListings(primary map[string]Listing, supplemental map[string]Listing) {
	for _, listing := range supplemental {
		considerCheapestListing(primary, listing)
	}
}

func looksLikeCloudflareChallenge(body []byte) bool {
	prefix := string(body)
	if len(prefix) > 512 {
		prefix = prefix[:512]
	}
	lower := strings.ToLower(prefix)
	return strings.Contains(lower, "just a moment") || strings.Contains(lower, "cloudflare")
}
