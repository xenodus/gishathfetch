package telegrambot

import "context"

const (
	// PriceRunAction is the Lambda async event action for card price lookups.
	PriceRunAction = "telegram-price-run"
	// CKRunAction is the Lambda async event action for Card Kingdom lookups.
	CKRunAction = "telegram-ck-run"
)

// PriceRunEvent is the async Lambda payload for a /price or /ck command.
type PriceRunEvent struct {
	Action string `json:"action"`
	ChatID int64  `json:"chatId"`
	Query  string `json:"query"`
}

// CommandAsync enqueues background bot command lookups.
type CommandAsync interface {
	EnqueuePriceSearch(ctx context.Context, chatID int64, query string) error
	EnqueueCKSearch(ctx context.Context, chatID int64, query string) error
}

// PriceSearchAsync is kept for backward compatibility.
type PriceSearchAsync = CommandAsync
