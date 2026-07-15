package binderpos

import (
	"time"

	"mtg-price-checker-sg/pkg/config"
)

const (
	binderposDecklistAPIURL = "https://portal.binderpos.com/external/shopify/decklist"
	binderposDecklistType   = "mtg"
	binderposAttemptTimeout = config.SearchAttemptTimeout

	// binderposDecklistMaxAttempts bounds how many times a single decklist call
	// is sent (initial try plus retries) when the shared portal host responds
	// with a transient rate-limit/5xx status or a network error. Retries stay
	// within the per-attempt timeout budget and back off between sends so a
	// struggling upstream is not hammered.
	binderposDecklistMaxAttempts = 3
	// binderposDecklistRetryBaseDelay is the first backoff step; subsequent
	// retries grow it exponentially with equal jitter.
	binderposDecklistRetryBaseDelay = 300 * time.Millisecond
	// binderposDecklistRetryMaxDelay caps a single backoff/Retry-After wait so a
	// large or hostile Retry-After value cannot stall the attempt.
	binderposDecklistRetryMaxDelay = 2500 * time.Millisecond
)

type storefrontDecklistRequestItem struct {
	Card     string `json:"card"`
	Quantity int    `json:"quantity"`
}

type storefrontDecklistLine struct {
	// The API can send a boolean validName; binding it would break decode, so it is not mapped.
	Products []storefrontDecklistProduct `json:"products"`
}

type storefrontDecklistProduct struct {
	Title    string                    `json:"title"`
	Name     string                    `json:"name"`
	Handle   string                    `json:"handle"`
	SetName  string                    `json:"setName"`
	Image    string                    `json:"img"`
	Variants []storefrontDecklistStock `json:"variants"`
}

type storefrontDecklistStock struct {
	ShopifyID int64   `json:"shopifyId"`
	Title     string  `json:"title"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
}
