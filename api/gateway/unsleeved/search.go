package unsleeved

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/util"
	"mtg-price-checker-sg/pkg/config"
	"net/http"
	"net/url"
	"strings"
)

const StoreName = "Unsleeved"
const StoreBaseURL = "https://hitpay.shop/unsleeved"
const StoreSearchURL = "/search"
const StoreApiURL = "https://hitpay.shop/api/v1/products/search?keywords=%s"
const StoreHeaderIDKey = "hitpay-identifier"
const StoreHeaderIDVal = "9879a70f-ebff-4612-9818-2ad353f94dee"

type Store struct {
	Name      string
	BaseUrl   string
	SearchUrl string
}

type response struct {
	Data []struct {
		ID         string `json:"id"`
		BusinessID string `json:"business_id"`
		Categories []struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Handle string `json:"handle"`
		} `json:"categories"`
		Name            string `json:"name"`
		Handle          string `json:"handle"`
		IsPinned        bool   `json:"is_pinned"`
		Order           int    `json:"order"`
		Price           int    `json:"price"`
		Emoji           any    `json:"emoji"`
		VariationsCount int    `json:"variations_count"`
		Variations      []any  `json:"variations"`
		Image           struct {
			ID              string `json:"id"`
			Caption         string `json:"caption"`
			AltText         any    `json:"alt_text"`
			URL             string `json:"url"`
			Height          int    `json:"height"`
			Width           int    `json:"width"`
			OtherDimensions struct {
				Icon struct {
					Size   string `json:"size"`
					URL    string `json:"url"`
					Height int    `json:"height"`
					Width  int    `json:"width"`
				} `json:"icon"`
				Large struct {
					Size   string `json:"size"`
					URL    string `json:"url"`
					Height int    `json:"height"`
					Width  int    `json:"width"`
				} `json:"large"`
				Small struct {
					Size   string `json:"size"`
					URL    string `json:"url"`
					Height int    `json:"height"`
					Width  int    `json:"width"`
				} `json:"small"`
				Medium struct {
					Size   string `json:"size"`
					URL    string `json:"url"`
					Height int    `json:"height"`
					Width  int    `json:"width"`
				} `json:"medium"`
				Thumbnail struct {
					Size   string `json:"size"`
					URL    string `json:"url"`
					Height int    `json:"height"`
					Width  int    `json:"width"`
				} `json:"thumbnail"`
			} `json:"other_dimensions"`
		} `json:"image"`
		Available                  bool   `json:"available"`
		PriceStored                int    `json:"price_stored"`
		PriceBeforeDiscount        any    `json:"price_before_discount"`
		PriceBeforeDiscountDisplay any    `json:"price_before_discount_display"`
		PriceDisplay               string `json:"price_display"`
		VariationsPriceRange       struct {
			Display any `json:"display"`
			Min     any `json:"min"`
			Max     any `json:"max"`
		} `json:"variations_price_range"`
	} `json:"data"`
	Links struct {
		First string `json:"first"`
		Last  string `json:"last"`
		Prev  any    `json:"prev"`
		Next  any    `json:"next"`
	} `json:"links"`
	Meta struct {
		CurrentPage int `json:"current_page"`
		From        int `json:"from"`
		LastPage    int `json:"last_page"`
		Links       []struct {
			URL    any    `json:"url"`
			Label  string `json:"label"`
			Page   any    `json:"page"`
			Active bool   `json:"active"`
		} `json:"links"`
		Path    string `json:"path"`
		PerPage int    `json:"per_page"`
		To      int    `json:"to"`
		Total   int    `json:"total"`
	} `json:"meta"`
}

func NewLGS() gateway.LGS {
	return Store{
		Name:      StoreName,
		BaseUrl:   StoreBaseURL,
		SearchUrl: StoreSearchURL,
	}
}

func (s Store) Search(searchStr string) ([]gateway.Card, error) {
	var (
		res   response
		cards []gateway.Card
	)

	apiURL := fmt.Sprintf(StoreApiURL, url.QueryEscape(searchStr))

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return cards, err
	}

	// Set custom headers
	req.Header.Set(StoreHeaderIDKey, StoreHeaderIDVal)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return cards, err
	}

	err = json.Unmarshal(body, &res)
	if err != nil {
		return cards, err
	}

	if len(res.Data) > 0 {
		for _, card := range res.Data {
			price, err := util.ParsePrice(card.PriceDisplay)
			if err != nil {
				log.Printf("error parsing price for %s with value [%s]: %v", s.Name, card.PriceDisplay, err)
				return cards, err
			}

			if card.Available {
				u := fmt.Sprintf("%s/product/%s", StoreBaseURL, card.Handle)
				cleanPageURL, err := url.Parse(u)
				if err != nil {
					log.Printf("error parsing url for %s with value [%s]: %v", s.Name, u, err)
					return cards, err
				}
				cleanPageURL.RawQuery = url.Values{
					"utm_source": []string{config.UtmSource},
				}.Encode()

				cards = append(cards, gateway.Card{
					Name:    strings.TrimSpace(card.Name),
					Url:     cleanPageURL.String(),
					InStock: true,
					Price:   price,
					Source:  s.Name,
					Img:     card.Image.URL,
				})
			}
		}
	}

	return cards, nil
}
