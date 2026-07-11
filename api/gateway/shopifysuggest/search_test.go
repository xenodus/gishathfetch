package shopifysuggest

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"mtg-price-checker-sg/gateway"
)

const singleVariantDetail = `{"variants": [
	{"id": 1, "title": "Near Mint", "price": 250, "available": true}
]}`

func TestMapProductsVariantResolveLimit(t *testing.T) {
	var detailCalls atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/products/") && strings.HasSuffix(r.URL.Path, ".js") {
			detailCalls.Add(1)
			_, _ = w.Write([]byte(singleVariantDetail))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	products := make([]Product, 0, variantResolveMaxProducts+2)
	for i := range variantResolveMaxProducts + 2 {
		products = append(products, Product{
			Title:       "Card",
			Handle:      fmt.Sprintf("card-%d", i),
			Available:   true,
			ProductType: "MTG Single Cards",
			Price:       "1.00",
		})
	}

	cfg := Config{StoreName: "Test", BaseURL: srv.URL}
	opts := Options{
		Config:          cfg,
		ResolveVariants: true,
		MapProduct: func(cfg Config, product Product) (gateway.Card, bool) {
			return gateway.Card{
				Name:    product.Title,
				Source:  cfg.StoreName,
				Price:   1.00,
				InStock: true,
				Url:     srv.URL + "/products/" + product.Handle,
			}, true
		},
	}

	cards := mapProducts(context.Background(), srv.Client(), opts, products)

	require.Equal(t, int32(variantResolveMaxProducts), detailCalls.Load())
	require.Len(t, cards, variantResolveMaxProducts+2)

	for i, card := range cards {
		if i < variantResolveMaxProducts {
			require.Equal(t, 2.50, card.Price, "product %d should use resolved variant price", i)
			continue
		}
		require.Equal(t, 1.00, card.Price, "product %d should fall back to suggest price", i)
	}
}
