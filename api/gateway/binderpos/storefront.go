package binderpos

import (
	"math/rand/v2"
	"time"
)

const (
	storefrontSuggestPath   = "/search/suggest.json"
	binderposDecklistAPIURL = "https://portal.binderpos.com/external/shopify/decklist"
	binderposDecklistType   = "mtg"
	binderposDecklistPct    = 70
	binderposAttemptTimeout = 2 * time.Second
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

var shouldUseDecklistEndpoint = func() bool {
	return useDecklistForRoll(rand.IntN(100))
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
