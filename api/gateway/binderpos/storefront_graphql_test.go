package binderpos

import (
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
		{Node: &graphQLVariant{
			ID:               "gid://shopify/ProductVariant/444",
			Title:            "Default Title",
			AvailableForSale: true,
			Price:            struct {
				Amount string `json:"amount"`
			}{Amount: "1.00"},
		}},
	}

	cards := mapGraphQLProduct(3, "Hideout", "https://hideoutcg.com", product)
	require.Len(t, cards, 2)
	require.Equal(t, "Abrade", cards[0].Name)
	require.Equal(t, "Near Mint", cards[0].Quality)
	require.False(t, cards[0].IsFoil)
	require.Equal(t, 0.50, cards[0].Price)
	require.Equal(t, []string{"Foundations"}, cards[0].ExtraInfo)
	require.Contains(t, cards[0].Url, "variant=111")
	require.Contains(t, cards[0].Url, "utm_source=")

	require.Equal(t, "Abrade", cards[1].Name)
	require.Empty(t, cards[1].Quality)
	require.Equal(t, 1.00, cards[1].Price)
	require.Contains(t, cards[1].Url, "variant=444")
}

func TestMapGraphQLProductDefaultTitleVariantNameForScrapVariant2(t *testing.T) {
	product := &graphQLProduct{
		Title:            "Opt [Ixalan]",
		Handle:           "opt-ixalan",
		AvailableForSale: true,
		ProductType:      "MTG Single",
	}
	product.Variants.Edges = []struct {
		Node *graphQLVariant `json:"node"`
	}{
		{Node: &graphQLVariant{
			ID:               "gid://shopify/ProductVariant/555",
			Title:            "Default Title",
			AvailableForSale: true,
			Price:            struct {
				Amount string `json:"amount"`
			}{Amount: "0.25"},
		}},
	}

	cards := mapGraphQLProduct(2, "Arcane Sanctum", "https://arcanesanctumtcg.com", product)
	require.Len(t, cards, 1)
	require.Equal(t, "Opt [Ixalan]", cards[0].Name)
	require.Empty(t, cards[0].Quality)
}

func TestMapGraphQLProductEmbeddedQualityInTitle(t *testing.T) {
	product := &graphQLProduct{
		Title:            "Reanimate — NM [Breaking News]",
		Handle:           "reanimate-nm-breaking-news",
		AvailableForSale: true,
		ProductType:      "MTG Single",
	}
	product.Variants.Edges = []struct {
		Node *graphQLVariant `json:"node"`
	}{
		{Node: &graphQLVariant{
			ID:               "gid://shopify/ProductVariant/555",
			Title:            "Default Title",
			AvailableForSale: true,
			Price:            struct {
				Amount string `json:"amount"`
			}{Amount: "13.5"},
		}},
	}

	cards := mapGraphQLProduct(2, "Arcane Sanctum", "https://arcanesanctumtcg.com", product)
	require.Len(t, cards, 1)
	require.Equal(t, "Reanimate [Breaking News]", cards[0].Name)
	require.Equal(t, "Near Mint", cards[0].Quality)
	require.False(t, cards[0].IsFoil)
}

func TestMapGraphQLProductEmbeddedQualityFoilInTitle(t *testing.T) {
	product := &graphQLProduct{
		Title:            "Subtlety — NM Foil [Secret Lair Drop]",
		Handle:           "subtlety-nm-foil-secret-lair-drop",
		AvailableForSale: true,
		ProductType:      "MTG Single",
	}
	product.Variants.Edges = []struct {
		Node *graphQLVariant `json:"node"`
	}{
		{Node: &graphQLVariant{
			ID:               "gid://shopify/ProductVariant/556",
			Title:            "Default Title",
			AvailableForSale: true,
			Price:            struct {
				Amount string `json:"amount"`
			}{Amount: "25.0"},
		}},
	}

	cards := mapGraphQLProduct(2, "Arcane Sanctum", "https://arcanesanctumtcg.com", product)
	require.Len(t, cards, 1)
	require.Equal(t, "Subtlety [Secret Lair Drop]", cards[0].Name)
	require.Equal(t, "Near Mint", cards[0].Quality)
	require.True(t, cards[0].IsFoil)
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
