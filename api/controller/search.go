package controller

import (
	"fmt"
	"log"
	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/agora"
	"mtg-price-checker-sg/gateway/arcanesanctum"
	"mtg-price-checker-sg/gateway/cardaffinity"
	"mtg-price-checker-sg/gateway/cardboardcrackgames"
	"mtg-price-checker-sg/gateway/cardsandcollection"
	"mtg-price-checker-sg/gateway/cardscitadel"
	"mtg-price-checker-sg/gateway/duellerpoint"
	"mtg-price-checker-sg/gateway/fivemana"
	"mtg-price-checker-sg/gateway/flagship"
	"mtg-price-checker-sg/gateway/gameshaven"
	"mtg-price-checker-sg/gateway/gog"
	"mtg-price-checker-sg/gateway/google"
	"mtg-price-checker-sg/gateway/hideout"
	"mtg-price-checker-sg/gateway/manapro"
	"mtg-price-checker-sg/gateway/moxandlotus"
	"mtg-price-checker-sg/gateway/mtgasia"
	"mtg-price-checker-sg/gateway/onemtg"
	"mtg-price-checker-sg/gateway/tcgmarketplace"
	"mtg-price-checker-sg/gateway/tefuda"
	"mtg-price-checker-sg/pkg/alert"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"
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
	IsFoil    bool    `json:"isFoil"`
	Source    string  `json:"src"`
	Quality   string  `json:"quality"`
	ExtraInfo string  `json:"extraInfo"`
}

func Search(input SearchInput) ([]Card, error) {
	shopNameToLGSMap := initAndMapShops(input.Lgs)
	return searchShops(input, shopNameToLGSMap)
}

