package handler

import (
	"mtg-price-checker-sg/controller"
	"mtg-price-checker-sg/pkg/telegramphoto"
)

func selectTelegramPhotoURL(cards []controller.Card) string {
	candidates := make([]string, 0, len(cards))
	for _, card := range cards {
		candidates = append(candidates, card.Img)
	}
	return telegramphoto.Select(candidates...)
}
