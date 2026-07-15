package fivemana

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"mtg-price-checker-sg/gateway/gatewaytest"
	"mtg-price-checker-sg/pkg/config"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/require"
)

func Test_ParseProductCardSkipsSoldOut(t *testing.T) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(soldOutProductHTML))
	require.NoError(t, err)

	card, ok := parseProductCard(doc.Find("ul.product-grid li").First(), StoreName)
	require.False(t, ok, "sold-out listing should be skipped")
	require.Empty(t, card.Name)
}

func Test_ParseProductCardKeepsInStock(t *testing.T) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(inStockProductHTML))
	require.NoError(t, err)

	card, ok := parseProductCard(doc.Find("ul.product-grid li").First(), StoreName)
	require.True(t, ok)
	require.Equal(t, "Abrade [Foundations]", card.Name)
	require.True(t, card.InStock)
	require.Equal(t, 0.40, card.Price)
	require.Equal(t, []string{"[Foundations]"}, card.ExtraInfo)
	require.Contains(t, card.Url, "utm_source="+config.UtmSource)
}

func Test_ParseNameAndFoil(t *testing.T) {
	name, foil := parseNameAndFoil("Electro's Bolt [Marvel's Spider-Man] [Foil]")
	require.Equal(t, "Electro's Bolt [Marvel's Spider-Man]", name)
	require.True(t, foil)

	name, foil = parseNameAndFoil("Abrade [Foundations]")
	require.Equal(t, "Abrade [Foundations]", name)
	require.False(t, foil)
}

func Test_QualityFromVariantTitle(t *testing.T) {
	require.Equal(t, "Near Mint", qualityFromVariantTitle("Near Mint"))
	require.Equal(t, "Near Mint", qualityFromVariantTitle("Near Mint Foil"))
	require.Equal(t, "Lightly Played", qualityFromVariantTitle("LP"))
}

func Test_MapGraphQLProduct(t *testing.T) {
	product := &graphQLProduct{
		Title:            "Abrade [Foundations]",
		Handle:           "abrade-foundations",
		AvailableForSale: true,
		ProductType:      storefrontMTGType,
		Tags:             []string{"Foundations", "Foundations Non-Foil", "Red"},
		FeaturedImage:    &struct {
			URL string `json:"url"`
		}{URL: "https://cdn.shopify.com/abrade.png"},
	}
	product.Variants.Edges = []struct {
		Node *graphQLVariant `json:"node"`
	}{
		{Node: &graphQLVariant{
			Title:            "Near Mint",
			AvailableForSale: true,
			Price:            struct {
				Amount string `json:"amount"`
			}{Amount: "0.40"},
		}},
		{Node: &graphQLVariant{
			Title:            "Near Mint Foil",
			AvailableForSale: true,
			Price:            struct {
				Amount string `json:"amount"`
			}{Amount: "0.00"},
		}},
		{Node: &graphQLVariant{
			Title:            "Lightly Played",
			AvailableForSale: false,
			Price:            struct {
				Amount string `json:"amount"`
			}{Amount: "0.20"},
		}},
	}

	cards := mapGraphQLProduct(StoreName, product)
	require.Len(t, cards, 1)
	require.Equal(t, "Abrade [Foundations]", cards[0].Name)
	require.Equal(t, "Near Mint", cards[0].Quality)
	require.False(t, cards[0].IsFoil)
	require.Equal(t, 0.40, cards[0].Price)
	require.Equal(t, []string{"[Foundations]"}, cards[0].ExtraInfo)
	require.Contains(t, cards[0].Url, "/products/abrade-foundations")
	require.Contains(t, cards[0].Url, "utm_source=")
}

func Test_MapGraphQLProductFoilVariant(t *testing.T) {
	product := &graphQLProduct{
		Title:            "Sol Ring (0358) (Surge Foil) [FINAL FANTASY Commander]",
		Handle:           "sol-ring-surge",
		AvailableForSale: true,
		ProductType:      storefrontMTGType,
	}
	product.Variants.Edges = []struct {
		Node *graphQLVariant `json:"node"`
	}{
		{Node: &graphQLVariant{
			Title:            "Near Mint Foil",
			AvailableForSale: true,
			Price:            struct {
				Amount string `json:"amount"`
			}{Amount: "15.4"},
		}},
	}

	cards := mapGraphQLProduct(StoreName, product)
	require.Len(t, cards, 1)
	require.True(t, cards[0].IsFoil)
	require.Equal(t, "Near Mint", cards[0].Quality)
	require.Equal(t, 15.4, cards[0].Price)
}

