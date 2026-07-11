package binderpos

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mtg-price-checker-sg/gateway/util"
	"net/url"
	"strconv"
	"strings"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/pkg/config"

	"github.com/gocolly/colly/v2"
)

func (i impl) Scrap(ctx context.Context, scrapVariant int, storeName, baseUrl, searchUrl, searchStr string) ([]gateway.Card, error) {
	return i.scrapWithCollectorFactory(ctx, scrapVariant, storeName, baseUrl, searchUrl, searchStr, newBinderposCollector)
}

func (i impl) scrapDynamic(ctx context.Context, scrapVariant int, storeName, baseUrl, searchUrl, searchStr string) ([]gateway.Card, error) {
	return i.scrapWithCollectorFactory(ctx, scrapVariant, storeName, baseUrl, searchUrl, searchStr, newDynamicNoRetryCollector)
}

func (i impl) scrapDirect(ctx context.Context, scrapVariant int, storeName, baseUrl, searchUrl, searchStr string) ([]gateway.Card, error) {
	return i.scrapWithCollectorFactory(ctx, scrapVariant, storeName, baseUrl, searchUrl, searchStr, newDirectNoRetryCollector)
}

func (i impl) scrapWithCollectorFactory(
	ctx context.Context,
	scrapVariant int,
	storeName, baseUrl, searchUrl, searchStr string,
	collectorFactory func(context.Context) (*colly.Collector, error),
) ([]gateway.Card, error) {
	switch scrapVariant {
	case 1:
		return scrapVariant1(ctx, storeName, baseUrl, searchUrl, searchStr, collectorFactory)
	case 2:
		return scrapVariant2(ctx, storeName, baseUrl, searchUrl, searchStr, collectorFactory)
	case 3:
		return scrapVariant3(ctx, storeName, baseUrl, searchUrl, searchStr, collectorFactory)
	case 4:
		return scrapVariant4(ctx, storeName, baseUrl, searchUrl, searchStr, collectorFactory)
	case 5:
		return scrapVariant5(ctx, storeName, baseUrl, searchUrl, searchStr, collectorFactory)
	}
	return []gateway.Card{}, fmt.Errorf("invalid scrap variant: %d", scrapVariant)
}

func newBinderposCollector(ctx context.Context) (*colly.Collector, error) {
	c := gateway.NewOptimizedCollectorForBinderpos(ctx)
	c.SetRequestTimeout(binderposAttemptTimeout)
	configureShopifyStorefrontCollector(c)
	return c, nil
}

func newDynamicNoRetryCollector(ctx context.Context) (*colly.Collector, error) {
	c, err := gateway.NewOptimizedCollectorNoRetryDynamic(ctx)
	if err != nil {
		return nil, err
	}
	c.SetRequestTimeout(binderposAttemptTimeout)
	configureShopifyStorefrontCollector(c)
	return c, nil
}

func newDirectNoRetryCollector(ctx context.Context) (*colly.Collector, error) {
	c := gateway.NewOptimizedCollectorNoRetryDirect(ctx)
	c.SetRequestTimeout(binderposAttemptTimeout)
	configureShopifyStorefrontCollector(c)
	return c, nil
}

func configureShopifyStorefrontCollector(c *colly.Collector) {
	c.OnRequest(func(r *colly.Request) {
		if r == nil || r.Headers == nil {
			return
		}
		gateway.ApplyShopifySGDCurrencyCookie(r.Headers)
	})
}

// buildSafeSearchURL safely constructs the URL using url.URL and url.Values to isolate user string input.
// This prevents SAST tools from flagging uncontrolled input data in network requests.
func buildSafeSearchURL(baseUrl, searchUrlTemplate, searchStr string) string {
	base, err := url.Parse(baseUrl)
	if err != nil {
		return baseUrl + fmt.Sprintf(searchUrlTemplate, url.QueryEscape(searchStr))
	}

	parts := strings.SplitN(searchUrlTemplate, "?", 2)
	base.Path = parts[0]

	if len(parts) > 1 {
		qVals := url.Values{}
		pairs := strings.SplitSeq(parts[1], "&")
		for pair := range pairs {
			kv := strings.SplitN(pair, "=", 2)
			if len(kv) == 2 {
				val := kv[1]
				val = strings.Replace(val, "%s", searchStr, 1)
				qVals.Add(kv[0], val)
			} else if len(kv) == 1 {
				qVals.Add(kv[0], "")
			}
		}

		// Un-escape asterisk so url matches old fmt.Sprintf matching for cardscitadel -> `?q=*%s*`
		base.RawQuery = strings.ReplaceAll(qVals.Encode(), "%2A", "*")
	}

	return base.String()
}

