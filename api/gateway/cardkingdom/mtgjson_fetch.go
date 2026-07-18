package cardkingdom

import (
	"bytes"
	"compress/bzip2"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"mtg-price-checker-sg/pkg/config"
)

const (
	defaultAllPricesTodayURL = "https://mtgjson.com/api/v5/AllPricesToday.json.bz2"
	defaultAllPrintingsURL   = "https://mtgjson.com/api/v5/AllPrintings.json.bz2"
	// mtgjsonFetchTimeout bounds the full MTGJSON download + parse path. AllPrintings
	// alone can take several minutes on Lambda; keep Lambda timeout at 600s.
	mtgjsonFetchTimeout            = 8 * time.Minute
	mtgjsonAllPricesTodayHTTPTimeout = 2 * time.Minute
	mtgjsonAllPrintingsHTTPTimeout  = 7 * time.Minute
)

type ckUUIDPrice struct {
	normal float64
	foil   float64
}

type mtgjsonCard struct {
	UUID         string   `json:"uuid"`
	Name         string   `json:"name"`
	FaceName     string   `json:"faceName"`
	Number       string   `json:"number"`
	Side         string   `json:"side"`
	Finishes     []string `json:"finishes"`
	PurchaseUrls struct {
		CardKingdom     string `json:"cardKingdom"`
		CardKingdomFoil string `json:"cardKingdomFoil"`
	} `json:"purchaseUrls"`
}

type printingKey struct {
	number string
	name   string
	isDFC  bool
}

type printingAggregate struct {
	cardName        string
	cardKingdom     string
	cardKingdomFoil string
	offersNonfoil   bool
	price           ckUUIDPrice
}

func fetchCheapestFromMTGJSON(ctx context.Context) (map[string]Listing, error) {
	ctx, cancel := context.WithTimeout(ctx, mtgjsonFetchTimeout)
	defer cancel()

	log.Printf("ck price refresh: downloading AllPricesToday from %s", allPricesTodayURL())
	pricesStarted := time.Now()
	pricesRaw, err := downloadMTGJSONBzip2(ctx, allPricesTodayURL(), mtgjsonAllPricesTodayHTTPTimeout)
	if err != nil {
		return nil, fmt.Errorf("all prices today: %w", err)
	}
	log.Printf("ck price refresh: downloaded AllPricesToday bytes=%d duration=%s", len(pricesRaw), time.Since(pricesStarted).Round(time.Millisecond))

	pricesByUUID, updatedAt, err := parseCKPricesByUUID(pricesRaw)
	if err != nil {
		return nil, fmt.Errorf("parse all prices today: %w", err)
	}
	if len(pricesByUUID) == 0 {
		return nil, fmt.Errorf("parse all prices today: no card kingdom prices found")
	}
	log.Printf("ck price refresh: parsed AllPricesToday uuid_prices=%d price_date=%s", len(pricesByUUID), updatedAt.Format("2006-01-02"))

	log.Printf("ck price refresh: streaming AllPrintings from %s", allPrintingsURL())
	printingsStarted := time.Now()
	listings, err := streamAllPrintingsSets(ctx, allPrintingsURL(), pricesByUUID, updatedAt)
	if err != nil {
		return nil, err
	}
	log.Printf("ck price refresh: streamed AllPrintings listings=%d duration=%s", len(listings), time.Since(printingsStarted).Round(time.Millisecond))

	return listings, nil
}

func allPricesTodayURL() string {
	if url := strings.TrimSpace(os.Getenv(config.MTGJSONAllPricesTodayURLEnv)); url != "" {
		return url
	}
	return defaultAllPricesTodayURL
}

func allPrintingsURL() string {
	if url := strings.TrimSpace(os.Getenv(config.MTGJSONAllPrintingsURLEnv)); url != "" {
		return url
	}
	return defaultAllPrintingsURL
}