func Test_SearchFallsBackToHTMLWhenGraphQLFails(t *testing.T) {
	var sawGraphQL, sawHTML bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "graphql"):
			sawGraphQL = true
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"errors":[{"message":"boom"}]}`))
		case strings.Contains(r.URL.Path, "search"):
			sawHTML = true
			require.Equal(t, storefrontSearchSectionID, r.URL.Query().Get("section_id"))
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(inStockProductHTML))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	store := Store{
		Name:       StoreName,
		BaseUrl:    server.URL,
		SearchPath: StoreSearchPath,
	}
	cards, err := store.Search(context.Background(), "Abrade")
	require.NoError(t, err)
	require.True(t, sawGraphQL)
	require.True(t, sawHTML)
	require.Len(t, cards, 1)
	require.Equal(t, "Abrade [Foundations]", cards[0].Name)
}

func Test_SearchDoesNotFallbackToHTMLWhenGraphQL5xx(t *testing.T) {
	var sawGraphQL, sawHTML bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "graphql"):
			sawGraphQL = true
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte(`{"errors":[{"message":"boom"}]}`))
		case strings.Contains(r.URL.Path, "search"):
			sawHTML = true
			http.NotFound(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	store := Store{
		Name:       StoreName,
		BaseUrl:    server.URL,
		SearchPath: StoreSearchPath,
	}
	_, err := store.Search(context.Background(), "Abrade")
	require.Error(t, err)
	require.True(t, sawGraphQL)
	require.False(t, sawHTML)
}

func Test_SearchUsesGraphQLWhenHealthy(t *testing.T) {
	payload := graphQLResponse{
		Data: &struct {
			Search *struct {
				Edges []graphQLEdge `json:"edges"`
			} `json:"search"`
		}{
			Search: &struct {
				Edges []graphQLEdge `json:"edges"`
			}{
				Edges: []graphQLEdge{{
					Node: &graphQLProduct{
						Title:            "Abrade [Foundations]",
						Handle:           "abrade-foundations",
						AvailableForSale: true,
						ProductType:      storefrontMTGType,
						FeaturedImage: &struct {
							URL string `json:"url"`
						}{URL: "https://cdn.shopify.com/abrade.png"},
					},
				}},
			},
		},
	}
	payload.Data.Search.Edges[0].Node.Variants.Edges = []struct {
		Node *graphQLVariant `json:"node"`
	}{
		{Node: &graphQLVariant{
			Title:            "Near Mint",
			AvailableForSale: true,
			Price:            struct {
				Amount string `json:"amount"`
			}{Amount: "0.4"},
		}},
	}

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	var sawHTML bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "graphql"):
			require.Equal(t, storefrontAccessToken, r.Header.Get("X-Shopify-Storefront-Access-Token"))
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(body)
		case strings.Contains(r.URL.Path, "search"):
			sawHTML = true
			http.Error(w, "should not hit HTML", http.StatusInternalServerError)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	store := Store{
		Name:       StoreName,
		BaseUrl:    server.URL,
		SearchPath: StoreSearchPath,
	}
	cards, err := store.Search(context.Background(), "Abrade")
	require.NoError(t, err)
	require.False(t, sawHTML)
	require.Len(t, cards, 1)
	require.Equal(t, "Near Mint", cards[0].Quality)
	require.Equal(t, 0.4, cards[0].Price)
}

func Test_Search(t *testing.T) {
	s := NewLGS()
	result, err := s.Search(context.Background(), "Abrade")
	gatewaytest.RequireSearchOrProbe(t, err, result, gatewaytest.CardExpect{
		URLContains: StoreBaseURL + "/products/",
	}, func(t *testing.T, ctx context.Context) {
		gatewaytest.RequireFiveManaSearchStructure(t, ctx, StoreBaseURL, StoreSearchPath, "Abrade")
	})
}
