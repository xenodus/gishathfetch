package cardscentral

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/util"
	"mtg-price-checker-sg/pkg/config"
)

// Cards Central runs its own platform (cardscentral.com), not BinderPOS or
// Shopify, so it can't be crawled like the other storefronts. It exposes a
// purpose-built JSON feed for aggregators instead: one row per in-stock raw
// single, MTG-filtered, cheapest-first. See
// https://cardscentral.com/api/lgs/search?q=<card name>
const StoreName = "Cards Central"
const StoreBaseURL = "https://cardscentral.com"
const StoreSearchAPI = "/api/lgs/search"

// item mirrors one element of the /api/lgs/search response array.
type item struct {
	Name      string  `json:"name"`
	Set       string  `json:"set"`
	Url       string  `json:"url"`
	Img       string  `json:"img"`
	Price     float64 `json:"price"`
	Currency  string  `json:"currency"`
	InStock   bool    `json:"inStock"`
	Finish    string  `json:"finish"`
	IsFoil    bool    `json:"isFoil"`
	Condition string  `json:"condition"`
}

type Store struct {
	Name      string
	BaseUrl   string
	SearchAPI string
}

func NewLGS() gateway.LGS {
	return Store{
		Name:      StoreName,
		BaseUrl:   StoreBaseURL,
		SearchAPI: StoreSearchAPI,
	}
}

func (s Store) Search(ctx context.Context, searchStr string) ([]gateway.Card, error) {
	var cards []gateway.Card

	apiURL := &url.URL{
		Scheme:   "https",
		Host:     strings.TrimPrefix(strings.TrimPrefix(s.BaseUrl, "https://"), "http://"),
		Path:     s.SearchAPI,
		RawQuery: url.Values{"q": {searchStr}}.Encode(),
	}

	resp, err := gateway.DoOutboundGET(
		ctx,
		apiURL.String(),
		gateway.OutboundRequestOptions{Style: gateway.OutboundStyleJSON},
		15*time.Second,
	)
	if err != nil {
		return cards, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return cards, fmt.Errorf("unexpected status for %s: %s", s.Name, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return cards, err
	}

	var items []item
	if err = json.Unmarshal(body, &items); err != nil {
		return cards, err
	}

	for _, it := range items {
		name := strings.TrimSpace(it.Name)
		if !it.InStock || it.Price <= 0 || name == "" {
			continue
		}

		// Stamp the aggregator's utm_source on the product link.
		productURL := it.Url
		if u, perr := url.Parse(it.Url); perr == nil {
			q := u.Query()
			q.Set("utm_source", config.UtmSource)
			u.RawQuery = q.Encode()
			productURL = u.String()
		}

		// ExtraInfo carries the set (foil rides IsFoil; condition rides Quality).
		// Matches how the other custom stores annotate a row.
		extra := []string{}
		if it.Set != "" {
			extra = append(extra, fmt.Sprintf("[%s]", it.Set))
		}

		cards = append(cards, gateway.Card{
			Name:      name,
			Url:       strings.TrimSpace(productURL),
			Img:       it.Img,
			Price:     it.Price,
			InStock:   true,
			IsFoil:    it.IsFoil,
			Source:    s.Name,
			Quality:   util.MapQuality(it.Condition),
			ExtraInfo: extra,
		})
	}

	return cards, nil
}