// arcane sanctum
func scrapVariant5(ctx context.Context, storeName, baseUrl, searchUrl, searchStr string, collectorFactory func(context.Context) (*colly.Collector, error)) ([]gateway.Card, error) {
	searchURL := buildSafeSearchURL(baseUrl, searchUrl, searchStr+" mtg")
	var cards []gateway.Card

	c, err := collectorFactory(ctx)
	if err != nil {
		return nil, err
	}

	c.OnHTML("body", func(e *colly.HTMLElement) {
		e.ForEach("div.product-grid-container ul.product-grid li", func(_ int, el *colly.HTMLElement) {
			name := el.ChildText("div.collection-product-info > h3.collection-product-title a")
			link := el.ChildAttr("div.collection-product-info > h3.collection-product-title a", "href")
			img := el.ChildAttr("div.collection-product-img img", "src")
			priceStr := el.ChildText("div.collection-product-info > span.collection-product-price")

			if priceStr != "" {
				price, err := util.ParsePrice(priceStr)
				if err != nil {
					log.Printf("error parsing price for %s with value [%s]: %v", storeName, priceStr, err)
					return
				}

				// url
				u := strings.TrimSpace(baseUrl + link)
				cleanPageURL, err := url.Parse(u)
				if err != nil {
					log.Printf("error parsing url for %s with value [%s]: %v", storeName, u, err)
					return
				}
				cleanPageURL.RawQuery = url.Values{
					"utm_source": []string{config.UtmSource},
				}.Encode()

				if price > 0 {
					cards = append(cards, gateway.Card{
						Name:      strings.TrimSpace(name),
						Url:       strings.TrimSpace(cleanPageURL.String()),
						InStock:   true,
						Price:     price,
						Source:    storeName,
						Img:       img,
						ExtraInfo: []string{el.ChildText("div.collection-variant-display")},
					})
				}
			}
		})
	})

	return cards, gateway.VisitWithProxyInfo(c, searchURL)
}

type pagination struct {
	last int
	url  string
}

// games haven
// gog
// hideout
func scrapVariant3(ctx context.Context, storeName, baseUrl, searchUrl, searchStr string, collectorFactory func(context.Context) (*colly.Collector, error)) ([]gateway.Card, error) {
	var cards []gateway.Card
	searchURL := buildSafeSearchURL(baseUrl, searchUrl, searchStr+" mtg")

	c, err := collectorFactory(ctx)
	if err != nil {
		return nil, err
	}

	c.OnHTML("body", func(e *colly.HTMLElement) {
		// get cards
		e.ForEach("div.productCard__card", func(_ int, el *colly.HTMLElement) {
			var (
				isInstock bool
				price     float64
			)

			// in stock
			if len(el.ChildTexts("div.productCard__button--outOfStock")) == 0 {
				isInstock = true
			}

			if isInstock {
				el.ForEach("ul.productChip__grid li", func(_ int, el2 *colly.HTMLElement) {
					if el2.Attr("data-variantavailable") == "true" && el2.Attr("data-variantqty") != "0" {
						priceStr := el2.Attr("data-variantprice")
						priceStr = strings.Replace(priceStr, "$", "", -1)
						priceStr = strings.Replace(priceStr, ",", "", -1)
						priceStr = strings.Replace(priceStr, "SGD", "", -1)
						price, _ = strconv.ParseFloat(strings.TrimSpace(priceStr), 64)
						price = price / 100

						u := strings.TrimSpace(baseUrl + el.ChildAttr("a", "href"))
						cleanPageURL, err := url.Parse(u)
						if err != nil {
							log.Printf("error parsing url for %s with value [%s]: %v", storeName, u, err)
							return
						}
						cleanPageURL.RawQuery = url.Values{
							"variant":    []string{el2.Attr("data-variantid")},
							"utm_source": []string{config.UtmSource},
						}.Encode()

						if price > 0 {
							cards = append(cards, gateway.Card{
								Name:      strings.TrimSpace(el.ChildText("p.productCard__title")),
								Url:       strings.TrimSpace(cleanPageURL.String()),
								InStock:   isInstock,
								Price:     price,
								Source:    storeName,
								Img:       strings.TrimSpace("https:" + el.ChildAttr("img", "data-src")),
								Quality:   util.MapQuality(el2.Text),
								IsFoil:    strings.Contains(strings.ToLower(el2.Attr("data-varianttitle")), "foil"),
								ExtraInfo: []string{el.ChildText("p.productCard__setName")},
							})
						}
					}
				})
			}
		})
	})

	return cards, gateway.VisitWithProxyInfo(c, searchURL)
}

