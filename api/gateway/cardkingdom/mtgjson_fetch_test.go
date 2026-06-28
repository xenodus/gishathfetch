package cardkingdom

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const sampleAllPricesToday = `{
  "meta": {"date": "2026-06-28"},
  "data": {
    "uuid-bolt-4ed": {
      "paper": {
        "cardkingdom": {
          "retail": {
            "normal": {"2026-06-28": 1.49}
          }
        }
      }
    },
    "uuid-bolt-mm-foil": {
      "paper": {
        "cardkingdom": {
          "retail": {
            "foil": {"2026-06-28": 3.99}
          }
        }
      }
    },
    "uuid-no-ck": {
      "paper": {
        "tcgplayer": {
          "retail": {
            "normal": {"2026-06-28": 0.15}
          }
        }
      }
    }
  }
}`

const sampleAllPrintings = `{
  "meta": {"date": "2026-06-28"},
  "data": {
    "4ED": {
      "name": "Fourth Edition",
      "cards": [
        {
          "uuid": "uuid-bolt-4ed",
          "name": "Lightning Bolt",
          "purchaseUrls": {
            "cardKingdom": "https://www.cardkingdom.com/mtg/fourth-edition/lightning-bolt"
          }
        }
      ]
    },
    "MM2": {
      "name": "Modern Masters 2015",
      "cards": [
        {
          "uuid": "uuid-bolt-mm-foil",
          "name": "Lightning Bolt",
          "purchaseUrls": {
            "cardKingdom": "https://www.cardkingdom.com/mtg/modern-masters-2015/lightning-bolt",
            "cardKingdomFoil": "https://www.cardkingdom.com/mtg/modern-masters-2015/lightning-bolt-foil"
          }
        }
      ]
    }
  }
}`

func TestParseCKPricesByUUID(t *testing.T) {
	prices, updatedAt, err := parseCKPricesByUUID([]byte(sampleAllPricesToday))
	require.NoError(t, err)
	require.Equal(t, time.Date(2026, 6, 28, 0, 0, 0, 0, time.UTC), updatedAt)
	require.InDelta(t, 1.49, prices["uuid-bolt-4ed"].normal, 0.001)
	require.InDelta(t, 3.99, prices["uuid-bolt-mm-foil"].foil, 0.001)
	_, ok := prices["uuid-no-ck"]
	require.False(t, ok)
}

func TestDecodeAllPrintingsSets(t *testing.T) {
	prices := map[string]ckUUIDPrice{
		"uuid-bolt-4ed":     {normal: 1.49},
		"uuid-bolt-mm-foil": {foil: 3.99},
	}
	updatedAt := time.Date(2026, 6, 28, 0, 0, 0, 0, time.UTC)

	cheapest, err := decodeAllPrintingsSets(
		json.NewDecoder(strings.NewReader(sampleAllPrintings)),
		prices,
		updatedAt,
	)
	require.NoError(t, err)
	require.Len(t, cheapest, 1)

	listing := cheapest["lightning bolt"]
	require.Equal(t, "Lightning Bolt", listing.CardName)
	require.Equal(t, "Fourth Edition", listing.Edition)
	require.InDelta(t, 1.49, listing.PriceUsd, 0.001)
	require.False(t, listing.IsFoil)
	require.Equal(t, "https://www.cardkingdom.com/mtg/fourth-edition/lightning-bolt", listing.URL)
	require.Equal(t, updatedAt.Format(time.RFC3339), listing.UpdatedAt)
}

func TestFetchCheapestFromMTGJSON_FromTestServers(t *testing.T) {
	pricesServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = writeBzip2(w, []byte(sampleAllPricesToday))
	}))
	defer pricesServer.Close()

	printingsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = writeBzip2(w, []byte(sampleAllPrintings))
	}))
	defer printingsServer.Close()

	t.Setenv("MTGJSON_ALL_PRICES_TODAY_URL", pricesServer.URL)
	t.Setenv("MTGJSON_ALL_PRINTINGS_URL", printingsServer.URL)

	cheapest, err := fetchCheapestFromMTGJSON(context.Background())
	require.NoError(t, err)
	require.Len(t, cheapest, 1)
	require.InDelta(t, 1.49, cheapest["lightning bolt"].PriceUsd, 0.001)
}

func writeBzip2(w http.ResponseWriter, raw []byte) (int, error) {
	cmd := exec.Command("bzip2", "-c")
	cmd.Stdin = bytes.NewReader(raw)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return 0, err
	}
	return w.Write(out.Bytes())
}
