package telegrambot

import "context"

const (
	// PriceRunAction is the Lambda async event action for card price lookups.
	PriceRunAction = "telegram-price-run"
)

// PriceRunEvent is the async Lambda payload for a /price command.
type PriceRunEvent struct {
	Action string `json:"action"`
	ChatID int64  `json:"chatId"`
	Query  string `json:"query"`
}

// PriceSearchAsync enqueues a background price lookup.
type PriceSearchAsync interface {
	EnqueuePriceSearch(ctx context.Context, chatID int64, query string) error
}
