package handler

import (
	"context"
	"strings"

	"mtg-price-checker-sg/controller"
	"mtg-price-checker-sg/gateway/scryfall"
	"mtg-price-checker-sg/pkg/telegramphoto"
)

var lookupScryfallImageURLFunc = scryfall.LookupImageURL

func selectTelegramPhotoURL(ctx context.Context, cards []controller.Card) string {
	if len(cards) == 0 {
		return ""
	}

	cardName := strings.TrimSpace(cards[0].Name)
	if cardName == "" {
		return ""
	}

	imageURL, err := lookupScryfallImageURLFunc(ctx, cardName)
	if err != nil || imageURL == "" {
		return ""
	}
	imageURL = telegramphoto.Normalize(imageURL)
	if !telegramphoto.IsSendable(imageURL) {
		return ""
	}
	return imageURL
}