func searchShops(input SearchInput, shopNameToLGSMap map[string]gateway.LGS) ([]Card, error) {
	var cards []gateway.Card
	var inStockCards, inStockExactMatchCards, inStockPartialMatchCards, inStockPrefixMatchCards []Card

	// use to track which lgs has result
	shopNameToHasResultMap := initShopHasResultMap(shopNameToLGSMap)

	if len(shopNameToLGSMap) > 0 {
		realStart := time.Now()
		responseThreshold := 1 * time.Second

		log.Printf("Start checking shops for [%s]...", input.SearchString)
		var wg sync.WaitGroup
		var mu sync.Mutex

		for shopName, lgs := range shopNameToLGSMap {
			wg.Go(func() {
				defer func() {
					if r := recover(); r != nil {
						errMsg := fmt.Sprintf("Recovered from panic in shop [%s]: %v", shopName, r)
						log.Println(errMsg)
						go alert.SendDiscordAlert(errMsg) // Send asynchronously
					}
				}()
				start := time.Now()
				c, err := lgs.Search(input.SearchString)
				if err != nil {
					log.Printf("Error encountered searching [%s]: %v", shopName, err)
				}
				log.Printf("Done searching [%s]. Took: [%s]", shopName, time.Since(start))

				if len(c) > 0 {
					mu.Lock()
					cards = append(cards, c...)
					mu.Unlock()
				}
			})
		}

		wg.Wait()
		log.Println("End checking shops...")

		if len(cards) > 0 {
			// Sort by price ASC
			sort.SliceStable(cards, func(i, j int) bool {
				return cards[i].Price < cards[j].Price
			})

			lowerSearchString := strings.ToLower(input.SearchString)

			// Only showing in stock, contains searched string and not art card
			for _, c := range cards {
				if c.InStock {
					cleanCardName, extraInfo := cleanName(c.Name, c.Quality, c.ExtraInfo)

					card := Card{
						Name:      cleanCardName,
						Url:       c.Url,
						Img:       c.Img,
						Price:     c.Price,
						InStock:   c.InStock,
						IsFoil:    c.IsFoil,
						Source:    c.Source,
						Quality:   c.Quality,
						ExtraInfo: strings.TrimSpace(strings.Join(extraInfo, " ")),
					}

					// replace all curly brackets with square brackets
					card.ExtraInfo = strings.Replace(card.ExtraInfo, "(", "[", -1)
					card.ExtraInfo = strings.Replace(card.ExtraInfo, ")", "]", -1)

					// Skip if detected as art card or Japanese
					if isArtCard(card.Name) || isJapanese(card.Name) || isArtCard(card.ExtraInfo) || isJapanese(card.ExtraInfo) {
						continue
					}

					lowerName := strings.ToLower(cleanCardName)

					// if in substring, mark lgs as having result
					if strings.Contains(lowerName, lowerSearchString) {
						shopNameToHasResultMap[c.Source] = true
					}

					// exact match
					if lowerName == lowerSearchString {
						inStockExactMatchCards = append(inStockExactMatchCards, card)
						continue
					}

					// prefix
					if strings.HasPrefix(lowerName, lowerSearchString) {
						inStockPrefixMatchCards = append(inStockPrefixMatchCards, card)
						continue
					}

					inStockPartialMatchCards = append(inStockPartialMatchCards, card)
				}
			}

			// order of results: exact > prefix > partial match
			inStockCards = append(inStockExactMatchCards, inStockPrefixMatchCards...)
			inStockCards = append(inStockCards, inStockPartialMatchCards...)
		}

		for shopName := range shopNameToHasResultMap {
			if !shopNameToHasResultMap[shopName] {
				log.Printf("Shop [%s] has no result for [%s]", shopName, input.SearchString)

				go func(lgs, searchString string) {
					err := google.LGSNoResultMeasurement(lgs, searchString)
					if err != nil {
						log.Printf("Error sending measurement for [%s]: %v", lgs, err)
					}
				}(shopName, input.SearchString)
			}
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
		arcanesanctum.StoreName:       arcanesanctum.NewLGS(),
		cardaffinity.StoreName:        cardaffinity.NewLGS(),
		cardboardcrackgames.StoreName: cardboardcrackgames.NewLGS(),
		cardscitadel.StoreName:        cardscitadel.NewLGS(),
		cardsandcollection.StoreName:  cardsandcollection.NewLGS(),
		duellerpoint.StoreName:        duellerpoint.NewLGS(),
		fivemana.StoreName:            fivemana.NewLGS(),
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
		// unsleeved.StoreName:           unsleeved.NewLGS(),
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

func initShopHasResultMap(lgsMap map[string]gateway.LGS) map[string]bool {
	shopNameToHasResultMap := make(map[string]bool, len(lgsMap))
	for shopName := range lgsMap {
		shopNameToHasResultMap[shopName] = false
	}
	return shopNameToHasResultMap
}

func isArtCard(s string) bool {
	return strings.Contains(strings.ToLower(s), "art card") ||
		strings.Contains(strings.ToLower(s), "art series")
}

func isJapanese(s string) bool {
	return strings.Contains(strings.ToLower(s), "japanese")
}

func cleanName(name, quality string, extraInfo []string) (string, []string) {
	cleanCardName := name

	// if we have quality, remove it from name
	if quality != "" {
		cleanCardName = strings.Replace(cleanCardName, quality, "", -1)

		if idx := strings.LastIndex(cleanCardName, " -"); idx != -1 {
			cleanCardName = cleanCardName[:idx]
		}
	}

	// if string has [, get index of it to strip [*] away
	squareBracketIndex := strings.Index(cleanCardName, "[")
	if squareBracketIndex > 0 {
		extraInfo = append(extraInfo, strings.TrimSpace(cleanCardName[squareBracketIndex:]))
		cleanCardName = strings.TrimSpace(cleanCardName[:squareBracketIndex])
	}

	// if string has (, get index of it to strip (*) away
	roundBracketIndex := strings.Index(cleanCardName, "(")
	if roundBracketIndex > 0 {
		extraInfo = append(extraInfo, strings.TrimSpace(cleanCardName[roundBracketIndex:]))
		cleanCardName = strings.TrimSpace(cleanCardName[:roundBracketIndex])
	}

	cleanCardName = strings.TrimSpace(cleanCardName)

	var extraInfoWithBrackets []string
	if len(extraInfo) > 0 {
		for _, info := range extraInfo {
			if !strings.HasPrefix(info, "[") && !strings.HasPrefix(info, "(") {
				extraInfoWithBrackets = append(extraInfoWithBrackets, "["+info+"]")
			} else {
				extraInfoWithBrackets = append(extraInfoWithBrackets, info)
			}
		}
	}
	return cleanCardName, extraInfoWithBrackets
}
