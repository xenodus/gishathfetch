package binderpos

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var shopifyDomainPattern = regexp.MustCompile(`Shopify\.shop\s*=\s*"([^"]+)"`)

type supportedGame struct {
	GameID   string `json:"gameId"`
	GameName string `json:"gameName"`
}

func Test_StorefrontSupportedGamesEndpoint_ExistsForGreyOgreAndMtgAsia(t *testing.T) {
	if os.Getenv("RUN_BINDERPOS_LIVE_TESTS") != "1" {
		t.Skip("set RUN_BINDERPOS_LIVE_TESTS=1 to run live BinderPOS portal / store HTML checks")
	}
	client := &http.Client{Timeout: 20 * time.Second}

	tests := []struct {
		name     string
		storeURL string
	}{
		{
			name:     "grey ogre games",
			storeURL: "https://www.greyogregames.com",
		},
		{
			name:     "mtg asia",
			storeURL: "https://www.mtg-asia.com",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()

			shopifyDomain, err := fetchShopifyShopDomain(ctx, client, test.storeURL)
			require.NoError(t, err)
			require.NotEmpty(t, shopifyDomain)

			storefrontAPIURL := fmt.Sprintf(
				"https://portal.binderpos.com/external/shopify/supportedGames?storeUrl=%s",
				url.QueryEscape(shopifyDomain),
			)
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, storefrontAPIURL, nil)
			require.NoError(t, err)
			req.Header.Set("User-Agent", "mtg-price-checker-sg-integration-test")

			res, err := client.Do(req)
			require.NoError(t, err)
			defer res.Body.Close()

			require.Equal(t, http.StatusOK, res.StatusCode)
			require.Contains(t, strings.ToLower(res.Header.Get("Content-Type")), "application/json")

			var games []supportedGame
			require.NoError(t, json.NewDecoder(res.Body).Decode(&games))
			require.NotEmpty(t, games)

			hasMTGGame := false
			for _, game := range games {
				if strings.HasPrefix(strings.ToLower(game.GameID), "mtg") {
					hasMTGGame = true
					break
				}
			}
			require.True(t, hasMTGGame, "expected BinderPOS supported games to include an mtg* game id")
		})
	}
}

func fetchShopifyShopDomain(ctx context.Context, client *http.Client, storeURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, storeURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "mtg-price-checker-sg-integration-test")

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %d while loading %s", res.StatusCode, storeURL)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	matches := shopifyDomainPattern.FindStringSubmatch(string(body))
	if len(matches) < 2 || strings.TrimSpace(matches[1]) == "" {
		return "", fmt.Errorf("shopify domain not found on %s", storeURL)
	}

	return strings.TrimSpace(matches[1]), nil
}
