package tcgmarketplace

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"mtg-price-checker-sg/gateway"
)

const StoreName = "The TCG Marketplace"
const StoreBaseURL = "https://www.thetcgmarketplace.com/"

const cardLinkAPI = "https://thetcgmarketplace.com:3501/encoder/advancedsearch"
const mtgCategoryNo = 3
const accessTokenKey = "TCG_MARKETPLACE_ACCESS_TOKEN"

type response struct {
	Status int `json:"status"`
	Data   struct {
		Message string `json:"message"`
		Data    []struct {
			Name                  string      `json:"name"`
			Setcode               string      `json:"setcode"`
			Setname               string      `json:"setname"`
			Image                 string      `json:"image"`
			Language              string      `json:"language"`
			CrdFoilType           interface{} `json:"crd_foil_type"`
			Rarity                string      `json:"rarity"`
			Available             interface{} `json:"available"`
			From                  interface{} `json:"from"`
			NonFoilReferencePrice interface{} `json:"non_foil_reference_price"`
			FoilReferencePrice    interface{} `json:"foil_reference_price"`
			URL                   string      `json:"url"`
		} `json:"data"`
	} `json:"data"`
	Meta struct {
		Total int `json:"total"`
	} `json:"meta"`
}

type Store struct {
	Name      string
	BaseUrl   string
	SearchUrl string
}

type payload struct {
	AccessToken string `json:"access_token"`
	Name        string `json:"name"`
	Category    int32  `json:"category"`
	Order       string `json:"order"`
}

func NewLGS() gateway.LGS {
	return Store{
		Name:    StoreName,
		BaseUrl: StoreBaseURL,
	}
}

func (s Store) Search(searchStr string) ([]gateway.Card, error) {
	var (
		res         response
		cards       []gateway.Card
		accessToken string
	)

	godotenv.Load()
	accessToken = os.Getenv(accessTokenKey)

	reqPayload, err := json.Marshal(payload{
		AccessToken: accessToken,
		Name:        searchStr,
		Category:    mtgCategoryNo,
		Order:       "name_asc",
	})
	if err != nil {
		return cards, err
	}

	res, err = getApiResponse(reqPayload)
	if err != nil {
		return cards, err
	}

	if len(res.Data.Data) > 0 {
		for _, card := range res.Data.Data {
			stock, err := strconv.ParseInt(fmt.Sprint(card.Available), 10, 64)
			if err != nil {
				continue
			}

			if stock > 0 {
				price, err := strconv.ParseFloat(fmt.Sprint(card.From), 64)
				if err != nil {
					continue
				}

				// Strip [XXX] prefix from card name
				// e.g. [CMM] Deflecting Swat (V2)(Etched foil)
				name := strings.TrimSpace(card.Name)
				squareBracketIndex := strings.Index(name, "]")
				if squareBracketIndex > 1 {
					name = strings.TrimSpace(name[squareBracketIndex+1:])
				}

				if card.Setname != "" {
					name += " [" + card.Setname + "]"
				}

				var img string
				images := strings.Split(card.Image, " ")
				if len(images) > 0 {
					img = images[0]
				}

				cards = append(cards, gateway.Card{
					Name:    strings.TrimSpace(name),
					Url:     card.URL,
					InStock: true,
					Price:   price,
					Source:  s.Name,
					Img:     img,
				})
			}
		}
	}
	return cards, nil
}

func getApiResponse(payload []byte) (response, error) {
	var res response

	resp, err := http.Post(cardLinkAPI, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return res, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return res, err
	}

	err = json.Unmarshal(body, &res)
	if err != nil {
		return res, err
	}

	return res, nil
}
