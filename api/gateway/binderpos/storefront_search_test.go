package binderpos

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

func init() {
	_ = godotenv.Load("../../.env")
}

type binderposStoreSearchCase struct {
	storeName    string
	baseURL      string
	searchURL    string
	scrapVariant int
	query        string
}

func Test_SearchByStorefrontAPI_SupportsAllBinderposStores(t *testing.T) {
	previousSelector := shouldUseDecklistEndpoint
	shouldUseDecklistEndpoint = func() bool { return false }
	t.Cleanup(func() { shouldUseDecklistEndpoint = previousSelector })

	client := &http.Client{Timeout: 20 * time.Second}
	for _, testCase := range binderposStoreSearchCases() {
		t.Run(testCase.storeName, func(t *testing.T) {
			cards, err := searchByStorefrontAPIWithClient(context.Background(), client, testCase.scrapVariant, testCase.storeName, testCase.baseURL, testCase.query)
			require.NoError(t, err)
			require.NotEmpty(t, cards)

			for _, card := range cards {
				require.NotEmpty(t, card.Name)
				require.NotEmpty(t, card.Url)
				require.NotEmpty(t, card.Img)
				require.NotEmpty(t, card.Source)
				require.Greater(t, card.Price, float64(0))
			}
		})
	}
}

func Test_SearchByStorefrontAPI_OverlapsLegacyScrapeResults(t *testing.T) {
	previousSelector := shouldUseDecklistEndpoint
	shouldUseDecklistEndpoint = func() bool { return false }
	t.Cleanup(func() { shouldUseDecklistEndpoint = previousSelector })

	client := &http.Client{Timeout: 20 * time.Second}
	for _, testCase := range binderposStoreSearchCases() {
		t.Run(testCase.storeName, func(t *testing.T) {
			storefrontCards, err := searchByStorefrontAPIWithClient(context.Background(), client, testCase.scrapVariant, testCase.storeName, testCase.baseURL, testCase.query)
			require.NoError(t, err)
			require.NotEmpty(t, storefrontCards)

			scrapedCards, err := New().Scrap(
				context.Background(),
				testCase.scrapVariant,
				testCase.storeName,
				testCase.baseURL,
				testCase.searchURL,
				testCase.query,
			)
			require.NoError(t, err)
			require.NotEmpty(t, scrapedCards)

			scrapedNames := make(map[string]struct{}, len(scrapedCards))
			for _, card := range scrapedCards {
				scrapedNames[normalizeCardName(card.Name)] = struct{}{}
			}

			overlapCount := 0
			for _, card := range storefrontCards {
				if _, exists := scrapedNames[normalizeCardName(card.Name)]; exists {
					overlapCount++
				}
			}

			// Ensures storefront API and legacy scraping return materially similar search data.
			require.Greater(t, overlapCount, 0)
		})
	}
}

func binderposStoreSearchCases() []binderposStoreSearchCase {
	return []binderposStoreSearchCase{
		{
			storeName:    "Cards Citadel",
			baseURL:      "https://cardscitadel.com",
			searchURL:    "/search?q=*%s*",
			scrapVariant: 1,
			query:        "Abrade",
		},
		{
			storeName:    "Card Affinity",
			baseURL:      "https://card-affinity.com",
			searchURL:    "/search?q=%s",
			scrapVariant: 2,
			query:        "Abrade",
		},
		{
			storeName:    "Cardboard Crack Games",
			baseURL:      "https://www.cardboardcrackgames.com",
			searchURL:    "/search?type=product&q=%s",
			scrapVariant: 2,
			query:        "Abrade",
		},
		{
			storeName:    "Flagship Games",
			baseURL:      "https://www.flagshipgames.sg",
			searchURL:    "/search?type=product&q=%s",
			scrapVariant: 2,
			query:        "Abrade",
		},
		{
			storeName:    "Games Haven",
			baseURL:      "https://www.gameshaventcg.com",
			searchURL:    "/search?q=%s",
			scrapVariant: 3,
			query:        "Abrade",
		},
		{
			storeName:    "Grey Ogre Games",
			baseURL:      "https://www.greyogregames.com",
			searchURL:    "/search?q=%s",
			scrapVariant: 3,
			query:        "Abrade",
		},
		{
			storeName:    "Hideout",
			baseURL:      "https://hideoutcg.com",
			searchURL:    "/search?q=%s",
			scrapVariant: 3,
			query:        "Abrade",
		},
		{
			storeName:    "Mana Pro",
			baseURL:      "https://sg-manapro.com",
			searchURL:    "/search?type=product&q=%s",
			scrapVariant: 2,
			query:        "Abrade",
		},
		{
			storeName:    "MTG Asia",
			baseURL:      "https://www.mtg-asia.com",
			searchURL:    "/search?q=%s",
			scrapVariant: 2,
			query:        "Abrade",
		},
		{
			storeName:    "OneMtg",
			baseURL:      "https://onemtg.com.sg",
			searchURL:    "/search?q=%s",
			scrapVariant: 2,
			query:        "Abrade",
		},
		{
			storeName:    "Tefuda",
			baseURL:      "https://tefudagames.com",
			searchURL:    "/search?q=%s",
			scrapVariant: 4,
			query:        "smothering tithe",
		},
		{
			storeName:    "Arcane Sanctum",
			baseURL:      "https://arcanesanctumtcg.com",
			searchURL:    "/search?q=%s",
			scrapVariant: 5,
			query:        "signet",
		},
	}
}

func normalizeCardName(name string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(name)), " "))
}
