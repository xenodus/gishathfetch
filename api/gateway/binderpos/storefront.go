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

// shouldStartWithDecklist decides whether a store leads its search with the
// shared BinderPOS decklist portal or with its own storefront scrape. It
// alternates deterministically (round-robin) so that, across all selected
// stores in a single search, roughly half lead with the shared portal host and
// half lead with their own domains. Splitting the load this way halves the
// first-attempt burst on portal.binderpos.com, the host most prone to 429/503
// rate limiting because every BinderPOS store would otherwise funnel into it at
// once. The family not chosen as the lead still runs as a fallback.
var shouldStartWithDecklist = func() bool {
	return useDecklistForRoute(binderposDecklistRouteSeq.Add(1) - 1)
}

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
