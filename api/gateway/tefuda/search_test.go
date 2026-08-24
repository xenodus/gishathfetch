package tefuda

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

func Test_MtgSinglesSearchQuery(t *testing.T) {
	require.Equal(t, `product_type:"Magic: The Gathering Singles" AND Polluted Delta`, mtgSinglesSearchQuery("Polluted Delta"))
}

func Test_IsMTGSingleProductType(t *testing.T) {
	require.True(t, isMTGSingleProductType(storefrontMTGType))
	require.False(t, isMTGSingleProductType("MTG Sealed"))
	require.False(t, isMTGSingleProductType("Pokemon Singles"))
	require.False(t, isMTGSingleProductType(""))
}

func Test_ParseProductCardSkipsSoldOut(t *testing.T) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(soldOutProductHTML))
	require.NoError(t, err)

	card, ok := parseProductCard(doc.Find("ul.product-grid li").First(), StoreName)
	require.False(t, ok)
	require.Empty(t, card.Name)
}

func Test_ParseProductCardKeepsInStock(t *testing.T) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(inStockProductHTML))
	require.NoError(t, err)

	card, ok := parseProductCard(doc.Find("ul.product-grid li").First(), StoreName)
	require.True(t, ok)
	require.Equal(t, "Polluted Delta", card.Name)
	require.True(t, card.InStock)
	require.Equal(t, 38.50, card.Price)
	require.Equal(t, []string{"[KTK]", "[239]"}, card.ExtraInfo)
	require.Contains(t, card.Url, "utm_source="+config.UtmSource)
}

func Test_ParseTefudaProductTitle(t *testing.T) {
	name, extra, foil := parseTefudaProductTitle("Belladonna Took [HOB - 4] R The Hobbit - Foil")
	require.Equal(t, "Belladonna Took", name)
	require.Equal(t, []string{"[HOB]", "[4]"}, extra)
	require.True(t, foil)

	name, extra, foil = parseTefudaProductTitle("Polluted Delta [KTK - 239] R Khans of Tarkir")
	require.Equal(t, "Polluted Delta", name)
	require.Equal(t, []string{"[KTK]", "[239]"}, extra)
	require.False(t, foil)

	name, extra, foil = parseTefudaProductTitle("Lightning Bolt (Foil) [FIN]")
	require.Equal(t, "Lightning Bolt (Foil)", name)
	require.Equal(t, []string{"[FIN]"}, extra)
	require.True(t, foil)

	name, extra, foil = parseTefudaProductTitle("The One Ring [LTR]")
	require.Equal(t, "The One Ring", name)
	require.Equal(t, []string{"[LTR]"}, extra)
	require.False(t, foil)
}

func Test_ParseNameAndFoil(t *testing.T) {
	name, foil := parseNameAndFoil("Lightning Bolt (Foil) [FIN]")
	require.Equal(t, "Lightning Bolt (Foil) [FIN]", name)
	require.True(t, foil)

	name, foil = parseNameAndFoil("Polluted Delta [KTK - 239] R Khans of Tarkir")
	require.Equal(t, "Polluted Delta [KTK - 239] R Khans of Tarkir", name)
	require.False(t, foil)

	name, foil = parseNameAndFoil("Belladonna Took [HOB - 4] R The Hobbit - Foil")
	require.Equal(t, "Belladonna Took [HOB - 4] R The Hobbit", name)
	require.True(t, foil)
}

func Test_HandleIndicatesFoil(t *testing.T) {
	require.True(t, handleIndicatesFoil("belladonna-took-hob-4-foil-en"))
	require.False(t, handleIndicatesFoil("belladonna-took-hob-4-normal-en"))
	require.False(t, handleIndicatesFoil("some-card-non-foil-en"))
}

func Test_QualityFromVariantTitle(t *testing.T) {
	require.Equal(t, "Near Mint", qualityFromVariantTitle("Near Mint / English / Normal"))
	require.Equal(t, "Lightly Played", qualityFromVariantTitle("Lightly Played / English / Foil"))
}

func Test_MapGraphQLProductSkipsSealed(t *testing.T) {
	product := &graphQLProduct{
		Title:            "Teenage Mutant Ninja Turtles - Play Booster Display",
		Handle:           "tmnt-display",
		AvailableForSale: true,
		ProductType:      "MTG Sealed",
	}
	product.Variants.Edges = []struct {
		Node *graphQLVariant `json:"node"`
	}{
		{Node: &graphQLVariant{
			Title:            "New",
			AvailableForSale: true,
			Price:            struct {
				Amount string `json:"amount"`
			}{Amount: "190.0"},
		}},
	}

	require.Empty(t, mapGraphQLProduct(StoreName, product))
}

