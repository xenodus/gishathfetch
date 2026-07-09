package shopifysuggest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"mtg-price-checker-sg/gateway"
)

// borosFuryShieldDetail mirrors the real MTG Asia /products/<handle>.js payload
// for "Boros Fury-Shield": every cheap non-foil condition is out of stock and
// the only purchasable variant is the Near Mint Foil at 1.40. Predictive search
// would otherwise report 0.13 (the Damaged non-foil price_min).
const borosFuryShieldDetail = `{
  "variants": [
    {"id": 1, "title": "Near Mint", "price": 25, "available": false},
    {"id": 2, "title": "Lightly Played", "price": 23, "available": false},
    {"id": 3, "title": "Damaged", "price": 13, "available": false},
    {"id": 4, "title": "Near Mint Foil", "price": 140, "available": true},
    {"id": 5, "title": "Damaged Foil", "price": 70, "available": false}
  ]
}`

func newDetailServer(t *testing.T, body string, status int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/products/") || !strings.HasSuffix(r.URL.Path, ".js") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
}

func baseCardFor(srv *httptest.Server) gateway.Card {
	return gateway.Card{
		Name:    "Boros Fury-Shield [Ravnica: City of Guilds]",
		Url:     srv.URL + "/products/boros-fury-shield-ravnica-city-of-guilds?utm_source=test",
		InStock: true,
		IsFoil:  true, // a tag-derived guess that variant resolution must correct per variant
		Price:   0.13,
		Source:  "MTG Asia",
	}
}

// TestResolveVariantCardsKeepsOnlyInStock verifies that only purchasable
// variants survive and that each card carries the variant's real price,
// condition, and foil flag rather than the predictive-search price_min.
func TestResolveVariantCardsKeepsOnlyInStock(t *testing.T) {
	srv := newDetailServer(t, borosFuryShieldDetail, http.StatusOK)
	defer srv.Close()

	cfg := Config{StoreName: "MTG Asia", BaseURL: srv.URL}
	product := Product{Handle: "boros-fury-shield-ravnica-city-of-guilds"}

	cards := resolveVariantCards(context.Background(), srv.Client(), cfg, product, baseCardFor(srv))

	require.Len(t, cards, 1, "only the in-stock Near Mint Foil variant should remain")
	card := cards[0]
	require.Equal(t, 1.40, card.Price)
	require.Equal(t, "Near Mint", card.Quality)
	require.True(t, card.IsFoil)
	require.True(t, card.InStock)
	require.Contains(t, card.Url, "variant=4")
	require.Contains(t, card.Url, "utm_source=test")
}

// TestResolveVariantCardsCorrectsFoilPerVariant verifies that the per-variant
// foil flag overrides the product-level guess: a non-foil in-stock variant is
// reported as non-foil even when the base card guessed foil from tags.
func TestResolveVariantCardsCorrectsFoilPerVariant(t *testing.T) {
	const body = `{"variants": [
		{"id": 10, "title": "Near Mint", "price": 100, "available": true},
		{"id": 11, "title": "Near Mint Foil", "price": 140, "available": true}
	]}`
	srv := newDetailServer(t, body, http.StatusOK)
	defer srv.Close()

	cfg := Config{StoreName: "MTG Asia", BaseURL: srv.URL}
	product := Product{Handle: "card-handle"}

	cards := resolveVariantCards(context.Background(), srv.Client(), cfg, product, baseCardFor(srv))

	require.Len(t, cards, 2)
	require.Equal(t, 1.00, cards[0].Price)
	require.False(t, cards[0].IsFoil)
	require.Equal(t, 1.40, cards[1].Price)
	require.True(t, cards[1].IsFoil)
}

// TestResolveVariantCardsDropsOnError verifies that a failed detail fetch omits
// the product rather than returning a misleading predictive-search price_min.
func TestResolveVariantCardsDropsOnError(t *testing.T) {
	srv := newDetailServer(t, "", http.StatusInternalServerError)
	defer srv.Close()

	cfg := Config{StoreName: "MTG Asia", BaseURL: srv.URL}
	product := Product{Handle: "card-handle"}
	base := baseCardFor(srv)

	cards := resolveVariantCards(context.Background(), srv.Client(), cfg, product, base)

	require.Empty(t, cards)
}

// TestResolveVariantCardsDropsWhenAllOutOfStock verifies that a product whose
// every variant is out of stock yields no cards, even though predictive search
// flagged the product as available.
func TestResolveVariantCardsDropsWhenAllOutOfStock(t *testing.T) {
	const body = `{"variants": [
		{"id": 20, "title": "Near Mint", "price": 100, "available": false},
		{"id": 21, "title": "Damaged", "price": 13, "available": false}
	]}`
	srv := newDetailServer(t, body, http.StatusOK)
	defer srv.Close()

	cfg := Config{StoreName: "MTG Asia", BaseURL: srv.URL}
	product := Product{Handle: "card-handle"}

	cards := resolveVariantCards(context.Background(), srv.Client(), cfg, product, baseCardFor(srv))
	require.Empty(t, cards)
}
