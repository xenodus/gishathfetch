package tcgmarketplace

import (
	"context"
	"testing"

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
	require.NoError(t, err)
	require.True(t, len(result) > 0)

	for _, card := range result {
		if card.InStock {
			require.NotEmpty(t, card.Name)
			require.NotEmpty(t, card.Source)
			require.NotEmpty(t, card.Url)
			require.NotEmpty(t, card.Img)
			require.NotEmpty(t, card.Price)
			require.Contains(t, card.Url, StoreBaseURL+"/product/B/")
		}
	}
}
