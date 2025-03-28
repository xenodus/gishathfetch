package controller

import (
	"log"
	"slices"
	"sort"
	"strings"
	"time"

	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/agora"
	"mtg-price-checker-sg/gateway/cardaffinity"
	"mtg-price-checker-sg/gateway/cardboardcrackgames"
	"mtg-price-checker-sg/gateway/cardsandcollection"
	"mtg-price-checker-sg/gateway/cardscitadel"
	"mtg-price-checker-sg/gateway/duellerpoint"
	"mtg-price-checker-sg/gateway/flagship"
	"mtg-price-checker-sg/gateway/gameshaven"
	"mtg-price-checker-sg/gateway/gog"
	"mtg-price-checker-sg/gateway/hideout"
	"mtg-price-checker-sg/gateway/manapro"
	"mtg-price-checker-sg/gateway/moxandlotus"
	"mtg-price-checker-sg/gateway/mtgasia"
	"mtg-price-checker-sg/gateway/onemtg"
	"mtg-price-checker-sg/gateway/tcgmarketplace"
	"mtg-price-checker-sg/gateway/tefuda"
)

type SearchInput struct {
	SearchString string
	Lgs          []string
}

type Card struct {
	Name      string  `json:"name"`
	Url       string  `json:"url"`
	Img       string  `json:"img"`
	Price     float64 `json:"price"`
	InStock   bool    `json:"inStock"`
	Source    string  `json:"src"`
	Quality   string  `json:"quality"`
	ExtraInfo string  `json:"extraInfo"`
}

func Search(input SearchInput) ([]Card, error) {
	var cards []gateway.Card
	var inStockCards, inStockExactMatchCards, inStockPartialMatchCards, inStockPrefixMatchCards []Card

	shopNameToLGSMap := initAndMapShops(input.Lgs)

	if len(shopNameToLGSMap) > 0 {
		// Create a channel with a buffer size of shopNameToLGSMap
		done := make(chan bool, len(shopNameToLGSMap))

		realStart := time.Now()
		responseThreshold := 1 * time.Second

		log.Printf("Start checking shops for [%s]...", input.SearchString)
		for shopName, lgs := range shopNameToLGSMap {
			sName := shopName
			sLGS := lgs
			go func() {
				start := time.Now()
				c, err := sLGS.Search(input.SearchString)
				if err != nil {
					log.Printf("Error encountered searching [%s]: %v", sName, err)
				}
				log.Printf("Done searching [%s]. Took: [%s]", sName, time.Since(start))

				if len(c) > 0 {
					cards = append(cards, c...)
				}

				// Signal that the goroutine is done
				done <- true
			}()
		}

		// Wait for all goroutines to finish
		for i := 0; i < len(shopNameToLGSMap); i++ {
			<-done
		}
		log.Println("End checking shops...")

		if len(cards) > 0 {
			// Sort by price ASC
			sort.SliceStable(cards, func(i, j int) bool {
				return cards[i].Price < cards[j].Price
			})

			// Only showing in stock, contains searched string and not art card
			for _, c := range cards {
				if c.InStock && !strings.Contains(strings.ToLower(c.Name), "art card") {
					cleanCardName := c.Name

					// if we have quality, remove it from name
					if c.Quality != "" {
						cleanCardName = strings.Replace(cleanCardName, c.Quality, "", -1)
						cleanCardName = strings.Replace(cleanCardName, "-", "", -1)
					}

					var extraInfo []string

					// if string has [, get index of it to strip [*] away
					squareBracketIndex := strings.Index(cleanCardName, "[")
					if squareBracketIndex > 1 {
						extraInfo = append(extraInfo, strings.TrimSpace(cleanCardName[squareBracketIndex:]))
						cleanCardName = strings.TrimSpace(cleanCardName[:squareBracketIndex-1])
					}

					// if string has (, get index of it to strip (*) away
					roundBracketIndex := strings.Index(cleanCardName, "(")
					if roundBracketIndex > 1 {
						extraInfo = append(extraInfo, strings.TrimSpace(cleanCardName[roundBracketIndex:]))
						cleanCardName = strings.TrimSpace(cleanCardName[:roundBracketIndex-1])
					}

					newCard := Card{
						Name:      cleanCardName,
						Url:       c.Url,
						Img:       c.Img,
						Price:     c.Price,
						InStock:   c.InStock,
						Source:    c.Source,
						Quality:   c.Quality,
						ExtraInfo: strings.Join(extraInfo, " "),
					}

					// exact match
					if strings.ToLower(cleanCardName) == strings.ToLower(input.SearchString) {
						inStockExactMatchCards = append(inStockExactMatchCards, newCard)
						continue
					}

					// prefix
					if strings.HasPrefix(strings.ToLower(cleanCardName), strings.ToLower(input.SearchString)) {
						inStockPrefixMatchCards = append(inStockPrefixMatchCards, newCard)
						continue
					}

					inStockPartialMatchCards = append(inStockPartialMatchCards, newCard)
				}
			}

			// order of results: exact > prefix > partial match
			inStockCards = append(inStockExactMatchCards, inStockPrefixMatchCards...)
			inStockCards = append(inStockCards, inStockPartialMatchCards...)
		}

		// ensure request takes at least X (responseThreshold) seconds
		if time.Since(realStart) < responseThreshold {
			time.Sleep(responseThreshold - time.Since(realStart))
			log.Printf("Sleeping for [%s]", responseThreshold-time.Since(realStart))
		}
	}
	return inStockCards, nil
}

func initAndMapShops(lgs []string) map[string]gateway.LGS {
	lgsMap := map[string]gateway.LGS{
		agora.StoreName:               agora.NewLGS(),
		cardaffinity.StoreName:        cardaffinity.NewLGS(),
		cardboardcrackgames.StoreName: cardboardcrackgames.NewLGS(),
		cardscitadel.StoreName:        cardscitadel.NewLGS(),
		cardsandcollection.StoreName:  cardsandcollection.NewLGS(),
		duellerpoint.StoreName:        duellerpoint.NewLGS(),
		flagship.StoreName:            flagship.NewLGS(),
		gameshaven.StoreName:          gameshaven.NewLGS(),
		gog.StoreName:                 gog.NewLGS(),
		hideout.StoreName:             hideout.NewLGS(),
		manapro.StoreName:             manapro.NewLGS(),
		moxandlotus.StoreName:         moxandlotus.NewLGS(),
		mtgasia.StoreName:             mtgasia.NewLGS(),
		onemtg.StoreName:              onemtg.NewLGS(),
		tefuda.StoreName:              tefuda.NewLGS(),
		tcgmarketplace.StoreName:      tcgmarketplace.NewLGS(),
	}

	if len(lgs) > 0 {
		for storeName := range lgsMap {
			if !slices.Contains(lgs, storeName) {
				delete(lgsMap, storeName)
			}
		}
	}
	return lgsMap
}
