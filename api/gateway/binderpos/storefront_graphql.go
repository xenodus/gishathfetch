package binderpos

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

const storefrontGraphQLPath = "/api/2024-10/graphql.json"

const storefrontGraphQLQuery = `query SearchCards($q: String!) {
  search(
    query: $q
    first: 25
    types: PRODUCT
    productFilters: [{ available: true }]
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
                id
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
	Title            string   `json:"title"`
	Handle           string   `json:"handle"`
	AvailableForSale bool     `json:"availableForSale"`
	ProductType      string   `json:"productType"`
	Tags             []string `json:"tags"`
	FeaturedImage    *struct {
		URL string `json:"url"`
	} `json:"featuredImage"`
	Variants struct {
		Edges []struct {
			Node *graphQLVariant `json:"node"`
		} `json:"edges"`
	} `json:"variants"`
}

type graphQLVariant struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	AvailableForSale bool   `json:"availableForSale"`
	Price            struct {
		Amount string `json:"amount"`
	} `json:"price"`
}

func searchByStorefrontGraphQLDedicated(ctx context.Context, scrapVariant int, storeName, baseURL, accessToken, searchStr string) ([]gateway.Card, error) {
	proxyURLs := util.GetDedicatedProxyURLs()
	if len(proxyURLs) == 0 {
		return nil, fmt.Errorf("no dedicated proxy configured for binderpos storefront graphql")
	}

	var proxyURL string
	if pinned, ok := gateway.RequestDedicatedProxyURL(ctx); ok {
		proxyURL = pinned
	} else if config.UseLeasedDedicatedProxy {
		leasedURL, release, err := gateway.LeaseDedicatedProxyURL(ctx, proxyURLs)
		if err != nil {
			return nil, fmt.Errorf("dedicated proxy lease for binderpos storefront graphql: %w", err)
		}
		defer release()
		proxyURL = leasedURL
	} else {
		proxyURL = nextBinderposStorefrontProxyURL(proxyURLs)
	}

	client, err := newHTTPClientWithProxyURL(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("invalid dedicated proxy configured for binderpos storefront graphql: %w", err)
	}
	return searchByStorefrontGraphQL(ctx, client, scrapVariant, storeName, baseURL, accessToken, searchStr)
}

func searchByStorefrontGraphQLDirect(ctx context.Context, scrapVariant int, storeName, baseURL, accessToken, searchStr string) ([]gateway.Card, error) {
	profile := gateway.PickBrowserProfile()
	if !gateway.ShouldUseBrowserTLSEmulationForScraping() {
		profile = gateway.BrowserEmulationProfile{}
	}
	client, err := gateway.NewBinderposHTTPClient("", profile)
	if err != nil {
		return nil, fmt.Errorf("binderpos graphql direct client: %w", err)
	}
	return searchByStorefrontGraphQL(ctx, client, scrapVariant, storeName, baseURL, accessToken, searchStr)
}

func searchByStorefrontGraphQL(
	ctx context.Context,
	client *http.Client,
	scrapVariant int,
	storeName, baseURL, accessToken, searchStr string,
) ([]gateway.Card, error) {
	accessToken = strings.TrimSpace(accessToken)
	if accessToken == "" {
		return nil, fmt.Errorf("missing storefront access token for binderpos graphql")
	}

	storeBase, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return nil, err
	}
	apiURL := &url.URL{
		Scheme: storeBase.Scheme,
		Host:   storeBase.Host,
		Path:   storefrontGraphQLPath,
	}

	payload, err := json.Marshal(graphQLRequest{
		Query:     storefrontGraphQLQuery,
		Variables: map[string]any{"q": strings.TrimSpace(searchStr)},
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL.String(), bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Shopify-Storefront-Access-Token", accessToken)
	req.ContentLength = int64(len(payload))
	if err := gateway.PrepareOutboundRequest(ctx, req, gateway.OutboundRequestOptions{
		Style:              gateway.OutboundStyleJSON,
		StoreBase:          storeBase,
		ShopifySGDCurrency: true,
	}); err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := gateway.ReadResponseBody(resp)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("%s", gateway.FormatUnexpectedHTTPStatus(storeName, resp, body))
	}

	var parsed graphQLResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, gateway.WrapJSONDecodeError(err, resp, body)
	}
	if len(parsed.Errors) > 0 {
		return nil, fmt.Errorf("%s graphql: %s", storeName, parsed.Errors[0].Message)
	}
	if parsed.Data == nil || parsed.Data.Search == nil {
		return nil, fmt.Errorf("%s graphql: missing search data", storeName)
	}

	return mapGraphQLProducts(scrapVariant, storeName, baseURL, parsed.Data.Search.Edges), nil
}

func mapGraphQLProducts(scrapVariant int, storeName, baseURL string, edges []graphQLEdge) []gateway.Card {
	var cards []gateway.Card
	for _, edge := range edges {
		if edge.Node == nil {
			continue
		}
		cards = append(cards, mapGraphQLProduct(scrapVariant, storeName, baseURL, edge.Node)...)
	}
	return cards
}

func mapGraphQLProduct(scrapVariant int, storeName, baseURL string, product *graphQLProduct) []gateway.Card {
	if product == nil || !product.AvailableForSale {
		return nil
	}
	if !isMagicProductType(product.ProductType, product.Tags) {
		return nil
	}

	handle := strings.TrimSpace(product.Handle)
	title := strings.TrimSpace(product.Title)
	if handle == "" || title == "" {
		return nil
	}

	img := ""
	if product.FeaturedImage != nil {
		img = buildCardImageURL(product.FeaturedImage.URL, title)
	} else {
		img = buildCardImageURL("", title)
	}

	setName := extractSetName(title)
	productPath := "/products/" + handle

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
		variantID, ok := shopifyVariantID(variant.ID)
		if !ok {
			continue
		}
		cardURL, err := buildProductURLWithVariant(baseURL, productPath, variantID)
		if err != nil {
			continue
		}

		quality := strings.TrimSpace(strings.ReplaceAll(variant.Title, "Foil", ""))
		quality = strings.Join(strings.Fields(quality), " ")
		card := gateway.Card{
			Name:    formatCardName(scrapVariant, title, variant.Title),
			Url:     cardURL,
			Img:     img,
			Price:   price,
			InStock: true,
			IsFoil:  strings.Contains(strings.ToLower(variant.Title), "foil") || titleIndicatesFoil(title),
			Source:  storeName,
			Quality: util.MapQuality(quality),
		}
		if scrapVariant == 3 && setName != "" {
			card.ExtraInfo = []string{setName}
		}
		cards = append(cards, card)
	}
	return cards
}

func titleIndicatesFoil(title string) bool {
	lower := strings.ToLower(title)
	return strings.Contains(lower, "[foil]") ||
		strings.Contains(lower, "(foil)") ||
		strings.Contains(lower, "surge foil") ||
		strings.Contains(lower, "rainbow foil") ||
		strings.Contains(lower, "etched foil")
}

func shopifyVariantID(gid string) (int64, bool) {
	gid = strings.TrimSpace(gid)
	if gid == "" {
		return 0, false
	}
	if id, err := strconv.ParseInt(gid, 10, 64); err == nil && id > 0 {
		return id, true
	}
	const prefix = "gid://shopify/ProductVariant/"
	if !strings.HasPrefix(gid, prefix) {
		return 0, false
	}
	id, err := strconv.ParseInt(strings.TrimPrefix(gid, prefix), 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}
