package scryfall

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"mtg-price-checker-sg/gateway"
)

const (
	autocompleteURL = "https://api.scryfall.com/cards/autocomplete"
	namedURL        = "https://api.scryfall.com/cards/named"
	// httpClientTimeout bounds each Scryfall round trip even when the caller
	// context has a long deadline (e.g. Lambda remaining time).
	httpClientTimeout = 3 * time.Second
)

var scryfallHTTPClient = &http.Client{Timeout: httpClientTimeout}

var httpGet = func(ctx context.Context, requestURL string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}
	if err := gateway.PrepareOutboundRequest(ctx, req, gateway.OutboundRequestOptions{
		Style: gateway.OutboundStyleJSON,
	}); err != nil {
		return nil, err
	}
	return scryfallHTTPClient.Do(req)
}

// VerifyCardName returns the canonical Scryfall card name when query matches a card.
func VerifyCardName(ctx context.Context, query string) (string, error) {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return "", nil
	}

	autocompleteRequestURL := fmt.Sprintf("%s?q=%s", autocompleteURL, url.QueryEscape(trimmed))
	resp, err := httpGet(ctx, autocompleteRequestURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		body, err := gateway.ReadResponseBody(resp)
		if err != nil {
			return "", err
		}
		var autocomplete struct {
			Data []string `json:"data"`
		}
		if err := json.Unmarshal(body, &autocomplete); err != nil {
			return "", err
		}
		for _, name := range autocomplete.Data {
			if cardNamesMatchForVerify(name, trimmed) {
				return name, nil
			}
		}
	}

	if verifiedName, ok, err := lookupNamedCard(ctx, trimmed, "exact"); err != nil {
		return "", err
	} else if ok {
		return verifiedName, nil
	}

	if verifiedName, ok, err := lookupNamedCard(ctx, trimmed, "fuzzy"); err != nil {
		return "", err
	} else if ok {
		return verifiedName, nil
	}

	return "", nil
}

func lookupNamedCard(ctx context.Context, query, matchMode string) (string, bool, error) {
	namedRequestURL := fmt.Sprintf("%s?%s=%s", namedURL, matchMode, url.QueryEscape(query))
	namedResp, err := httpGet(ctx, namedRequestURL)
	if err != nil {
		return "", false, err
	}
	defer namedResp.Body.Close()

	if namedResp.StatusCode != http.StatusOK {
		return "", false, nil
	}

	body, err := gateway.ReadResponseBody(namedResp)
	if err != nil {
		return "", false, err
	}
	var card struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(body, &card); err != nil {
		return "", false, err
	}
	if card.Name == "" || !cardNamesMatchForVerify(card.Name, query) {
		return "", false, nil
	}
	return card.Name, true, nil
}
