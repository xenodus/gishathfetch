package binderpos

import (
	"context"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"mtg-price-checker-sg/pkg/config"

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
			if len(cards) > 0 {
				for _, card := range cards {
					require.NotEmpty(t, card.Name)
					require.NotEmpty(t, card.Url)
					require.NotEmpty(t, card.Img)
					require.NotEmpty(t, card.Source)
					require.Greater(t, card.Price, float64(0))
				}
				return
			}

			require.NoError(t, ProbeDecklistStructure(context.Background(), testCase.shopifyDomain, testCase.query))
		})
	}
}

func Test_SearchByStorefrontAPI_AndScrapeStructuresRemainCompatible(t *testing.T) {
	if os.Getenv("RUN_BINDERPOS_LIVE_TESTS") != "1" {
		t.Skip("set RUN_BINDERPOS_LIVE_TESTS=1 to run live storefront vs scrape structure checks")
	}

	client := &http.Client{Timeout: 20 * time.Second}
	for _, testCase := range binderposStoreSearchCases() {
		t.Run(testCase.storeName, func(t *testing.T) {
			if strings.TrimSpace(testCase.shopifyDomain) == "" {
				t.Skip("decklist API requires a shopify domain mapping")
			}
			storefrontCards, err := searchByBinderposDecklistAPI(context.Background(), client, testCase.scrapVariant, testCase.storeName, testCase.baseURL, testCase.shopifyDomain, testCase.query)
			require.NoError(t, err)

			scrapedCards, err := New().Scrap(
				context.Background(),
				testCase.scrapVariant,
				testCase.storeName,
				testCase.baseURL,
				testCase.searchURL,
				testCase.query,
			)
			require.NoError(t, err)

			if len(storefrontCards) > 0 && len(scrapedCards) > 0 {
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
				if overlapCount > 0 {
					return
				}
			}

			require.NoError(t, ProbeDecklistStructure(context.Background(), testCase.shopifyDomain, testCase.query))
			require.NoError(t, ProbeScrapeStructure(
				context.Background(),
				testCase.scrapVariant,
				testCase.baseURL,
				testCase.searchURL,
				testCase.query,
			))
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
			storeName:     "Hideyoshi",
			baseURL:       "https://hideyoshitcg.com",
			shopifyDomain: "bposacct-9.myshopify.com",
			searchURL:     "/search?q=%s",
			scrapVariant:  2,
			query:         "Abrade",
		},
		{
			storeName:     "Fyendal Hobby",
			baseURL:       "https://fyendalhobby.com",
			shopifyDomain: "fyendal-hobby.myshopify.com",
			searchURL:     "/search?q=%s",
			scrapVariant:  4,
			query:         "Abrade",
		},
		{
			storeName:     "Arcane Sanctum",
			baseURL:       "https://arcanesanctumtcg.com",
			shopifyDomain: "30uetm-1y.myshopify.com",
			searchURL:     "/search?q=%s",
			scrapVariant:  5,
			query:         "signet",
		},
	}
}

func TestSelectStorefrontStrategies_ScrapOnly(t *testing.T) {
	scrap := [3]storefrontStrategy{
		{name: "scrap-dedicated"},
		{name: "scrap-direct"},
		{name: "scrap-dynamic"},
	}
	decklist := [3]storefrontStrategy{
		{name: "decklist-dedicated"},
		{name: "decklist-direct"},
		{name: "decklist-dynamic"},
	}

	t.Run("scrapOnly with domain keeps scrape strategies only", func(t *testing.T) {
		got := selectStorefrontStrategies(true, "shop.example.com", scrap, decklist)
		assertStrategyOrder(t, []string{"scrap-dedicated", "scrap-direct", "scrap-dynamic"}, strategyNames(got))
	})

	t.Run("empty domain keeps scrape strategies only", func(t *testing.T) {
		got := selectStorefrontStrategies(false, "", scrap, decklist)
		assertStrategyOrder(t, []string{"scrap-dedicated", "scrap-direct", "scrap-dynamic"}, strategyNames(got))
	})

	t.Run("domain without scrapOnly uses decklist and scrap when BinderposScrapOnly is false", func(t *testing.T) {
		previousScrapOnly := config.BinderposScrapOnly
		config.BinderposScrapOnly = false
		t.Cleanup(func() { config.BinderposScrapOnly = previousScrapOnly })

		previousSelector := shouldStartWithDecklist
		shouldStartWithDecklist = func() bool { return true }
		t.Cleanup(func() { shouldStartWithDecklist = previousSelector })

		got := selectStorefrontStrategies(false, "shop.example.com", scrap, decklist)
		assertStrategyOrder(t, []string{
			"decklist-dedicated", "decklist-direct",
			"scrap-dedicated", "scrap-direct",
			"decklist-dynamic", "scrap-dynamic",
		}, strategyNames(got))
	})

	t.Run("BinderposScrapOnly with domain keeps scrape strategies only", func(t *testing.T) {
		previousScrapOnly := config.BinderposScrapOnly
		config.BinderposScrapOnly = true
		t.Cleanup(func() { config.BinderposScrapOnly = previousScrapOnly })

		got := selectStorefrontStrategies(false, "shop.example.com", scrap, decklist)
		assertStrategyOrder(t, []string{"scrap-dedicated", "scrap-direct", "scrap-dynamic"}, strategyNames(got))
	})
}

func normalizeCardName(name string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(name)), " "))
}
