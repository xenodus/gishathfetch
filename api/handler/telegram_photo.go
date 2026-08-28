package handler

import (
	"context"

	"mtg-price-checker-sg/controller"
	"mtg-price-checker-sg/pkg/telegramphoto"
)

func selectTelegramPhotoURL(_ context.Context, cards []controller.Card) string {
	if len(cards) == 0 {
		return ""
	}
	return telegramphoto.Select(cards[0].Img)
}
