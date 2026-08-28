package handler

import (
	"context"

	"mtg-price-checker-sg/pkg/telegrambot"
)

const telegramCKRunAction = telegrambot.CKRunAction

func runTelegramCKRun(ctx context.Context, chatID int64, query string) error {
	svc, err := getTelegramService(ctx)
	if err != nil {
		return err
	}
	return svc.RunCKSearch(ctx, chatID, query)
}