func Test_MapGraphQLProductDetectsTefudaTitleSuffixFoil(t *testing.T) {
	product := &graphQLProduct{
		Title:            "Belladonna Took [HOB - 4] R The Hobbit - Foil",
		Handle:           "belladonna-took-hob-4-foil-en",
		AvailableForSale: true,
		ProductType:      storefrontMTGType,
		Tags:             []string{"English", "Rare", "The Hobbit"},
	}
	product.Variants.Edges = []struct {
		Node *graphQLVariant `json:"node"`
	}{
		{Node: &graphQLVariant{
			Title:            "Near Mint",
			AvailableForSale: true,
			Price: struct {
				Amount string `json:"amount"`
			}{Amount: "3.5"},
		}},
	}

	cards := mapGraphQLProduct(StoreName, product)
	require.Len(t, cards, 1)
	require.Equal(t, "Belladonna Took", cards[0].Name)
	require.Equal(t, []string{"[HOB]", "[4]"}, cards[0].ExtraInfo)
	require.True(t, cards[0].IsFoil)
	require.Equal(t, "Near Mint", cards[0].Quality)
}

func Test_MapGraphQLProductKeepsSingles(t *testing.T) {
	product := &graphQLProduct{
		Title:            "Polluted Delta [KTK - 239] R Khans of Tarkir",
		Handle:           "polluted-delta-ktk-239",
		AvailableForSale: true,
		ProductType:      storefrontMTGType,
		Tags:             []string{"Khans of Tarkir", "Rare", "Land"},
		FeaturedImage: &struct {
			URL string `json:"url"`
		}{URL: "https://cdn.shopify.com/polluted-delta.png"},
	}
	product.Variants.Edges = []struct {
		Node *graphQLVariant `json:"node"`
	}{
		{Node: &graphQLVariant{
			Title:            "Near Mint / English / Normal",
			AvailableForSale: true,
			Price:            struct {
				Amount string `json:"amount"`
			}{Amount: "38.5"},
		}},
	}

	cards := mapGraphQLProduct(StoreName, product)
	require.Len(t, cards, 1)
	require.Equal(t, "Polluted Delta", cards[0].Name)
	require.Equal(t, []string{"[KTK]", "[239]"}, cards[0].ExtraInfo)
	require.Equal(t, "Near Mint", cards[0].Quality)
	require.Equal(t, 38.5, cards[0].Price)
	require.Contains(t, cards[0].Url, "/products/polluted-delta-ktk-239")
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
			require.Equal(t, "product", r.URL.Query().Get("type"))
			require.Contains(t, r.URL.Query().Get("q"), `product_type:"Magic: The Gathering Singles" AND`)
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
	cards, err := store.Search(context.Background(), "Polluted Delta")
	require.NoError(t, err)
	require.True(t, sawGraphQL)
	require.True(t, sawHTML)
	require.Len(t, cards, 1)
	require.Equal(t, "Polluted Delta", cards[0].Name)
	require.Equal(t, []string{"[KTK]", "[239]"}, cards[0].ExtraInfo)
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
						Title:            "Polluted Delta [KTK - 239] R Khans of Tarkir",
						Handle:           "polluted-delta-ktk-239",
						AvailableForSale: true,
						ProductType:      storefrontMTGType,
					},
				}},
			},
		},
	}
	payload.Data.Search.Edges[0].Node.Variants.Edges = []struct {
		Node *graphQLVariant `json:"node"`
	}{
		{Node: &graphQLVariant{
			Title:            "Near Mint / English / Normal",
			AvailableForSale: true,
			Price:            struct {
				Amount string `json:"amount"`
			}{Amount: "38.5"},
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
	cards, err := store.Search(context.Background(), "Polluted Delta")
	require.NoError(t, err)
	require.False(t, sawHTML)
	require.Len(t, cards, 1)
	require.Equal(t, "Near Mint", cards[0].Quality)
}

func Test_Search(t *testing.T) {
	s := NewLGS()
	result, err := s.Search(context.Background(), "Polluted Delta")
	gatewaytest.RequireSearchOrProbe(t, err, result, gatewaytest.CardExpect{
		URLContains: StoreBaseURL + "/products/",
	}, func(t *testing.T, ctx context.Context) {
		gatewaytest.RequireTefudaSearchStructure(t, ctx, StoreBaseURL, StoreSearchPath, "Polluted Delta")
	})
}
