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
    "uuid-tony-80-price": {
      "paper": {
        "cardkingdom": {
          "retail": {
            "normal": {"2026-06-28": 7.49},
            "foil": {"2026-06-28": 8.99}
          }
        }
      }
    },
    "uuid-tony-363-price": {
      "paper": {
        "cardkingdom": {
          "retail": {
            "foil": {"2026-06-28": 44.99}
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
          "number": "162",
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
          "number": "141",
          "purchaseUrls": {
            "cardKingdom": "https://www.cardkingdom.com/mtg/modern-masters-2015/lightning-bolt",
            "cardKingdomFoil": "https://www.cardkingdom.com/mtg/modern-masters-2015/lightning-bolt-foil"
          }
        }
      ]
    }
  }
}`

const sampleAllPrintingsSplitUUID = `{
  "meta": {"date": "2026-06-28"},
  "data": {
    "MSH": {
      "name": "Marvel Super Heroes",
      "cards": [
        {
          "uuid": "uuid-tony-80-url",
          "name": "Tony Stark // The Invincible Iron Man",
          "number": "80",
          "purchaseUrls": {
            "cardKingdom": "https://www.cardkingdom.com/mtg/marvel-super-heroes/tony-stark",
            "cardKingdomFoil": "https://www.cardkingdom.com/mtg/marvel-super-heroes/tony-stark-foil"
          }
        },
        {
          "uuid": "uuid-tony-80-price",
          "name": "Tony Stark // The Invincible Iron Man",
          "number": "80",
          "purchaseUrls": {}
        },
        {
          "uuid": "uuid-tony-363-url",
          "name": "Tony Stark // The Invincible Iron Man",
          "number": "363",
          "purchaseUrls": {
            "cardKingdom": "https://www.cardkingdom.com/mtg/marvel-super-heroes-variants/tony-stark-0363-borderless",
            "cardKingdomFoil": "https://www.cardkingdom.com/mtg/marvel-super-heroes-variants/tony-stark-0363-borderless-foil"
          }
        },
        {
          "uuid": "uuid-tony-363-price",
          "name": "Tony Stark // The Invincible Iron Man",
          "number": "363",
          "purchaseUrls": {}
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

func TestDecodeAllPrintingsSets_MergesSplitUUIDPrintings(t *testing.T) {
	prices := map[string]ckUUIDPrice{
		"uuid-tony-80-price":  {normal: 7.49, foil: 8.99},
		"uuid-tony-363-price": {foil: 44.99},
	}
	updatedAt := time.Date(2026, 6, 28, 0, 0, 0, 0, time.UTC)

	cheapest, err := decodeAllPrintingsSets(
		json.NewDecoder(strings.NewReader(sampleAllPrintingsSplitUUID)),
		prices,
		updatedAt,
	)
	require.NoError(t, err)
	require.Len(t, cheapest, 1)

	listing := cheapest["tony stark // the invincible iron man"]
	require.Equal(t, "Tony Stark // The Invincible Iron Man", listing.CardName)
	require.Equal(t, "Marvel Super Heroes", listing.Edition)
	require.InDelta(t, 7.49, listing.PriceUsd, 0.001)
	require.False(t, listing.IsFoil)
	require.Equal(t, "https://www.cardkingdom.com/mtg/marvel-super-heroes/tony-stark", listing.URL)
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
