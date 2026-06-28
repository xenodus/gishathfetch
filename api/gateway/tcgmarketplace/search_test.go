package tcgmarketplace

import (
	"context"
	"os"
	"testing"

	"mtg-price-checker-sg/gateway/gatewaytest"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

func init() {
	_ = godotenv.Load("../../.env")
}

func TestIsSurgeFoil(t *testing.T) {
	tests := map[string]struct {
		extraInfo []string
		name      string
		want      bool
	}{
		"surge foil in name": {
			name: "Abaddon the Despoiler (V1)(Surge Foil)",
			want: true,
		},
		"surge foil in extra info": {
			extraInfo: []string{"[Warhammer 40,000 Commander]", "[Surge Foil]"},
			name:      "Abaddon the Despoiler",
			want:      true,
		},
		"non-foil card": {
			extraInfo: []string{"[Double Masters]"},
			name:      "Abrade",
			want:      false,
		},
		"etched foil not surge": {
			extraInfo: []string{"[Commander Masters]"},
			name:      "Deflecting Swat (V2)(Etched foil)",
			want:      false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, tt.want, isSurgeFoil(tt.extraInfo, tt.name))
		})
	}
}

func Test_Search(t *testing.T) {
	s := NewLGS()
	result, err := s.Search(context.Background(), "abrade")
	gatewaytest.RequireSearchOrProbe(t, err, result, gatewaytest.CardExpect{
		URLContains: StoreBaseURL + "/product/B/",
	}, func(t *testing.T, ctx context.Context) {
		gatewaytest.RequireTCGMarketplaceAPIStructure(t, ctx, os.Getenv(accessTokenKey), "abrade")
	})
}
