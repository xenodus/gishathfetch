package tcgmarketplace

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/pkg/config"
)

func displayNameAndImage(card listing) (name string, img string) {
	name = strings.TrimSpace(card.Name)
	squareBracketIndex := strings.Index(name, "]")
	if squareBracketIndex > 1 {
		name = strings.TrimSpace(name[squareBracketIndex+1:])
	}
	images := strings.Split(card.Image, " ")
	if len(images) > 0 {
		img = images[0]
	}
	return name, img
}

func cardFromListing(s Store, card listing) (gateway.Card, bool) {
	stock, err := strconv.ParseInt(fmt.Sprint(card.Available), 10, 64)
	if err != nil || stock <= 0 {
		return gateway.Card{}, false
	}
	price, err := strconv.ParseFloat(fmt.Sprint(card.From), 64)
	if err != nil {
		return gateway.Card{}, false
	}

	pageURL, err := resolveProductPageURL(card)
	if err != nil {
		return gateway.Card{}, false
	}

	name, img := displayNameAndImage(card)
	extraInfo := []string{fmt.Sprintf("[%s]", card.Setname)}
	return gateway.Card{
		Name:      strings.TrimSpace(name),
		Url:       pageURL,
		InStock:   true,
		Price:     price,
		Source:    s.Name,
		Img:       img,
		IsFoil:    isSurgeFoil(extraInfo, name),
		ExtraInfo: extraInfo,
	}, true
}

func resolveProductPageURL(card listing) (string, error) {
	if raw := strings.TrimSpace(card.URL); raw != "" {
		cleanPageURL, err := canonicalProductURL(raw)
		if err != nil {
			return "", err
		}
		cleanPageURL.RawQuery = url.Values{
			"utm_source": []string{config.UtmSource},
		}.Encode()
		return cleanPageURL.String(), nil
	}
	if card.ID <= 0 {
		return "", fmt.Errorf("%s: listing missing product id and url", StoreName)
	}
	u, err := productPageURLFromID(card.ID)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}
