package handler

import (
	"context"

	"mtg-price-checker-sg/controller"
	"mtg-price-checker-sg/gateway/scryfall"
	"mtg-price-checker-sg/pkg/telegramphoto"
)

func selectTelegramPhotoURL(ctx context.Context, cards []controller.Card, searchQuery string) string {
	candidates := make([]string, 0, len(cards))
	for _, card := range cards {
		candidates = append(candidates, card.Img)
	}
	if photoURL := telegramphoto.Select(candidates...); photoURL != "" {
		return photoURL
	}

	fallbackName := searchQuery
	if len(cards) > 0 {
		if name := cards[0].Name; name != "" {
			fallbackName = name
		}
	}
	imageURL, err := scryfall.LookupImageURL(ctx, fallbackName)
	if err != nil || imageURL == "" {
		return ""
	}
	return telegramphoto.Normalize(imageURL)
}
