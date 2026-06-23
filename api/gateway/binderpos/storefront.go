package binderpos

import (
	"time"
)

const (
	storefrontSuggestPath   = "/search/suggest.json"
	binderposDecklistAPIURL = "https://portal.binderpos.com/external/shopify/decklist"
	binderposDecklistType   = "mtg"
	binderposAttemptTimeout = 10 * time.Second

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

// shouldUseDecklistEndpoint decides whether one first-attempt storefront lookup
// is routed through the shared BinderPOS decklist portal or the per-store
// product-details path. It alternates deterministically (round-robin) so that,
// across all selected stores in a single search, half of the first attempts hit
// the shared portal host and half hit their own Shopify domains. Splitting the
// load this way halves the concurrent burst on portal.binderpos.com, the host
// most prone to 429/503 rate limiting because every BinderPOS store would
// otherwise funnel into it at once.
var shouldUseDecklistEndpoint = func() bool {
	return useDecklistForRoute(binderposDecklistRouteSeq.Add(1) - 1)
}

type storefrontSuggestResponse struct {
	Resources struct {
		Results struct {
			Products []storefrontProduct `json:"products"`
		} `json:"results"`
	} `json:"resources"`
}

type storefrontProduct struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Image string `json:"image"`
}

type storefrontProductDetail struct {
	Title    string                   `json:"title"`
	Variants []storefrontProductStock `json:"variants"`
}

type storefrontProductStock struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Available bool   `json:"available"`
	Price     int    `json:"price"`
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
