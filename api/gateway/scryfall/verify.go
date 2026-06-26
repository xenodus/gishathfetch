package scryfall

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"mtg-price-checker-sg/gateway"
)

const (
	autocompleteURL = "https://api.scryfall.com/cards/autocomplete"
	namedURL        = "https://api.scryfall.com/cards/named"
)

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
	return http.DefaultClient.Do(req)
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
			if strings.EqualFold(name, trimmed) {
				return name, nil
			}
		}
	}

	namedRequestURL := fmt.Sprintf("%s?exact=%s", namedURL, url.QueryEscape(trimmed))
	namedResp, err := httpGet(ctx, namedRequestURL)
	if err != nil {
		return "", err
	}
	defer namedResp.Body.Close()

	if namedResp.StatusCode != http.StatusOK {
		return "", nil
	}

	body, err := gateway.ReadResponseBody(namedResp)
	if err != nil {
		return "", err
	}
	var card struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(body, &card); err != nil {
		return "", err
	}
	return card.Name, nil
}
