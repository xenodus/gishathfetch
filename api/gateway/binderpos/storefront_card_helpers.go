package binderpos

import (
	"fmt"
	"net/url"
	"strings"

	"mtg-price-checker-sg/pkg/config"
)

func isPlaceholderVariantTitle(title string) bool {
	return strings.EqualFold(strings.TrimSpace(title), "Default Title")
}

func effectiveVariantTitle(title string) string {
	title = strings.TrimSpace(title)
	if isPlaceholderVariantTitle(title) {
		return ""
	}
	return title
}

func qualityFromVariantTitle(title string) string {
	title = effectiveVariantTitle(title)
	if title == "" {
		return ""
	}
	quality := strings.TrimSpace(strings.ReplaceAll(title, "Foil", ""))
	return strings.Join(strings.Fields(quality), " ")
}

func resolveCardQuality(productTitle, variantTitle string) string {
	if quality := qualityFromVariantTitle(variantTitle); quality != "" {
		return quality
	}
	return qualityFromProductTitle(productTitle)
}

func qualityFromProductTitle(productTitle string) string {
	segment := embeddedQualitySegment(productTitle)
	if segment == "" {
		return ""
	}
	quality := strings.TrimSpace(strings.ReplaceAll(segment, "Foil", ""))
	quality = strings.Join(strings.Fields(quality), " ")
	if !looksLikeCardQuality(quality) {
		return ""
	}
	return quality
}

func embeddedQualitySegment(productTitle string) string {
	productTitle = strings.TrimSpace(productTitle)
	dash := embeddedQualityDashIndex(productTitle)
	if dash < 0 {
		return ""
	}
	after := strings.TrimSpace(productTitle[dash+len(embeddedQualitySeparator(productTitle, dash)):])
	if after == "" {
		return ""
	}
	if bracket := strings.Index(after, "["); bracket >= 0 {
		return strings.TrimSpace(after[:bracket])
	}
	return after
}

func embeddedQualityDashIndex(productTitle string) int {
	for _, sep := range []string{" — ", " – "} {
		if i := strings.Index(productTitle, sep); i >= 0 {
			return i
		}
	}
	return -1
}

func embeddedQualitySeparator(productTitle string, dash int) string {
	if dash >= 0 && dash+len(" — ") <= len(productTitle) && productTitle[dash:dash+len(" — ")] == " — " {
		return " — "
	}
	return " – "
}

func looksLikeCardQuality(quality string) bool {
	switch strings.ToUpper(strings.TrimSpace(quality)) {
	case "NM", "NM/M", "LP", "MP", "HP", "DM", "EX/EX+",
		"NEAR MINT", "LIGHTLY PLAYED", "MODERATELY PLAYED", "HEAVILY PLAYED", "DAMAGED", "EXCELLENT":
		return true
	default:
		return false
	}
}

func stripEmbeddedQualityFromProductTitle(productTitle string) string {
	productTitle = strings.TrimSpace(productTitle)
	dash := embeddedQualityDashIndex(productTitle)
	if dash < 0 {
		return productTitle
	}
	segment := embeddedQualitySegment(productTitle)
	if segment == "" {
		return productTitle
	}
	quality := strings.TrimSpace(strings.ReplaceAll(segment, "Foil", ""))
	quality = strings.Join(strings.Fields(quality), " ")
	if !looksLikeCardQuality(quality) {
		return productTitle
	}

	name := strings.TrimSpace(productTitle[:dash])
	after := strings.TrimSpace(productTitle[dash+len(embeddedQualitySeparator(productTitle, dash)):])
	if bracket := strings.Index(after, "["); bracket >= 0 {
		rest := strings.TrimSpace(after[bracket:])
		if rest != "" {
			return strings.TrimSpace(name + " " + rest)
		}
	}
	return name
}

func productTitleIndicatesFoil(productTitle string) bool {
	return strings.Contains(strings.ToLower(embeddedQualitySegment(productTitle)), "foil")
}

func formatCardName(scrapVariant int, productTitle, variantTitle string) string {
	productTitle = strings.TrimSpace(productTitle)
	rawVariantTitle := strings.TrimSpace(variantTitle)
	variantTitle = effectiveVariantTitle(variantTitle)

	switch scrapVariant {
	case 2:
		if variantTitle == "" {
			if isPlaceholderVariantTitle(rawVariantTitle) {
				return stripEmbeddedQualityFromProductTitle(productTitle)
			}
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