const fyendalMTGSingleProductType = "MTG Single Cards"

func fyendalSearchQuery(searchStr string) string {
	return fmt.Sprintf(`product_type:%q AND %s`, fyendalMTGSingleProductType, searchStr)
}

const fyendalFoilTitlePrefix = "[foil]"

func parseFyendalNameAndFoil(title string) (string, bool) {
	trimmed := strings.TrimSpace(title)
	if strings.HasPrefix(strings.ToLower(trimmed), fyendalFoilTitlePrefix) {
		return strings.TrimSpace(trimmed[len(fyendalFoilTitlePrefix):]), true
	}
	return trimmed, false
}

func fyendalPriceTextFromSpans(moneyText, priceText string) string {
	if trimmed := strings.TrimSpace(moneyText); trimmed != "" {
		return trimmed
	}

	text := strings.TrimSpace(priceText)
	text = strings.TrimPrefix(text, "Sale price")
	text = strings.TrimPrefix(text, "Regular price")
	return strings.TrimSpace(text)
}

// fyendal hobby
func scrapVariant4(ctx context.Context, storeName, baseUrl, searchUrl, searchStr string, collectorFactory func(context.Context) (*colly.Collector, error)) ([]gateway.Card, error) {
	searchURL := buildSafeSearchURL(baseUrl, searchUrl, fyendalSearchQuery(searchStr))
	var cards []gateway.Card

	c, err := collectorFactory(ctx)
	if err != nil {
		return nil, err
	}

	c.OnHTML("body", func(e *colly.HTMLElement) {
		e.ForEach("div.product-item.product-item--vertical", func(_ int, el *colly.HTMLElement) {
			if strings.Contains(el.Text, "Sold out") {
				return
			}

			variantID := el.ChildAttr("form.product-item__action-list input[name='id']", "value")
			if strings.TrimSpace(variantID) == "" {
				return
			}

			title := strings.TrimSpace(el.ChildText("a.product-item__title"))
			if title == "" {
				return
			}

			priceStr := fyendalPriceTextFromSpans(
				el.ChildText("div.product-item__price-list span.money"),
				el.ChildText("div.product-item__price-list span.price"),
			)
			price, err := util.ParsePrice(priceStr)
			if err != nil || price <= 0 {
				log.Printf("error parsing price for %s with value [%s]: %v", storeName, priceStr, err)
				return
			}

			link := el.ChildAttr("a.product-item__image-wrapper", "href")
			if link == "" {
				link = el.ChildAttr("a.product-item__title", "href")
			}
			if link == "" {
				return
			}

			img := strings.TrimSpace(el.ChildAttr("img.product-item__primary-image", "src"))
			if strings.HasPrefix(img, "//") {
				img = "https:" + img
			}

			name, isFoil := parseFyendalNameAndFoil(title)

			cleanPageURL, err := url.Parse(strings.TrimSpace(baseUrl + link))
			if err != nil {
				log.Printf("error parsing url for %s with value [%s]: %v", storeName, link, err)
				return
			}
			cleanPageURL.RawQuery = url.Values{
				"variant":    []string{variantID},
				"utm_source": []string{config.UtmSource},
			}.Encode()

			cards = append(cards, gateway.Card{
				Name:    name,
				Url:     strings.TrimSpace(cleanPageURL.String()),
				InStock: true,
				IsFoil:  isFoil,
				Price:   price,
				Source:  storeName,
				Img:     img,
			})
		})
	})

	return cards, gateway.VisitWithProxyInfo(c, searchURL)
}

