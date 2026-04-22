package binderpos

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/util"
	"mtg-price-checker-sg/pkg/config"
)

const (
	storefrontSuggestPath   = "/search/suggest.json"
	binderposDecklistAPIURL = "https://portal.binderpos.com/external/shopify/decklist"
	binderposDecklistType   = "mtg"
	binderposAttemptTimeout = 4 * time.Second
)

var binderposShopifyDomainByStoreHost = map[string]string{
	"cardscitadel.com":        "card-citadel.myshopify.com",
	"card-affinity.com":       "563304-2.myshopify.com",
	"cardboardcrackgames.com": "cardboardcrackgames.myshopify.com",
	"flagshipgames.sg":        "flagship-games.myshopify.com",
	"gameshaventcg.com":       "games-haven-sg.myshopify.com",
	"greyogregames.com":       "grey-ogre-games-singapore.myshopify.com",
	"hideoutcg.com":           "220022-20.myshopify.com",
	"sg-manapro.com":          "mana-pro-sg.myshopify.com",
	"mtg-asia.com":            "mtgasia.myshopify.com",
	"onemtg.com.sg":           "one-mtg.myshopify.com",
	"tefudagames.com":         "bacc1b-3.myshopify.com",
	// arcanesanctumtcg.com intentionally omitted: BinderPOS decklist API returns 401.
}

type storefrontSuggestResponse struct {
	Resources struct {
		Results struct {
			Products []storefrontProduct `json:"products"`
		} `json:"results"`
	} `json:"resources"`
}

type storefrontProduct struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Image string `json:"image"`
}

type storefrontProductDetail struct {
	Title    string                   `json:"title"`
	Variants []storefrontProductStock `json:"variants"`
}

type storefrontProductStock struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Available bool   `json:"available"`
	Price     int    `json:"price"`
}

type storefrontDecklistRequestItem struct {
	Card     string `json:"card"`
	Quantity int    `json:"quantity"`
}

type storefrontDecklistLine struct {
	Products  []storefrontDecklistProduct `json:"products"`
	ValidName string                      `json:"validName"`
}

type storefrontDecklistProduct struct {
	Title    string                    `json:"title"`
	Name     string                    `json:"name"`
	Handle   string                    `json:"handle"`
	SetName  string                    `json:"setName"`
	Image    string                    `json:"img"`
	Variants []storefrontDecklistStock `json:"variants"`
}