func downloadMTGJSONBzip2(ctx context.Context, downloadURL string, timeout time.Duration) ([]byte, error) {
	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/octet-stream")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	compressed, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	decompressed, err := io.ReadAll(bzip2.NewReader(bytes.NewReader(compressed)))
	if err != nil {
		return nil, fmt.Errorf("bzip2 decompress: %w", err)
	}
	return decompressed, nil
}

func parseCKPricesByUUID(raw []byte) (map[string]ckUUIDPrice, time.Time, error) {
	var payload struct {
		Meta struct {
			Date string `json:"date"`
		} `json:"meta"`
		Data map[string]struct {
			Paper *struct {
				CardKingdom *struct {
					Retail map[string]map[string]float64 `json:"retail"`
				} `json:"cardkingdom"`
			} `json:"paper"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, time.Time{}, err
	}

	updatedAt, err := time.Parse("2006-01-02", payload.Meta.Date)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("meta date: %w", err)
	}

	pricesByUUID := make(map[string]ckUUIDPrice)
	for uuid, entry := range payload.Data {
		if entry.Paper == nil || entry.Paper.CardKingdom == nil {
			continue
		}
		price := ckUUIDPrice{
			normal: latestRetailPrice(entry.Paper.CardKingdom.Retail["normal"]),
			foil:   latestRetailPrice(entry.Paper.CardKingdom.Retail["foil"]),
		}
		if price.normal <= 0 && price.foil <= 0 {
			continue
		}
		pricesByUUID[uuid] = price
	}

	return pricesByUUID, updatedAt.UTC(), nil
}

func latestRetailPrice(byDate map[string]float64) float64 {
	var latest float64
	for _, price := range byDate {
		if price > latest {
			latest = price
		}
	}
	return latest
}

func streamAllPrintingsSets(
	ctx context.Context,
	downloadURL string,
	pricesByUUID map[string]ckUUIDPrice,
	updatedAt time.Time,
) (map[string]Listing, error) {
	client := &http.Client{Timeout: mtgjsonAllPrintingsHTTPTimeout}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/octet-stream")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	decoder := json.NewDecoder(bzip2.NewReader(resp.Body))
	return decodeAllPrintingsSets(decoder, pricesByUUID, updatedAt)
}

func decodeAllPrintingsSets(
	decoder *json.Decoder,
	pricesByUUID map[string]ckUUIDPrice,
	updatedAt time.Time,
) (map[string]Listing, error) {
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	if delim, ok := token.(json.Delim); !ok || delim != '{' {
		return nil, fmt.Errorf("expected top-level object")
	}

	cheapest := make(map[string]Listing)
	for decoder.More() {
		keyToken, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		key, ok := keyToken.(string)
		if !ok {
			return nil, fmt.Errorf("expected object key")
		}

		switch key {
		case "meta":
			if err := skipJSONValue(decoder); err != nil {
				return nil, err
			}
		case "data":
			if err := decodeSetCatalog(decoder, pricesByUUID, updatedAt, cheapest); err != nil {
				return nil, err
			}
		default:
			if err := skipJSONValue(decoder); err != nil {
				return nil, err
			}
		}
	}

	if _, err := decoder.Token(); err != nil {
		return nil, err
	}
	return cheapest, nil
}

func decodeSetCatalog(
	decoder *json.Decoder,
	pricesByUUID map[string]ckUUIDPrice,
	updatedAt time.Time,
	cheapest map[string]Listing,
) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	if delim, ok := token.(json.Delim); !ok || delim != '{' {
		return fmt.Errorf("expected data object")
	}

	for decoder.More() {
		if _, err := decoder.Token(); err != nil {
			return err
		}

		var set struct {
			Name  string        `json:"name"`
			Cards []mtgjsonCard `json:"cards"`
		}
		if err := decoder.Decode(&set); err != nil {
			return err
		}
		if excludedCKPriceEdition(set.Name) {
			continue
		}

		aggregates := make(map[printingKey]printingAggregate)
		for _, card := range set.Cards {
			mergePrintingAggregate(aggregates, card, pricesByUUID)
		}
		for _, aggregate := range aggregates {
			applyPrintingAggregate(cheapest, aggregate.cardName, set.Name, aggregate, updatedAt)
		}
	}

	_, err = decoder.Token()
	return err
}

func mergePrintingAggregate(
	aggregates map[printingKey]printingAggregate,
	card mtgjsonCard,
	pricesByUUID map[string]ckUUIDPrice,
) {
	name := strings.TrimSpace(card.Name)
	if name == "" {
		return
	}

	key := printingKeyFor(card)
	aggregate := aggregates[key]
	aggregate.cardName = preferCardName(aggregate.cardName, name)
	if printingOffersNonfoil(card.Finishes) {
		aggregate.offersNonfoil = true
	}

	if card.PurchaseUrls.CardKingdom != "" {
		aggregate.cardKingdom = card.PurchaseUrls.CardKingdom
	}
	if card.PurchaseUrls.CardKingdomFoil != "" {
		aggregate.cardKingdomFoil = card.PurchaseUrls.CardKingdomFoil
	}
	if price, ok := pricesByUUID[card.UUID]; ok {
		if price.normal > 0 && (aggregate.price.normal <= 0 || price.normal < aggregate.price.normal) {
			aggregate.price.normal = price.normal
		}
		if price.foil > 0 && (aggregate.price.foil <= 0 || price.foil < aggregate.price.foil) {
			aggregate.price.foil = price.foil
		}
	}

	aggregates[key] = aggregate
}

func printingKeyFor(card mtgjsonCard) printingKey {
	number := strings.TrimSpace(card.Number)
	side := strings.TrimSpace(card.Side)
	if side == "a" || side == "b" {
		return printingKey{number: number, isDFC: true}
	}
	return printingKey{number: number, name: strings.TrimSpace(card.Name)}
}

func printingOffersNonfoil(finishes []string) bool {
	for _, finish := range finishes {
		if strings.EqualFold(strings.TrimSpace(finish), "nonfoil") {
			return true
		}
	}
	return false
}

func preferCardName(current string, candidate string) string {
	current = strings.TrimSpace(current)
	candidate = strings.TrimSpace(candidate)
	if candidate == "" {
		return current
	}
	if current == "" {
		return candidate
	}
	if strings.Contains(candidate, doubleFacedNameSeparator) {
		return candidate
	}
	if strings.Contains(current, doubleFacedNameSeparator) {
		return current
	}
	return candidate
}

func applyPrintingAggregate(
	cheapest map[string]Listing,
	cardName string,
	setName string,
	aggregate printingAggregate,
	updatedAt time.Time,
) {
	updatedAtValue := updatedAt.Format(time.RFC3339)
	if aggregate.price.normal > 0 && aggregate.cardKingdom != "" {
		considerCheapestListing(cheapest, Listing{
			CardName:  cardName,
			Edition:   setName,
			PriceUsd:  aggregate.price.normal,
			URL:       aggregate.cardKingdom,
			IsFoil:    false,
			UpdatedAt: updatedAtValue,
		})
	}

	if aggregate.price.foil <= 0 {
		return
	}
	// Foil-only printings (borderless/showcase variants) should not become the
	// cheapest listing for a card name when MTGJSON lacks the default printing.
	if !aggregate.offersNonfoil {
		return
	}

	foilURL := aggregate.cardKingdomFoil
	if foilURL == "" {
		foilURL = aggregate.cardKingdom
	}
	if foilURL == "" {
		return
	}

	considerCheapestListing(cheapest, Listing{
		CardName:  cardName,
		Edition:   setName,
		PriceUsd:  aggregate.price.foil,
		URL:       foilURL,
		IsFoil:    true,
		UpdatedAt: updatedAtValue,
	})
}

func skipJSONValue(decoder *json.Decoder) error {
	var discard json.RawMessage
	return decoder.Decode(&discard)
}
