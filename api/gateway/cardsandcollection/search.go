package cardsandcollection

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/pkg/config"
)

const StoreName = "Cards & Collections"
const StoreBaseURL = "https://cardsandcollections.com"
const StoreApiURL = "/api/catalog/"
const StoreSearchURL = "/?q="

type Store struct {
	Name      string
	BaseUrl   string
	SearchUrl string
}

func NewLGS() gateway.LGS {
	return Store{
		Name:      StoreName,
		BaseUrl:   StoreBaseURL,
		SearchUrl: StoreSearchURL,
	}
}

func (s Store) Search(ctx context.Context, searchStr string) ([]gateway.Card, error) {
	var (
		res   catalogSearchResponse
		cards []gateway.Card
	)

	apiURL := s.BaseUrl + StoreApiURL
	requestBody := fmt.Appendf(nil, `{"query":{"bool":{"should":[{"simple_query_string":{"query":"%s","fields":["name","setCode","setName"],"default_operator":"AND"}},{"multi_match":{"query":"%s","type":"phrase_prefix","fields":["name","setCode","setName"]}}]}},"post_filter":{"bool":{"must":{"terms":{"collectableContext.raw":["MTG","ACCESSORY"]}}}},"aggs":{"productCategory4":{"filter":{"bool":{"must":{"terms":{"collectableContext.raw":["MTG","ACCESSORY"]}}}},"aggs":{"productCategory.raw":{"terms":{"field":"productCategory.raw","size":50}},"productCategory.raw_count":{"cardinality":{"field":"productCategory.raw"}}}}},"size":20,"sort":[{"quantityOnSale":"desc"}]}`, searchStr, searchStr)
	resp, err := gateway.DoOutboundRoundTrip(ctx, gateway.OutboundRequestOptions{}, config.DedicatedSearchAttemptTimeout, func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewBuffer(requestBody))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		return req, nil
	})
	if err != nil {
		return cards, err
	}
	defer resp.Body.Close()

	body, err := gateway.ReadResponseBody(resp)
	if err != nil {
		return cards, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return cards, fmt.Errorf("%s", gateway.FormatUnexpectedHTTPStatus(s.Name, resp, body))
	}

	if err = json.Unmarshal(body, &res); err != nil {
		return cards, gateway.WrapJSONDecodeError(err, resp, body)
	}

	if len(res.Hits.Hits) > 0 {
		for _, hit := range res.Hits.Hits {
			if quantityOnSaleStr, quantityOk := hit.Source.QuantityOnSale.(string); quantityOk {
				if minPriceStr, priceOk := hit.Source.MinPriceSale.(string); priceOk {
					quantity, _ := strconv.Atoi(quantityOnSaleStr)
					minPrice, _ := strconv.ParseFloat(minPriceStr, 64)

					if quantity > 0 && minPrice > 0 {
						// url
						u := fmt.Sprintf(StoreBaseURL+"/product/%s", hit.ID)
						cleanPageURL, err := url.Parse(u)
						if err != nil {
							slog.Warn("error parsing url", "store", s.Name, "value", u, "err", err)
							continue
						}
						cleanPageURL.RawQuery = url.Values{
							"utm_source": []string{config.UtmSource},
						}.Encode()

						cards = append(cards, gateway.Card{
							Name:      strings.TrimSpace(hit.Source.Name),
							Url:       strings.TrimSpace(cleanPageURL.String()),
							InStock:   true,
							Price:     minPrice,
							Source:    s.Name,
							Img:       hit.Source.Img,
							ExtraInfo: []string{fmt.Sprintf("[%s]", hit.Source.SetName)},
						})
					}
				}
			}
		}
	}

	return cards, nil
}