type storefrontDecklistStock struct {
	ShopifyID int64   `json:"shopifyId"`
	Title     string  `json:"title"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
}

func (i impl) Search(ctx context.Context, scrapVariant int, storeName, baseURL, searchURL, searchStr string) ([]gateway.Card, error) {
	return searchWithFallback(
		func() ([]gateway.Card, error) {
			return runWithAttemptTimeout(ctx, func(attemptCtx context.Context) ([]gateway.Card, error) {
				return searchByStorefrontAPI(attemptCtx, scrapVariant, storeName, baseURL, searchStr)
			})
		},
		func() ([]gateway.Card, error) {
			return runWithAttemptTimeout(ctx, func(attemptCtx context.Context) ([]gateway.Card, error) {
				return i.scrapDedicatedProxy(attemptCtx, scrapVariant, storeName, baseURL, searchURL, searchStr)
			})
		},
		func() ([]gateway.Card, error) {
			return runWithAttemptTimeout(ctx, func(attemptCtx context.Context) ([]gateway.Card, error) {
				return searchByStorefrontAPISharedProxy(attemptCtx, scrapVariant, storeName, baseURL, searchStr)
			})
		},
		func() ([]gateway.Card, error) {
			return runWithAttemptTimeout(ctx, func(attemptCtx context.Context) ([]gateway.Card, error) {
				return i.scrapDirect(attemptCtx, scrapVariant, storeName, baseURL, searchURL, searchStr)
			})
		},
	)
}

func runWithAttemptTimeout(ctx context.Context, fn func(context.Context) ([]gateway.Card, error)) ([]gateway.Card, error) {
	attemptCtx, cancel := context.WithTimeout(ctx, binderposAttemptTimeout)
	defer cancel()
	return fn(attemptCtx)
}

func searchByStorefrontAPI(ctx context.Context, scrapVariant int, storeName, baseURL, searchStr string) ([]gateway.Card, error) {
	client, ok := newDedicatedProxyHTTPClient()
	if !ok {
		return nil, fmt.Errorf("no dedicated proxy configured for binderpos storefront api")
	}

	return searchByStorefrontAPIWithClient(ctx, client, scrapVariant, storeName, baseURL, searchStr)
}

func searchByStorefrontAPIDirect(ctx context.Context, scrapVariant int, storeName, baseURL, searchStr string) ([]gateway.Card, error) {
	client := &http.Client{Timeout: binderposAttemptTimeout}
	return searchByStorefrontAPIWithClient(ctx, client, scrapVariant, storeName, baseURL, searchStr)
}

func searchByStorefrontAPISharedProxy(ctx context.Context, scrapVariant int, storeName, baseURL, searchStr string) ([]gateway.Card, error) {
	sharedProxyURL := strings.TrimSpace(os.Getenv("PROXY_URL"))
	if sharedProxyURL == "" {
		return nil, fmt.Errorf("no shared proxy configured for binderpos storefront api")
	}

	client, err := newHTTPClientWithProxyURL(sharedProxyURL)
	if err != nil {
		return nil, fmt.Errorf("invalid shared proxy configured for binderpos storefront api: %w", err)
	}

	return searchByStorefrontAPIWithClient(ctx, client, scrapVariant, storeName, baseURL, searchStr)
}

func searchWithFallback(
	searchAPIDedicatedFn func() ([]gateway.Card, error),
	scrapDedicatedFn func() ([]gateway.Card, error),
	searchAPISharedFn func() ([]gateway.Card, error),
	scrapDirectFn func() ([]gateway.Card, error),
) ([]gateway.Card, error) {
	// Some stores legitimately return no matches for the query.
	// If any attempt completes without an error, do not surface earlier failures.
	hasSuccessfulAttempt := false

	apiDedicatedCards, apiDedicatedErr := searchAPIDedicatedFn()
	apiDedicatedErr = annotateAttemptError(1, "api-dedicated", apiDedicatedErr)
	if apiDedicatedErr == nil {
		hasSuccessfulAttempt = true
	}
	if len(apiDedicatedCards) > 0 && apiDedicatedErr == nil {
		return apiDedicatedCards, nil
	}

	scrapedCards, scrapErr := scrapDedicatedFn()
	scrapErr = annotateAttemptError(2, "scrap-dedicated", scrapErr)
	if scrapErr == nil {
		hasSuccessfulAttempt = true
	}
	if len(scrapedCards) > 0 && scrapErr == nil {
		return scrapedCards, nil
	}

	apiSharedCards, apiSharedErr := searchAPISharedFn()
	apiSharedErr = annotateAttemptError(3, "api-shared", apiSharedErr)
	if apiSharedErr == nil {
		hasSuccessfulAttempt = true
	}
	if len(apiSharedCards) > 0 && apiSharedErr == nil {
		return apiSharedCards, nil
	}

	scrapedDirectCards, scrapDirectErr := scrapDirectFn()
	scrapDirectErr = annotateAttemptError(4, "scrap-direct", scrapDirectErr)
	if scrapDirectErr == nil {
		hasSuccessfulAttempt = true
	}
	if len(scrapedDirectCards) > 0 && scrapDirectErr == nil {
		return scrapedDirectCards, nil
	}

	if hasSuccessfulAttempt {
		return []gateway.Card{}, nil
	}

	if scrapDirectErr != nil {
		return scrapedDirectCards, scrapDirectErr
	}
	if apiSharedErr != nil {
		return apiSharedCards, apiSharedErr
	}
	if scrapErr != nil {
		return scrapedCards, scrapErr
	}
	if apiDedicatedErr != nil {
		return apiDedicatedCards, apiDedicatedErr
	}

	return []gateway.Card{}, nil
}

func annotateAttemptError(attempt int, strategy string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("attempt %d (%s): %w", attempt, strategy, err)
}

func searchByStorefrontAPIWithClient(ctx context.Context, client *http.Client, scrapVariant int, storeName, baseURL, searchStr string) ([]gateway.Card, error) {
	decklistCards, err := searchByBinderposDecklistAPI(ctx, client, scrapVariant, storeName, baseURL, searchStr)
	if err == nil {
		return decklistCards, nil
	}

	products, err := fetchSuggestProducts(ctx, client, baseURL, searchStr)
	if err != nil {
		return nil, err
	}
	if len(products) == 0 {
		return []gateway.Card{}, nil
	}

	cards := make([]gateway.Card, 0, len(products)*4)
	seen := map[string]struct{}{}
	for _, product := range products {
		productURL := product.URL
		if productURL == "" {
			continue
		}

		detail, err := fetchProductDetail(ctx, client, baseURL, productURL)
		if err != nil {
			continue
		}

		for _, variant := range detail.Variants {
			if variant.Price <= 0 {
				continue
			}

			quality := strings.TrimSpace(strings.ReplaceAll(variant.Title, "Foil", ""))
			cardURL, err := buildProductURLWithVariant(baseURL, productURL, variant.ID)
			if err != nil {
				continue
			}

			setName := extractSetName(detail.Title)
			image := buildCardImageURL(product.Image, detail.Title)
			card := gateway.Card{
				Name:    formatCardName(scrapVariant, detail.Title, variant.Title),
				Url:     cardURL,
				Img:     image,
				Price:   float64(variant.Price) / 100,
				InStock: variant.Available,
				IsFoil:  strings.Contains(strings.ToLower(variant.Title), "foil"),
				Source:  storeName,
				Quality: util.MapQuality(quality),
			}
			if scrapVariant == 3 && setName != "" {
				card.ExtraInfo = []string{setName}
			}

			key := fmt.Sprintf("%s|%.2f|%t", card.Url, card.Price, card.InStock)
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			cards = append(cards, card)
		}
	}

	return cards, nil
}

func searchByBinderposDecklistAPI(ctx context.Context, client *http.Client, scrapVariant int, storeName, baseURL, searchStr string) ([]gateway.Card, error) {
	shopifyDomain, ok := storefrontShopifyDomainForBaseURL(baseURL)
	if !ok {
		return nil, fmt.Errorf("missing shopify domain mapping for binderpos storefront")
	}

	decklistURL, err := url.Parse(binderposDecklistAPIURL)
	if err != nil {
		return nil, err
	}
	query := decklistURL.Query()
	query.Set("storeUrl", shopifyDomain)
	query.Set("type", binderposDecklistType)
	decklistURL.RawQuery = query.Encode()

	payload := []storefrontDecklistRequestItem{
		{
			Card:     strings.TrimSpace(searchStr),
			Quantity: 1,
		},
	}
	payloadBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, decklistURL.String(), bytes.NewReader(payloadBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", gateway.RandomBrowserUserAgent())
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	if err := gateway.WaitForDomainRequestSlot(ctx, req.URL); err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("binderpos decklist request failed status=%d body=%s", res.StatusCode, strings.TrimSpace(string(body)))
	}

	var lines []storefrontDecklistLine
	if err := json.NewDecoder(res.Body).Decode(&lines); err != nil {
		return nil, err
	}

	return mapDecklistLinesToCards(scrapVariant, storeName, baseURL, lines), nil
}

func storefrontShopifyDomainForBaseURL(baseURL string) (string, bool) {
	parsedURL, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return "", false
	}

	host := strings.ToLower(strings.TrimSpace(parsedURL.Hostname()))
	host = strings.TrimPrefix(host, "www.")
	if host == "" {
		return "", false
	}

	shopifyDomain, ok := binderposShopifyDomainByStoreHost[host]
	return shopifyDomain, ok
}

func mapDecklistLinesToCards(scrapVariant int, storeName, baseURL string, lines []storefrontDecklistLine) []gateway.Card {
	cards := make([]gateway.Card, 0, len(lines)*4)
	seen := map[string]struct{}{}

	for _, line := range lines {
		for _, product := range line.Products {
			productTitle := strings.TrimSpace(product.Title)
			if productTitle == "" {
				productTitle = strings.TrimSpace(product.Name)
			}
			if productTitle == "" {
				productTitle = strings.TrimSpace(line.ValidName)
			}
			if productTitle == "" {
				continue
			}

			productPath := ""
			if strings.TrimSpace(product.Handle) != "" {
				productPath = "/products/" + strings.TrimSpace(product.Handle)
			}

			setName := strings.TrimSpace(product.SetName)
			if setName == "" {
				setName = extractSetName(productTitle)
			}

			image := buildCardImageURL(product.Image, productTitle)
			for _, variant := range product.Variants {
				if variant.Price <= 0 {
					continue
				}
				if variant.ShopifyID <= 0 || productPath == "" {
					continue
				}

				cardURL, err := buildProductURLWithVariant(baseURL, productPath, variant.ShopifyID)
				if err != nil {
					continue
				}

				quality := strings.TrimSpace(strings.ReplaceAll(variant.Title, "Foil", ""))
				card := gateway.Card{
					Name:    formatCardName(scrapVariant, productTitle, variant.Title),
					Url:     cardURL,
					Img:     image,
					Price:   variant.Price,
					InStock: variant.Quantity > 0,
					IsFoil:  strings.Contains(strings.ToLower(variant.Title), "foil"),
					Source:  storeName,
					Quality: util.MapQuality(quality),
				}
				if scrapVariant == 3 && setName != "" {
					card.ExtraInfo = []string{setName}
				}

				key := fmt.Sprintf("%s|%.2f|%t", card.Url, card.Price, card.InStock)
				if _, exists := seen[key]; exists {
					continue
				}
				seen[key] = struct{}{}
				cards = append(cards, card)
			}
		}
	}

	return cards
}

func newDedicatedProxyHTTPClient() (*http.Client, bool) {
	proxyURLs := util.GetDedicatedProxyURLs()
	if len(proxyURLs) == 0 {
		return nil, false
	}

	proxyURL := strings.TrimSpace(proxyURLs[rand.IntN(len(proxyURLs))])
	if proxyURL == "" {
		return nil, false
	}

	client, err := newHTTPClientWithProxyURL(proxyURL)
	if err != nil {
		return nil, false
	}

	return client, true
}

func newHTTPClientWithProxyURL(proxyURL string) (*http.Client, error) {
	parsedProxyURL, err := url.Parse(proxyURL)
	if err != nil {
		return nil, err
	}

	return &http.Client{
		Timeout: binderposAttemptTimeout,
		Transport: &http.Transport{
			Proxy: http.ProxyURL(parsedProxyURL),
		},
	}, nil
}

func fetchSuggestProducts(ctx context.Context, client *http.Client, baseURL, searchStr string) ([]storefrontProduct, error) {
	suggestURL, err := url.Parse(baseURL + storefrontSuggestPath)
	if err != nil {
		return nil, err
	}

	query := suggestURL.Query()
	query.Set("q", strings.TrimSpace(searchStr+" mtg"))
	query.Set("resources[type]", "product")
	query.Set("resources[limit]", "8")
	query.Set("resources[options][unavailable_products]", "hide")
	suggestURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, suggestURL.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", gateway.RandomBrowserUserAgent())
	if err := gateway.WaitForDomainRequestSlot(ctx, req.URL); err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("storefront suggest request returned status %d", res.StatusCode)
	}

	var payload storefrontSuggestResponse
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return nil, err
	}

	return payload.Resources.Results.Products, nil
}

func fetchProductDetail(ctx context.Context, client *http.Client, baseURL, productPath string) (*storefrontProductDetail, error) {
	productJSONURL, err := productPathToJSONURL(baseURL, productPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, productJSONURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", gateway.RandomBrowserUserAgent())
	if err := gateway.WaitForDomainRequestSlot(ctx, req.URL); err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("product detail request failed status=%d body=%s", res.StatusCode, strings.TrimSpace(string(body)))
	}

	var detail storefrontProductDetail
	if err := json.NewDecoder(res.Body).Decode(&detail); err != nil {
		return nil, err
	}

	return &detail, nil
}

func productPathToJSONURL(baseURL, productPath string) (string, error) {
	u, err := url.Parse(productPath)
	if err != nil {
		return "", err
	}

	path := strings.TrimSpace(u.Path)
	if path == "" {
		return "", fmt.Errorf("missing product path")
	}

	path = strings.TrimSuffix(path, "/")
	if !strings.HasSuffix(path, ".js") {
		path += ".js"
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	base.Path = path
	base.RawQuery = ""

	return base.String(), nil
}

func buildProductURLWithVariant(baseURL, productPath string, variantID int64) (string, error) {
	u, err := url.Parse(baseURL + productPath)
	if err != nil {
		return "", err
	}

	u.RawQuery = ""
	query := u.Query()
	query.Set("variant", fmt.Sprint(variantID))
	query.Set("utm_source", config.UtmSource)
	u.RawQuery = query.Encode()

	return u.String(), nil
}

func formatCardName(scrapVariant int, productTitle, variantTitle string) string {
	productTitle = strings.TrimSpace(productTitle)
	variantTitle = strings.TrimSpace(variantTitle)

	switch scrapVariant {
	case 2:
		if variantTitle == "" {
			return productTitle
		}
		return strings.TrimSpace(productTitle + " - " + variantTitle)
	case 3:
		return stripTrailingSet(productTitle)
	default:
		return productTitle
	}
}

func stripTrailingSet(productTitle string) string {
	title := strings.TrimSpace(productTitle)
	open := strings.LastIndex(title, "[")
	close := strings.LastIndex(title, "]")
	if open >= 0 && close > open && close == len(title)-1 {
		return strings.TrimSpace(title[:open])
	}
	return title
}

func extractSetName(productTitle string) string {
	title := strings.TrimSpace(productTitle)
	open := strings.LastIndex(title, "[")
	close := strings.LastIndex(title, "]")
	if open >= 0 && close > open && close == len(title)-1 {
		return strings.TrimSpace(title[open+1 : close])
	}
	return ""
}

func buildCardImageURL(rawImageURL, cardTitle string) string {
	img := strings.TrimSpace(rawImageURL)
	if strings.HasPrefix(img, "//") {
		return "https:" + img
	}
	if strings.HasPrefix(img, "http://") || strings.HasPrefix(img, "https://") {
		return img
	}
	return fmt.Sprintf("https://placehold.co/304x424?text=%s", url.QueryEscape(strings.TrimSpace(cardTitle)))
}
