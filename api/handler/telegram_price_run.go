package handler

import (
	"context"

	"mtg-price-checker-sg/pkg/telegrambot"
)

const telegramPriceRunAction = telegrambot.PriceRunAction

func runTelegramPriceRun(ctx context.Context, chatID int64, query string) error {
	svc, err := getTelegramService(ctx)
	if err != nil {
		return err
	}
	return svc.RunPriceSearch(ctx, chatID, query)
}
