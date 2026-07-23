package binderpos

import (
	"fmt"
	"testing"

	"mtg-price-checker-sg/gateway"

	"github.com/stretchr/testify/require"
)

func TestShopifyVariantID(t *testing.T) {
	id, ok := shopifyVariantID("gid://shopify/ProductVariant/42851708993708")
	require.True(t, ok)
	require.Equal(t, int64(42851708993708), id)

	id, ok = shopifyVariantID("12345")
	require.True(t, ok)
	require.Equal(t, int64(12345), id)

	_, ok = shopifyVariantID("gid://shopify/Product/1")
	require.False(t, ok)
}

func TestMapGraphQLProduct(t *testing.T) {
	product := &graphQLProduct{
		Title:            "Abrade [Foundations]",
		Handle:           "abrade-foundations",
		AvailableForSale: true,
		ProductType:      "MTG Single",
		Tags:             []string{"Foundations"},
		FeaturedImage: &struct {
			URL string `json:"url"`
		}{URL: "https://cdn.shopify.com/abrade.png"},
	}
	product.Variants.Edges = []struct {
		Node *graphQLVariant `json:"node"`
	}{
		{Node: &graphQLVariant{
			ID:               "gid://shopify/ProductVariant/111",
			Title:            "Near Mint",
			AvailableForSale: true,
			Price:            struct {
				Amount string `json:"amount"`
			}{Amount: "0.50"},
		}},
		{Node: &graphQLVariant{
			ID:               "gid://shopify/ProductVariant/222",
			Title:            "Lightly Played",
			AvailableForSale: false,
			Price:            struct {
				Amount string `json:"amount"`
			}{Amount: "0.40"},
		}},
		{Node: &graphQLVariant{
			ID:               "gid://shopify/ProductVariant/333",
			Title:            "Near Mint Foil",
			AvailableForSale: true,
			Price:            struct {
				Amount string `json:"amount"`
			}{Amount: "0.00"},
		}},
	}

	cards := mapGraphQLProduct(3, "Hideout", "https://hideoutcg.com", product)
	require.Len(t, cards, 1)
	require.Equal(t, "Abrade", cards[0].Name)
	require.Equal(t, "Near Mint", cards[0].Quality)
	require.False(t, cards[0].IsFoil)
	require.Equal(t, 0.50, cards[0].Price)
	require.Equal(t, []string{"Foundations"}, cards[0].ExtraInfo)
	require.Contains(t, cards[0].Url, "variant=111")
	require.Contains(t, cards[0].Url, "utm_source=")
}

func TestMapGraphQLProductSkipsNonMTG(t *testing.T) {
	product := &graphQLProduct{
		Title:            "Pikachu",
		Handle:           "pikachu",
		AvailableForSale: true,
		ProductType:      "Pokemon Single",
	}
	product.Variants.Edges = []struct {
		Node *graphQLVariant `json:"node"`
	}{
		{Node: &graphQLVariant{
			ID:               "gid://shopify/ProductVariant/1",
			Title:            "Near Mint",
			AvailableForSale: true,
			Price:            struct {
				Amount string `json:"amount"`
			}{Amount: "1.00"},
		}},
	}
	require.Empty(t, mapGraphQLProduct(2, "Hideyoshi", "https://hideyoshitcg.com", product))
}

func TestRunFallbackAttemptsGraphQLIrrelevantFallsBackToScrap(t *testing.T) {
	sequence := []string{}
	cards, err := runFallbackAttempts(
		fallbackAttempt{strategy: "graphql-dedicated", family: strategyFamilyGraphQL, fn: func() ([]gateway.Card, error) {
			sequence = append(sequence, "graphql-dedicated")
			return nil, fmt.Errorf("Hideyoshi graphql: results do not match %q", "lightning bolt")
		}},
		fallbackAttempt{strategy: "scrap-dedicated", family: strategyFamilyScrap, fn: func() ([]gateway.Card, error) {
			sequence = append(sequence, "scrap-dedicated")
			return []gateway.Card{{Name: "Lightning Bolt [Unlimited Edition]"}}, nil
		}},
	)
	require.NoError(t, err)
	require.Len(t, cards, 1)
	require.Equal(t, "Lightning Bolt [Unlimited Edition]", cards[0].Name)
	require.Equal(t, []string{"graphql-dedicated", "scrap-dedicated"}, sequence)
}

func TestStorefrontGraphQLSearchQuery(t *testing.T) {
	require.Equal(t, "Abrade mtg", storefrontGraphQLSearchQuery(2, "Abrade"))
	require.Equal(t, `product_type:"MTG Single Cards" AND Abrade`, storefrontGraphQLSearchQuery(4, "Abrade"))
	require.Equal(t, "", storefrontGraphQLSearchQuery(2, "  "))
}

func TestRunFallbackAttemptsGraphQLEmptyIsFinal(t *testing.T) {
	sequence := []string{}
	cards, err := runFallbackAttempts(
		fallbackAttempt{strategy: "graphql-dedicated", family: strategyFamilyGraphQL, fn: func() ([]gateway.Card, error) {
			sequence = append(sequence, "graphql-dedicated")
			return nil, nil
		}},
		fallbackAttempt{strategy: "scrap-dedicated", family: strategyFamilyScrap, fn: func() ([]gateway.Card, error) {
			t.Fatal("scrap should not run after empty graphql")
			return nil, nil
		}},
	)
	require.NoError(t, err)
	require.Empty(t, cards)
	require.Equal(t, []string{"graphql-dedicated"}, sequence)
}

func TestRunFallbackAttemptsGraphQLErrorFallsBackToScrap(t *testing.T) {
	sequence := []string{}
	cards, err := runFallbackAttempts(
		fallbackAttempt{strategy: "graphql-dedicated", family: strategyFamilyGraphQL, fn: func() ([]gateway.Card, error) {
			sequence = append(sequence, "graphql-dedicated")
			return nil, errTest("403 Forbidden")
		}},
		fallbackAttempt{strategy: "graphql-direct", family: strategyFamilyGraphQL, fn: func() ([]gateway.Card, error) {
			sequence = append(sequence, "graphql-direct")
			return nil, errTest("403 Forbidden")
		}},
		fallbackAttempt{strategy: "scrap-dedicated", family: strategyFamilyScrap, fn: func() ([]gateway.Card, error) {
			sequence = append(sequence, "scrap-dedicated")
			return []gateway.Card{{Name: "from-scrap"}}, nil
		}},
	)
	require.NoError(t, err)
	require.Len(t, cards, 1)
	require.Equal(t, "from-scrap", cards[0].Name)
	require.Equal(t, []string{"graphql-dedicated", "graphql-direct", "scrap-dedicated"}, sequence)
}

func TestRunFallbackAttemptsGraphQL5xxIsFinal(t *testing.T) {
	sequence := []string{}
	cards, err := runFallbackAttempts(
		fallbackAttempt{strategy: "graphql-dedicated", family: strategyFamilyGraphQL, fn: func() ([]gateway.Card, error) {
			sequence = append(sequence, "graphql-dedicated")
			return nil, errTest("503 Service Unavailable")
		}},
		fallbackAttempt{strategy: "scrap-dedicated", family: strategyFamilyScrap, fn: func() ([]gateway.Card, error) {
			t.Fatal("scrap should not run after graphql 5xx")
			return nil, nil
		}},
	)
	require.Error(t, err)
	require.Empty(t, cards)
	require.Equal(t, []string{"graphql-dedicated"}, sequence)
}

type errTest string

func (e errTest) Error() string { return string(e) }