// card affinity
// cardboard crack games
// flagship games
// onemtg
// manapro
// mtgasia
func scrapVariant2(ctx context.Context, storeName, baseUrl, searchUrl, searchStr string, collectorFactory func(context.Context) (*colly.Collector, error)) ([]gateway.Card, error) {
	searchURL := buildSafeSearchURL(baseUrl, searchUrl, searchStr+" mtg")
	var cards []gateway.Card

	c, err := collectorFactory(ctx)
	if err != nil {
		return nil, err
	}

	c.OnHTML("body", func(e *colly.HTMLElement) {
		e.ForEach("div", func(_ int, el *colly.HTMLElement) {
			cardInfoStr := el.Attr("data-product-variants")
			if len(cardInfoStr) > 0 {
				if !shouldIncludeBinderposProduct(el.Attr("data-product-type"), el.Attr("data-product-tags")) {
					return
				}

				productId := el.Attr("data-product-id")
				var pageUrl, imgUrl string
				if len(productId) > 0 {
					pageUrl = e.ChildAttr("div.product-card-list2__"+productId+" a", "href")
					imgUrl = e.ChildAttr("div.product-card-list2__"+productId+" img", "src")
				}

				var cardInfo []CardInfo
				err := json.Unmarshal([]byte(cardInfoStr), &cardInfo)
				if err == nil {
					if len(cardInfo) > 0 && len(pageUrl) > 0 && len(imgUrl) > 0 {
						for _, card := range cardInfo {
							if !card.Available {
								continue
							}

							// url with variant (quality)
							cleanPageURL, err := url.Parse(strings.TrimSpace(baseUrl + pageUrl))
							if err != nil {
								log.Printf("error parsing url for %s with value [%s]: %v", storeName, pageUrl, err)
								return
							}
							cleanPageURL.RawQuery = url.Values{
								"variant":    []string{fmt.Sprint(card.ID)},
								"utm_source": []string{config.UtmSource},
							}.Encode()

							// use placeholder if no image url detected
							if strings.Contains(imgUrl, "no-image") {
								imgUrl = fmt.Sprintf("//placehold.co/304x424?text=%s", url.QueryEscape(strings.TrimSpace(card.Name)))
							}

							cards = append(cards, gateway.Card{
								Name:    strings.TrimSpace(card.Name),
								Url:     strings.TrimSpace(cleanPageURL.String()),
								InStock: card.Available,
								Price:   float64(card.Price) / 100,
								Source:  storeName,
								Img:     strings.TrimSpace("https:" + imgUrl),
								IsFoil:  strings.Contains(strings.ToLower(card.Title), "foil"),
								Quality: strings.TrimSpace(strings.Replace(card.Title, "Foil", "", -1)),
							})
						}
					}
				}
			}
		})
	})

	return cards, gateway.VisitWithProxyInfo(c, searchURL)
}

// cards citadel
func scrapVariant1(ctx context.Context, storeName, baseUrl, searchUrl, searchStr string, collectorFactory func(context.Context) (*colly.Collector, error)) ([]gateway.Card, error) {
	searchURL := buildSafeSearchURL(baseUrl, searchUrl, searchStr+" mtg")
	var cards []gateway.Card

	c, err := collectorFactory(ctx)
	if err != nil {
		return nil, err
	}

	c.OnHTML("div.container", func(e *colly.HTMLElement) {
		e.ForEach("div.Norm", func(_ int, el *colly.HTMLElement) {
			var isInstock bool

			addNowTexts := el.ChildTexts("div.addNow")
			addNowAttrs := el.ChildAttrs("div.addNow", "onclick")

			if len(addNowTexts) > 0 {
				for i := range addNowTexts {
					isInstock = addNowTexts[i] != ""

					if isInstock {
						priceStr := strings.TrimSpace(addNowTexts[i])

						variantId := ""
						if i < len(addNowAttrs) {
							parts := strings.Split(addNowAttrs[i], "'")
							if len(parts) > 1 {
								variantId = parts[1]
							}
						}

						price, quality, err := parsePriceAndQuality(priceStr)
						if err != nil {
							continue
						}

						u := strings.TrimSpace(baseUrl + el.ChildAttr("a", "href"))
						cleanPageURL, err := url.Parse(u)
						if err != nil {
							log.Printf("error parsing url for %s with value [%s]: %v", storeName, u, err)
							return
						}

						rawQuery := url.Values{
							"utm_source": []string{config.UtmSource},
						}
						if variantId != "" {
							rawQuery.Add("variant", variantId)
						}
						cleanPageURL.RawQuery = rawQuery.Encode()

						if price > 0 {
							cards = append(cards, gateway.Card{
								Name:    strings.TrimSpace(el.ChildText("p.productTitle")),
								Url:     strings.TrimSpace(cleanPageURL.String()),
								InStock: isInstock,
								Price:   price,
								Source:  storeName,
								Img:     strings.TrimSpace("https:" + el.ChildAttr("img", "src")),
								Quality: strings.TrimSpace(strings.Replace(quality, "Foil", "", -1)),
								IsFoil:  strings.Contains(strings.ToLower(quality), "foil"),
							})
						}
					}
				}
			}
		})
	})
	return cards, gateway.VisitWithProxyInfo(c, searchURL)
}

func parsePriceAndQuality(priceQualityStr string) (float64, string, error) {
	priceQualityStrSlice := strings.Split(priceQualityStr, " - ")
	if len(priceQualityStrSlice) == 2 {
		quality := strings.TrimSpace(priceQualityStrSlice[0])
		price, err := util.ParsePrice(priceQualityStrSlice[1])
		return price, quality, err
	}
	return 0, "", nil
}
