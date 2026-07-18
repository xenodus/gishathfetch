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
          "side": "a",
          "purchaseUrls": {
            "cardKingdom": "https://www.cardkingdom.com/mtg/marvel-super-heroes/tony-stark",
            "cardKingdomFoil": "https://www.cardkingdom.com/mtg/marvel-super-heroes/tony-stark-foil"
          }
        },
        {
          "uuid": "uuid-tony-80-price",
          "name": "Tony Stark // The Invincible Iron Man",
          "number": "80",
          "side": "b",
          "purchaseUrls": {}
        },
        {
          "uuid": "uuid-tony-363-url",
          "name": "Tony Stark // The Invincible Iron Man",
          "number": "363",
          "side": "a",
          "purchaseUrls": {
            "cardKingdom": "https://www.cardkingdom.com/mtg/marvel-super-heroes-variants/tony-stark-0363-borderless",
            "cardKingdomFoil": "https://www.cardkingdom.com/mtg/marvel-super-heroes-variants/tony-stark-0363-borderless-foil"
          }
        },
        {
          "uuid": "uuid-tony-363-price",
          "name": "Tony Stark // The Invincible Iron Man",
          "number": "363",
          "side": "b",
          "purchaseUrls": {}
        }
      ]
    }
  }
}`

const sampleAllPrintingsJenniferWalters = `{
  "meta": {"date": "2026-07-02"},
  "data": {
    "MSH": {
      "name": "Marvel Super Heroes",
      "cards": [
        {
          "uuid": "uuid-jw-18-url",
          "name": "Jennifer Walters // The Sensational She-Hulk",
          "faceName": "Jennifer Walters",
          "number": "18",
          "side": "a",
          "purchaseUrls": {
            "cardKingdom": "https://www.cardkingdom.com/mtg/marvel-super-heroes/jennifer-walters"
          }
        },
        {
          "uuid": "uuid-jw-18-price",
          "name": "Jennifer Walters // The Sensational She-Hulk",
          "faceName": "The Sensational She-Hulk",
          "number": "18",
          "side": "b",
          "purchaseUrls": {}
        }
      ]
    }
  }
}`

const sampleAllPrintingsWorldChampionshipDecks = `{
  "meta": {"date": "2026-06-28"},
  "data": {
    "WC04": {
      "name": "World Championship Decks 2004",
      "cards": [
        {
          "uuid": "uuid-wcd-bolt",
          "name": "Lightning Bolt",
          "number": "1",
          "purchaseUrls": {
            "cardKingdom": "https://www.cardkingdom.com/mtg/world-championships/lightning-bolt"
          }
        }
      ]
    },
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

	listing := cheapest["lightning bolt"]
	require.Equal(t, "Lightning Bolt", listing.CardName)
	require.Equal(t, "Fourth Edition", listing.Edition)
	require.InDelta(t, 1.49, listing.PriceUsd, 0.001)
	require.False(t, listing.IsFoil)
	require.Equal(t, "https://www.cardkingdom.com/mtg/fourth-edition/lightning-bolt", listing.URL)
	require.Equal(t, updatedAt.Format(time.RFC3339), listing.UpdatedAt)
}

func TestDecodeAllPrintingsSets_ExcludesWorldChampionshipDecks(t *testing.T) {
	prices := map[string]ckUUIDPrice{
		"uuid-wcd-bolt": {normal: 0.99},
		"uuid-bolt-4ed": {normal: 1.49},
	}
	updatedAt := time.Date(2026, 6, 28, 0, 0, 0, 0, time.UTC)

	cheapest, err := decodeAllPrintingsSets(
		json.NewDecoder(strings.NewReader(sampleAllPrintingsWorldChampionshipDecks)),
		prices,
		updatedAt,
	)
	require.NoError(t, err)

	listing := cheapest["lightning bolt"]
	require.Equal(t, "Fourth Edition", listing.Edition)
	require.InDelta(t, 1.49, listing.PriceUsd, 0.001)
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

	listing := cheapest["tony stark // the invincible iron man"]
	require.Equal(t, "Tony Stark // The Invincible Iron Man", listing.CardName)
	require.Equal(t, "Marvel Super Heroes", listing.Edition)
	require.InDelta(t, 7.49, listing.PriceUsd, 0.001)
	require.False(t, listing.IsFoil)
	require.Equal(t, "https://www.cardkingdom.com/mtg/marvel-super-heroes/tony-stark", listing.URL)

	faceListing := cheapest["tony stark"]
	require.InDelta(t, 7.49, faceListing.PriceUsd, 0.001)
}

func TestDecodeAllPrintingsSets_MergesDoubleFacedFacesByNumber(t *testing.T) {
	prices := map[string]ckUUIDPrice{
		"uuid-jw-18-price": {normal: 10.99},
	}
	updatedAt := time.Date(2026, 7, 2, 0, 0, 0, 0, time.UTC)

	cheapest, err := decodeAllPrintingsSets(
		json.NewDecoder(strings.NewReader(sampleAllPrintingsJenniferWalters)),
		prices,
		updatedAt,
	)
	require.NoError(t, err)

	listing := cheapest["jennifer walters // the sensational she-hulk"]
	require.Equal(t, "Jennifer Walters // The Sensational She-Hulk", listing.CardName)
	require.InDelta(t, 10.99, listing.PriceUsd, 0.001)
	require.Equal(t, "https://www.cardkingdom.com/mtg/marvel-super-heroes/jennifer-walters", listing.URL)

	faceListing := cheapest["jennifer walters"]
	require.InDelta(t, 10.99, faceListing.PriceUsd, 0.001)
}

func TestDecodeAllPrintingsSets_MergesDoubleFacedFacesUsingCheapestSidePrice(t *testing.T) {
	prices := map[string]ckUUIDPrice{
		"uuid-jw-18-url":   {normal: 14.99},
		"uuid-jw-18-price": {normal: 10.99},
	}
	updatedAt := time.Date(2026, 7, 3, 0, 0, 0, 0, time.UTC)

	const sample = `{
  "meta": {"date": "2026-07-03"},
  "data": {
    "MSH": {
      "name": "Marvel Super Heroes",
      "cards": [
        {
          "uuid": "uuid-jw-18-url",
          "name": "Jennifer Walters // The Sensational She-Hulk",
          "faceName": "Jennifer Walters",
          "number": "18",
          "side": "a",
          "purchaseUrls": {
            "cardKingdom": "https://www.cardkingdom.com/mtg/marvel-super-heroes/jennifer-walters"
          }
        },
        {
          "uuid": "uuid-jw-18-price",
          "name": "Jennifer Walters // The Sensational She-Hulk",
          "faceName": "The Sensational She-Hulk",
          "number": "18",
          "side": "b",
          "purchaseUrls": {}
        }
      ]
    }
  }
}`

	cheapest, err := decodeAllPrintingsSets(
		json.NewDecoder(strings.NewReader(sample)),
		prices,
		updatedAt,
	)
	require.NoError(t, err)

	listing := cheapest["jennifer walters // the sensational she-hulk"]
	require.InDelta(t, 10.99, listing.PriceUsd, 0.001)
	require.False(t, listing.IsFoil)
}

func TestDecodeAllPrintingsSets_SkipsFoilOnlyVariantPrintings(t *testing.T) {
	const sample = `{
  "meta": {"date": "2026-07-18"},
  "data": {
    "TMC": {
      "name": "Teenage Mutant Ninja Turtles Eternal",
      "cards": [
        {
          "uuid": "uuid-krang-13",
          "name": "Krang, the All-Powerful",
          "number": "13",
          "finishes": ["foil", "nonfoil"],
          "purchaseUrls": {
            "cardKingdom": "https://mtgjson.com/links/regular",
            "cardKingdomFoil": "https://mtgjson.com/links/regular-foil"
          }
        },
        {
          "uuid": "uuid-krang-86",
          "name": "Krang, the All-Powerful",
          "number": "86",
          "finishes": ["foil"],
          "purchaseUrls": {
            "cardKingdomFoil": "https://mtgjson.com/links/borderless-foil"
          }
        }
      ]
    }
  }
}`

	prices := map[string]ckUUIDPrice{
		"uuid-krang-86": {foil: 54.99},
	}
	updatedAt := time.Date(2026, 7, 18, 0, 0, 0, 0, time.UTC)

	cheapest, err := decodeAllPrintingsSets(
		json.NewDecoder(strings.NewReader(sample)),
		prices,
		updatedAt,
	)
	require.NoError(t, err)
	_, hasListing := cheapest["krang, the all-powerful"]
	require.False(t, hasListing)
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

func TestConsiderCheapestListing_FoilDoubleFacedDoesNotPolluteFaceAlias(t *testing.T) {
	updatedAt := time.Date(2026, 7, 2, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
	listings := make(map[string]Listing)
	considerCheapestListing(listings, Listing{
		CardName:  "Jennifer Walters // The Sensational She-Hulk",
		Edition:   "Marvel Super Heroes",
		PriceUsd:  69.99,
		URL:       "https://mtgjson.com/links/d8187fd1eef32412",
		IsFoil:    true,
		UpdatedAt: updatedAt,
	})

	require.InDelta(t, 69.99, listings["jennifer walters // the sensational she-hulk"].PriceUsd, 0.001)
	_, hasFrontFace := listings["jennifer walters"]
	require.False(t, hasFrontFace)
	_, hasBackFace := listings["the sensational she-hulk"]
	require.False(t, hasBackFace)
}
