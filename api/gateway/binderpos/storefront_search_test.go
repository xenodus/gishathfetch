package binderpos

import (
	"context"
	"net/http"
	"os"
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
	storeName     string
	baseURL       string
	shopifyDomain string
	searchURL     string
	scrapVariant  int
	query         string
}

func Test_SearchByStorefrontAPI_SupportsAllBinderposStores(t *testing.T) {
	if os.Getenv("RUN_BINDERPOS_LIVE_TESTS") != "1" {
		t.Skip("set RUN_BINDERPOS_LIVE_TESTS=1 to run live storefront API checks against real stores")
	}

	client := &http.Client{Timeout: 20 * time.Second}
	for _, testCase := range binderposStoreSearchCases() {
		t.Run(testCase.storeName, func(t *testing.T) {
			if strings.TrimSpace(testCase.shopifyDomain) == "" {
				t.Skip("decklist API requires a shopify domain mapping")
			}
			cards, err := searchByBinderposDecklistAPI(context.Background(), client, testCase.scrapVariant, testCase.storeName, testCase.baseURL, testCase.shopifyDomain, testCase.query)
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
	if os.Getenv("RUN_BINDERPOS_LIVE_TESTS") != "1" {
		t.Skip("set RUN_BINDERPOS_LIVE_TESTS=1 to run live storefront vs scrape overlap checks")
	}

	client := &http.Client{Timeout: 20 * time.Second}
	for _, testCase := range binderposStoreSearchCases() {
		t.Run(testCase.storeName, func(t *testing.T) {
			if strings.TrimSpace(testCase.shopifyDomain) == "" {
				t.Skip("decklist API requires a shopify domain mapping")
			}
			storefrontCards, err := searchByBinderposDecklistAPI(context.Background(), client, testCase.scrapVariant, testCase.storeName, testCase.baseURL, testCase.shopifyDomain, testCase.query)
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
			storeName:     "Cards Citadel",
			baseURL:       "https://cardscitadel.com",
			shopifyDomain: "card-citadel.myshopify.com",
			searchURL:     "/search?q=*%s*",
			scrapVariant:  1,
			query:         "Abrade",
		},
		{
			storeName:     "Card Affinity",
			baseURL:       "https://card-affinity.com",
			shopifyDomain: "563304-2.myshopify.com",
			searchURL:     "/search?q=%s",
			scrapVariant:  2,
			query:         "Abrade",
		},
		{
			storeName:     "Cardboard Crack Games",
			baseURL:       "https://www.cardboardcrackgames.com",
			shopifyDomain: "cardboardcrackgames.myshopify.com",
			searchURL:     "/search?type=product&q=%s",
			scrapVariant:  2,
			query:         "Abrade",
		},
		{
			storeName:     "Flagship Games",
			baseURL:       "https://www.flagshipgames.sg",
			shopifyDomain: "flagship-games.myshopify.com",
			searchURL:     "/search?type=product&q=%s",
			scrapVariant:  2,
			query:         "Abrade",
		},
		{
			storeName:     "Games Haven",
			baseURL:       "https://www.gameshaventcg.com",
			shopifyDomain: "games-haven-sg.myshopify.com",
			searchURL:     "/search?q=%s",
			scrapVariant:  3,
			query:         "Abrade",
		},
		{
			storeName:     "Grey Ogre Games",
			baseURL:       "https://www.greyogregames.com",
			shopifyDomain: "grey-ogre-games-singapore.myshopify.com",
			searchURL:     "/search?q=%s",
			scrapVariant:  3,
			query:         "Abrade",
		},
		{
			storeName:     "Hideout",
			baseURL:       "https://hideoutcg.com",
			shopifyDomain: "220022-20.myshopify.com",
			searchURL:     "/search?q=%s",
			scrapVariant:  3,
			query:         "Abrade",
		},
		{
			storeName:     "Mana Pro",
			baseURL:       "https://sg-manapro.com",
			shopifyDomain: "mana-pro-sg.myshopify.com",
			searchURL:     "/search?type=product&q=%s",
			scrapVariant:  2,
			query:         "Abrade",
		},
		{
			storeName:     "MTG Asia",
			baseURL:       "https://www.mtg-asia.com",
			shopifyDomain: "mtgasia.myshopify.com",
			searchURL:     "/search?q=%s",
			scrapVariant:  2,
			query:         "Abrade",
		},
		{
			storeName:     "OneMtg",
			baseURL:       "https://onemtg.com.sg",
			shopifyDomain: "one-mtg.myshopify.com",
			searchURL:     "/search?q=%s",
			scrapVariant:  2,
			query:         "Abrade",
		},
		{
			storeName:     "Tefuda",
			baseURL:       "https://tefudagames.com",
			shopifyDomain: "bacc1b-3.myshopify.com",
			searchURL:     "/search?q=%s",
			scrapVariant:  4,
			query:         "smothering tithe",
		},
		{
			storeName:     "Arcane Sanctum",
			baseURL:       "https://arcanesanctumtcg.com",
			shopifyDomain: "",
			searchURL:     "/search?q=%s",
			scrapVariant:  5,
			query:         "signet",
		},
	}
}

func normalizeCardName(name string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(name)), " "))
}
