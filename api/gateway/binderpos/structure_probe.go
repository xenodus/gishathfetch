package binderpos

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"mtg-price-checker-sg/gateway"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/require"
)

// StructureProbeConfig identifies a BinderPOS storefront for structure checks.
type StructureProbeConfig struct {
	ScrapVariant int
	BaseURL      string
	SearchURL    string
	Query        string
}

// ProbeScrapeStructure fetches the storefront search page and verifies expected
// HTML markers for the scrap variant are still present. It does not require
// matching in-stock inventory.
func ProbeScrapeStructure(ctx context.Context, scrapVariant int, baseURL, searchURLTemplate, searchStr string) error {
	searchQuery := searchStr + " mtg"
	if scrapVariant == 4 {
		searchQuery = fyendalSearchQuery(searchStr)
	}
	searchURL := buildSafeSearchURL(baseURL, searchURLTemplate, searchQuery)
	pageURL, err := url.Parse(searchURL)
	if err != nil {
		return err
	}

	resp, err := gateway.DoOutboundGET(ctx, searchURL, gateway.OutboundRequestOptions{
		Style:              gateway.OutboundStyleHTML,
		PageURL:            pageURL,
		ShopifySGDCurrency: true,
	}, binderposAttemptTimeout)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %s loading %s", resp.Status, searchURL)
	}

	body, err := gateway.ReadResponseBody(resp)
	if err != nil {
		return err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return err
	}

	primary, fallback := scrapeStructureSelectors(scrapVariant)
	if doc.Find(primary).Length() > 0 {
		return nil
	}
	if fallback != "" && doc.Find(fallback).Length() > 0 {
		return nil
	}

	return fmt.Errorf("scrape structure for variant %d not found on %s (selectors %q, fallback %q)",
		scrapVariant, searchURL, primary, fallback)
}

// RequireStorefrontStructure verifies the storefront scrape upstream structure.
func RequireStorefrontStructure(t *testing.T, ctx context.Context, cfg StructureProbeConfig) {
	t.Helper()
	query := cfg.Query
	if strings.TrimSpace(query) == "" {
		query = "Abrade"
	}

	require.NoError(t, ProbeScrapeStructure(ctx, cfg.ScrapVariant, cfg.BaseURL, cfg.SearchURL, query))
}

// RequireScrapeStructure is a testify wrapper around ProbeScrapeStructure.
func RequireScrapeStructure(t *testing.T, ctx context.Context, scrapVariant int, baseURL, searchURLTemplate, searchStr string) {
	t.Helper()
	require.NoError(t, ProbeScrapeStructure(ctx, scrapVariant, baseURL, searchURLTemplate, searchStr))
}

func scrapeStructureSelectors(scrapVariant int) (primary, fallback string) {
	switch scrapVariant {
	case 1:
		return "div.Norm", "div.container"
	case 2:
		return "div[data-product-variants]", "div.product-card-list2"
	case 3:
		return "div.productCard__card", "div.productChip__grid"
	case 4:
		return "div.product-item.product-item--vertical", "a.product-item__title"
	case 5:
		return "div.product-grid-container ul.product-grid", "div.product-grid-container"
	default:
		return "body", ""
	}
}
