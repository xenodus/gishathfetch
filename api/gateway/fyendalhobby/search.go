package fyendalhobby

import (
	"context"
	"log"
	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/util"
	"mtg-price-checker-sg/pkg/config"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const StoreName = "Fyendal Hobby"
const StoreBaseURL = "https://fyendalhobby.com"
const StoreSearchPath = "/collections/magic-the-gathering"

type Store struct {
	Name       string
	BaseUrl    string
	SearchPath string
}

func NewLGS() gateway.LGS {
	return Store{
		Name:       StoreName,
		BaseUrl:      StoreBaseURL,
		SearchPath: StoreSearchPath,
	}
}

func (s Store) Search(ctx context.Context, searchStr string) ([]gateway.Card, error) {
	var cards []gateway.Card

	apiURL := &url.URL{
		Scheme: "https",
		Host:   "fyendalhobby.com",
		Path:   StoreSearchPath,
		RawQuery: url.Values{
			"q":                     {searchStr},
			"filter.v.availability": {"1"},
		}.Encode(),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL.String(), nil)
	if err != nil {
		return cards, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return cards, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return cards, err
	}

	doc.Find("div.product-item").Each(func(i int, se *goquery.Selection) {
		c := gateway.Card{
			Source: s.Name,
		}

		title := se.Find("a.product-item__title")
		titleText := strings.TrimSpace(title.Text())
		c.Name = strings.TrimSpace(strings.Replace(titleText, "[Foil]", "", -1))
		c.IsFoil = strings.Contains(strings.ToLower(titleText), "[foil]")

		productPath := title.AttrOr("href", "")
		if productPath == "" {
			productPath = se.Find("a.product-item__image-wrapper").AttrOr("href", "")
		}
		c.Url = StoreBaseURL + productPath

		c.Img = se.Find("img.product-item__primary-image").AttrOr("src", "")
		if strings.HasPrefix(c.Img, "//") {
			c.Img = "https:" + c.Img
		}

		addToCartBtn := se.Find("button.product-item__action-button.button--primary")
		c.InStock = addToCartBtn.Length() > 0 &&
			!addToCartBtn.HasClass("button--disabled") &&
			!strings.EqualFold(strings.TrimSpace(addToCartBtn.Text()), "sold out")

		price, err := util.ParsePrice(se.Find("div.product-item__price-list span.money").First().Text())
		if err != nil {
			c.InStock = false
		}
		c.Price = price

		cleanPageURL, err := url.Parse(c.Url)
		if err != nil {
			log.Printf("error parsing url for %s with value [%s]: %v", s.Name, c.Url, err)
			return
		}
		cleanPageURL.RawQuery = url.Values{
			"utm_source": []string{config.UtmSource},
		}.Encode()
		c.Url = cleanPageURL.String()

		if c.Name != "" && c.InStock {
			cards = append(cards, c)
		}
	})

	return cards, nil
}
