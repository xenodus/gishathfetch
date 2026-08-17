package cardscentral

import (
	"context"
	"os"
	"strings"
	"testing"

	"mtg-price-checker-sg/gateway/gatewaytest"
	"mtg-price-checker-sg/pkg/config"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

func init() {
	_ = godotenv.Load("../../.env")
}

func Test_Search(t *testing.T) {
	if strings.TrimSpace(os.Getenv(config.CardsCentralKeyEnv)) == "" {
		t.Skip("CARDS_CENTRAL_KEY not set")
	}

	s := NewLGS()
	result, err := s.Search(context.Background(), "lightning bolt")
	require.NoError(t, err)

	if len(result) > 0 {
		for _, card := range result {
			if card.InStock {
				require.NotEmpty(t, card.Name)
				require.Equal(t, StoreName, card.Source)
				require.Contains(t, card.Url, StoreBaseURL+"/card/")
				require.NotEmpty(t, card.Img)
				require.Greater(t, card.Price, float64(0))
				require.NotEmpty(t, card.Quality)
				require.NotEmpty(t, card.ExtraInfo)
			}
		}
		return
	}

	gatewaytest.RequireCardsCentralAPIStructure(t, context.Background(), StoreBaseURL, "lightning bolt")
}
