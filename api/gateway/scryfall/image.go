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

// LookupImageURL returns a Scryfall card image URL for the given card name.
func LookupImageURL(ctx context.Context, cardName string) (string, error) {
	trimmed := strings.TrimSpace(cardName)
	if trimmed == "" {
		return "", nil
	}

	requestURL := fmt.Sprintf("%s?%s=%s", namedURL, "fuzzy", url.QueryEscape(trimmed))
	resp, err := httpGet(ctx, requestURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil
	}

	body, err := gateway.ReadResponseBody(resp)
	if err != nil {
		return "", err
	}

	var card struct {
		ImageURIs *struct {
			Normal string `json:"normal"`
		} `json:"image_uris"`
		CardFaces []struct {
			ImageURIs *struct {
				Normal string `json:"normal"`
			} `json:"image_uris"`
		} `json:"card_faces"`
	}
	if err := json.Unmarshal(body, &card); err != nil {
		return "", err
	}

	if card.ImageURIs != nil {
		if imageURL := strings.TrimSpace(card.ImageURIs.Normal); imageURL != "" {
			return imageURL, nil
		}
	}
	for _, face := range card.CardFaces {
		if face.ImageURIs == nil {
			continue
		}
		if imageURL := strings.TrimSpace(face.ImageURIs.Normal); imageURL != "" {
			return imageURL, nil
		}
	}

	return "", nil
}
