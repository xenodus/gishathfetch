package fivemana

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/util"
	"mtg-price-checker-sg/pkg/config"
)

const (
	storefrontGraphQLPath = "/api/2024-10/graphql.json"
	// storefrontAccessToken is the public Storefront API token published in the
	// shop's shopify-features / checkout meta. It is intended for storefront use.
	storefrontAccessToken = "9e4cb078af6a814458ce898eb9631fe6"
	storefrontMTGType     = "MTG Single"
)

const storefrontSearchQuery = `query SearchCards($q: String!) {
  search(
    query: $q
    first: 25
    types: PRODUCT
    productFilters: [{ available: true }, { productType: "MTG Single" }]
  ) {
    edges {
      node {
        ... on Product {
          title
          handle
          availableForSale
          productType
          tags
          featuredImage { url }
          variants(first: 20) {
            edges {
              node {
                title
                availableForSale
                price { amount }
              }
            }
          }
        }
      }
    }
  }
}`

type graphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

type graphQLResponse struct {
	Data *struct {
		Search *struct {
			Edges []graphQLEdge `json:"edges"`
		} `json:"search"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

type graphQLEdge struct {
	Node *graphQLProduct `json:"node"`
}

type graphQLProduct struct {
	Title           string   `json:"title"`
	Handle          string   `json:"handle"`
	AvailableForSale bool    `json:"availableForSale"`
	ProductType     string   `json:"productType"`
	Tags            []string `json:"tags"`
	FeaturedImage   *struct {
		URL string `json:"url"`
	} `json:"featuredImage"`
	Variants struct {
		Edges []struct {
			Node *graphQLVariant `json:"node"`
		} `json:"edges"`
	} `json:"variants"`
}

type graphQLVariant struct {
	Title            string `json:"title"`
	AvailableForSale bool   `json:"availableForSale"`
	Price            struct {
		Amount string `json:"amount"`
	} `json:"price"`
}

func (s Store) searchGraphQL(ctx context.Context, searchStr string) ([]gateway.Card, error) {
	storeBase, err := s.storeBaseURL()
	if err != nil {
		return nil, err
	}

	apiURL := &url.URL{
		Scheme: storeBase.Scheme,
		Host:   storeBase.Host,
		Path:   storefrontGraphQLPath,
	}

	payload, err := json.Marshal(graphQLRequest{
		Query:     storefrontSearchQuery,
		Variables: map[string]any{"q": searchStr},
	})
	if err != nil {
		return nil, err
	}

	opts := fiveManaOutboundOpts(storeBase, storeBase, gateway.OutboundStyleJSON)

	resp, err := gateway.DoOutboundRoundTrip(ctx, opts, config.SearchAttemptTimeout, func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL.String(), bytes.NewReader(payload))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Shopify-Storefront-Access-Token", storefrontAccessToken)
		req.ContentLength = int64(len(payload))
		return req, nil
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := gateway.ReadResponseBody(resp)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("%s", gateway.FormatUnexpectedHTTPStatus(s.Name, resp, body))
	}

	var parsed graphQLResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, gateway.WrapJSONDecodeError(err, resp, body)
	}
	if len(parsed.Errors) > 0 {
		return nil, fmt.Errorf("%s graphql: %s", s.Name, parsed.Errors[0].Message)
	}
	if parsed.Data == nil || parsed.Data.Search == nil {
		return nil, fmt.Errorf("%s graphql: missing search data", s.Name)
	}

	return mapGraphQLProducts(s.Name, parsed.Data.Search.Edges), nil
}

func mapGraphQLProducts(storeName string, edges []graphQLEdge) []gateway.Card {
	var cards []gateway.Card
	for _, edge := range edges {
		if edge.Node == nil {
			continue
		}
		cards = append(cards, mapGraphQLProduct(storeName, edge.Node)...)
	}
	return cards
}

func mapGraphQLProduct(storeName string, product *graphQLProduct) []gateway.Card {
	if product == nil || !product.AvailableForSale {
		return nil
	}
	if product.ProductType != "" && product.ProductType != storefrontMTGType {
		return nil
	}

	name, titleFoil := parseNameAndFoil(product.Title)
	if name == "" || strings.TrimSpace(product.Handle) == "" {
		return nil
	}

	img := ""
	if product.FeaturedImage != nil {
		img = strings.TrimSpace(product.FeaturedImage.URL)
	}
	extra := extraInfoFromTitle(product.Title)
	if len(extra) == 0 {
		extra = extraInfoFromTags(product.Tags)
	}

	var cards []gateway.Card
	for _, edge := range product.Variants.Edges {
		variant := edge.Node
		if variant == nil || !variant.AvailableForSale {
			continue
		}
		price, err := strconv.ParseFloat(strings.TrimSpace(variant.Price.Amount), 64)
		if err != nil || price <= 0 {
			continue
		}

		quality := qualityFromVariantTitle(variant.Title)
		isFoil := titleFoil || variantIsFoil(variant.Title) || tagsIndicateFoil(product.Tags)

		cardURL, err := productURLWithUTM(StoreBaseURL + "/products/" + product.Handle)
		if err != nil {
			continue
		}

		cards = append(cards, gateway.Card{
			Name:      name,
			Url:       cardURL,
			Img:       img,
			Price:     price,
			InStock:   true,
			IsFoil:    isFoil,
			Source:    storeName,
			Quality:   quality,
			ExtraInfo: append([]string(nil), extra...),
		})
	}
	return cards
}

func parseNameAndFoil(raw string) (name string, isFoil bool) {
	name = strings.TrimSpace(raw)
	if name == "" {
		return "", false
	}
	lower := strings.ToLower(name)
	isFoil = strings.Contains(lower, "[foil]") ||
		strings.Contains(lower, "(foil)") ||
		strings.Contains(lower, "surge foil") ||
		strings.Contains(lower, "rainbow foil") ||
		strings.Contains(lower, "etched foil")
	name = strings.ReplaceAll(name, "[Foil]", "")
	name = strings.ReplaceAll(name, "[foil]", "")
	name = strings.Join(strings.Fields(name), " ")
	return name, isFoil
}

func variantIsFoil(variantTitle string) bool {
	return strings.Contains(strings.ToLower(variantTitle), "foil")
}

func tagsIndicateFoil(tags []string) bool {
	for _, tag := range tags {
		lower := strings.ToLower(tag)
		if strings.Contains(lower, "foil") && !strings.Contains(lower, "non-foil") {
			return true
		}
	}
	return false
}

func qualityFromVariantTitle(variantTitle string) string {
	quality := strings.TrimSpace(variantTitle)
	quality = strings.ReplaceAll(quality, "Foil", "")
	quality = strings.ReplaceAll(quality, "foil", "")
	quality = strings.Join(strings.Fields(quality), " ")
	return util.MapQuality(quality)
}

func extraInfoFromTitle(title string) []string {
	start := strings.LastIndex(title, "[")
	end := strings.LastIndex(title, "]")
	if start < 0 || end <= start {
		return nil
	}
	setName := strings.TrimSpace(title[start+1 : end])
	if setName == "" || strings.EqualFold(setName, "foil") {
		return nil
	}
	return []string{fmt.Sprintf("[%s]", setName)}
}

func extraInfoFromTags(tags []string) []string {
	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed == "" {
			continue
		}
		lower := strings.ToLower(trimmed)
		if lower == "red" || lower == "blue" || lower == "green" || lower == "white" ||
			lower == "black" || lower == "colorless" || lower == "multicolor" ||
			strings.Contains(lower, "foil") ||
			lower == "common" || lower == "uncommon" || lower == "rare" ||
			lower == "mythic" || lower == "mythic rare" {
			continue
		}
		return []string{fmt.Sprintf("[%s]", trimmed)}
	}
	return nil
}
