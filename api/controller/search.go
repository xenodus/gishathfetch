package controller

import (
	"context"
	"errors"
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
	"mtg-price-checker-sg/pkg/config"
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

type StoreError struct {
	Store string `json:"store"`
	Error string `json:"error"`
}

const binderposMaxConcurrent = 2

var sendDiscordAlert = alert.SendDiscordAlert

type shopSpec struct {
	name         string
	newLGS       func() gateway.LGS
	isBinderpos  bool
}

var shopRegistry = []shopSpec{
	{name: agora.StoreName, newLGS: agora.NewLGS},
	{name: arcanesanctum.StoreName, newLGS: arcanesanctum.NewLGS, isBinderpos: true},
	{name: cardaffinity.StoreName, newLGS: cardaffinity.NewLGS, isBinderpos: true},
	{name: cardboardcrackgames.StoreName, newLGS: cardboardcrackgames.NewLGS, isBinderpos: true},
	{name: cardscitadel.StoreName, newLGS: cardscitadel.NewLGS, isBinderpos: true},
	{name: cardsandcollection.StoreName, newLGS: cardsandcollection.NewLGS},
	{name: duellerpoint.StoreName, newLGS: duellerpoint.NewLGS},
	{name: fivemana.StoreName, newLGS: fivemana.NewLGS},
	{name: flagship.StoreName, newLGS: flagship.NewLGS, isBinderpos: true},
	{name: gameshaven.StoreName, newLGS: gameshaven.NewLGS, isBinderpos: true},
	{name: gog.StoreName, newLGS: gog.NewLGS, isBinderpos: true},
	{name: hideout.StoreName, newLGS: hideout.NewLGS, isBinderpos: true},
	{name: manapro.StoreName, newLGS: manapro.NewLGS, isBinderpos: true},
	{name: moxandlotus.StoreName, newLGS: moxandlotus.NewLGS},
	{name: mtgasia.StoreName, newLGS: mtgasia.NewLGS, isBinderpos: true},
	{name: onemtg.StoreName, newLGS: onemtg.NewLGS, isBinderpos: true},
	{name: tefuda.StoreName, newLGS: tefuda.NewLGS, isBinderpos: true},
	{name: tcgmarketplace.StoreName, newLGS: tcgmarketplace.NewLGS},
	// {name: unsleeved.StoreName, newLGS: unsleeved.NewLGS},
}

var binderposStoreNames = func() map[string]struct{} {
	storeNames := make(map[string]struct{}, len(shopRegistry))
	for _, shop := range shopRegistry {
		if shop.isBinderpos {
			storeNames[shop.name] = struct{}{}
		}
	}
	return storeNames
}()

func Search(ctx context.Context, input SearchInput) ([]Card, []StoreError, error) {
	shopNameToLGSMap := initAndMapShops(input.Lgs)
	return searchShops(ctx, input, shopNameToLGSMap)
}

func searchShops(ctx context.Context, input SearchInput, shopNameToLGSMap map[string]gateway.LGS) ([]Card, []StoreError, error) {
	shopNameToHasResultMap := initShopHasResultMap(shopNameToLGSMap)

	if len(shopNameToLGSMap) == 0 {
		return nil, []StoreError{}, nil
	}

	realStart := time.Now()
	responseThreshold := 1 * time.Second

	// 1. Fetch concurrently
	cards, siteErrors := fetchCardsConcurrently(ctx, input.SearchString, shopNameToLGSMap)
	_ = siteErrors // available for future use (e.g. partial-failure UX)

	// 2. Filter and Sort
	var inStockCards []Card
	if len(cards) > 0 {
		inStockCards = filterAndSortCards(cards, input.SearchString, shopNameToHasResultMap)
	}

	// 3. Report metrics for shops with no results
	reportNoResultMetrics(shopNameToHasResultMap, input.SearchString)

	// 4. Ensure request takes at least the threshold
	if time.Since(realStart) < responseThreshold {
		sleepDuration := responseThreshold - time.Since(realStart)
		time.Sleep(sleepDuration)
		log.Printf("Sleeping for [%s]", sleepDuration)
	}

	return inStockCards, buildStoreErrors(siteErrors), nil
}

func fetchCardsConcurrently(ctx context.Context, searchString string, shops map[string]gateway.LGS) ([]gateway.Card, map[string]error) {
	var wg sync.WaitGroup
	aggregator := newFetchResultAggregator(len(shops))
	binderposGate := make(chan struct{}, binderposMaxConcurrent)

	log.Printf("Start checking shops for [%s]...", searchString)

	for shopName, lgs := range shops {
		shopName := shopName
		lgs := lgs
		wg.Go(func() {
			searchShop(ctx, searchString, shopName, lgs, binderposGate, aggregator)
		})
	}

	wg.Wait()
	cards, siteErrors, discordErrorMessages := aggregator.snapshot()
	if len(discordErrorMessages) > 0 {
		go sendDiscordAlert(formatDiscordErrorSummary(searchString, discordErrorMessages))
	}
	if len(siteErrors) > 0 {
		log.Printf("Shops with errors for [%s]: %d", searchString, len(siteErrors))
	}
	log.Println("End checking shops...")
	return cards, siteErrors
}

type fetchResultAggregator struct {
	mu                   sync.Mutex
	cards                []gateway.Card
	siteErrors           map[string]error
	discordErrorMessages []string
}

func newFetchResultAggregator(shopCount int) *fetchResultAggregator {
	return &fetchResultAggregator{
		cards:                []gateway.Card{},
		siteErrors:           make(map[string]error, shopCount),
		discordErrorMessages: make([]string, 0, shopCount),
	}
}

func (f *fetchResultAggregator) addCards(cards []gateway.Card) {
	if len(cards) == 0 {
		return
	}
	f.mu.Lock()
	f.cards = append(f.cards, cards...)
	f.mu.Unlock()
}

func (f *fetchResultAggregator) addSiteError(shopName string, err error) {
	f.mu.Lock()
	f.siteErrors[shopName] = err
	f.mu.Unlock()
}

func (f *fetchResultAggregator) addDiscordErrorMessage(message string) {
	f.mu.Lock()
	f.discordErrorMessages = append(f.discordErrorMessages, message)
	f.mu.Unlock()
}

func (f *fetchResultAggregator) snapshot() ([]gateway.Card, map[string]error, []string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	cards := append([]gateway.Card(nil), f.cards...)
	siteErrors := make(map[string]error, len(f.siteErrors))
	for shopName, err := range f.siteErrors {
		siteErrors[shopName] = err
	}
	discordErrorMessages := append([]string(nil), f.discordErrorMessages...)

	return cards, siteErrors, discordErrorMessages
}

func searchShop(
	ctx context.Context,
	searchString string,
	shopName string,
	lgs gateway.LGS,
	binderposGate chan struct{},
	aggregator *fetchResultAggregator,
) {
	defer recoverShopPanic(shopName, aggregator)
	start := time.Now()
	defer log.Printf("Done searching [%s]. Took: [%s]", shopName, time.Since(start))

	shopCtx, cancel := context.WithTimeout(ctx, config.PerSiteTimeout)
	defer cancel()

	if isBinderposStore(shopName) {
		binderposGate <- struct{}{}
		defer func() { <-binderposGate }()
	}

	cards, err := lgs.Search(shopCtx, searchString)
	if err != nil {
		recordShopSearchError(searchString, shopName, err, aggregator)
	}
	aggregator.addCards(cards)
}

func recoverShopPanic(shopName string, aggregator *fetchResultAggregator) {
	if r := recover(); r != nil {
		errMsg := fmt.Sprintf("Recovered from panic in shop [%s]: %v", shopName, r)
		log.Println(errMsg)
		aggregator.addSiteError(shopName, fmt.Errorf("panic: %v", r))
		aggregator.addDiscordErrorMessage(errMsg)
	}
}

func recordShopSearchError(searchString, shopName string, err error, aggregator *fetchResultAggregator) {
	if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
		errMsg := fmt.Sprintf("Error encountered searching [%s] for [%s]: %v", shopName, searchString, err)
		log.Println(errMsg)
		aggregator.addDiscordErrorMessage(errMsg)
	}
	aggregator.addSiteError(shopName, err)
}

func formatDiscordErrorSummary(searchString string, errorMessages []string) string {
	sortedMessages := append([]string(nil), errorMessages...)
	sort.Strings(sortedMessages)

	formattedLines := make([]string, 0, len(sortedMessages))
	for _, message := range sortedMessages {
		shopName, details, ok := parseSearchErrorMessage(message, searchString)
		if ok {
			formattedLines = append(formattedLines, fmt.Sprintf("- [%s] %s", shopName, details))
			continue
		}

		formattedLines = append(formattedLines, fmt.Sprintf("- %s", message))
	}

	return fmt.Sprintf(
		"Encountered %d error(s) while searching [%s]:\n%s",
		len(sortedMessages),
		searchString,
		strings.Join(formattedLines, "\n"),
	)
}

func buildStoreErrors(siteErrors map[string]error) []StoreError {
	if len(siteErrors) == 0 {
		return []StoreError{}
	}

	storeNames := make([]string, 0, len(siteErrors))
	for storeName := range siteErrors {
		storeNames = append(storeNames, storeName)
	}
	sort.Strings(storeNames)

	storeErrors := make([]StoreError, 0, len(storeNames))
	for _, storeName := range storeNames {
		err := siteErrors[storeName]
		if err == nil {
			continue
		}
		storeErrors = append(storeErrors, StoreError{
			Store: storeName,
			Error: err.Error(),
		})
	}

	if len(storeErrors) == 0 {
		return []StoreError{}
	}

	return storeErrors
}

func parseSearchErrorMessage(message, searchString string) (shopName, details string, ok bool) {
	const prefix = "Error encountered searching ["
	if !strings.HasPrefix(message, prefix) {
		return "", "", false
	}

	withoutPrefix := strings.TrimPrefix(message, prefix)
	const shopSuffix = "] for ["
	shopSuffixIndex := strings.Index(withoutPrefix, shopSuffix)
	if shopSuffixIndex == -1 {
		return "", "", false
	}

	shopName = withoutPrefix[:shopSuffixIndex]
	withoutShop := withoutPrefix[shopSuffixIndex+len(shopSuffix):]

	searchSuffix := fmt.Sprintf("%s]: ", searchString)
	if !strings.HasPrefix(withoutShop, searchSuffix) {
		return "", "", false
	}

	details = strings.TrimPrefix(withoutShop, searchSuffix)
	if strings.TrimSpace(details) == "" {
		return "", "", false
	}

	return shopName, details, true
}

func filterAndSortCards(cards []gateway.Card, searchString string, shopNameToHasResultMap map[string]bool) []Card {
	var inStockCards, inStockExactMatchCards, inStockPartialMatchCards, inStockPrefixMatchCards []Card

	// Sort by price ASC
	sort.SliceStable(cards, func(i, j int) bool {
		return cards[i].Price < cards[j].Price
	})

	lowerSearchString := strings.ToLower(searchString)

	// Only showing in stock, contains searched string and not art card
	for _, c := range cards {
		if c.InStock && c.Price > 0 {
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
			} else {
				// skip card if not in substring
				continue
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

	return inStockCards
}

func reportNoResultMetrics(shopNameToHasResultMap map[string]bool, searchString string) {
	for shopName, hasResult := range shopNameToHasResultMap {
		if !hasResult {
			log.Printf("Shop [%s] has no result for [%s]", shopName, searchString)

			go func(lgs, searchString string) {
				err := google.LGSNoResultMeasurement(lgs, searchString)
				if err != nil {
					log.Printf("Error sending measurement for [%s]: %v", lgs, err)
				}
			}(shopName, searchString)
		}
	}
}

func initAndMapShops(lgs []string) map[string]gateway.LGS {
	selectedLGS := map[string]struct{}{}
	for _, storeName := range lgs {
		selectedLGS[storeName] = struct{}{}
	}

	lgsMap := make(map[string]gateway.LGS, len(shopRegistry))
	for _, shop := range shopRegistry {
		if len(selectedLGS) > 0 {
			if _, exists := selectedLGS[shop.name]; !exists {
				continue
			}
		}
		lgsMap[shop.name] = shop.newLGS()
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

func isBinderposStore(shopName string) bool {
	_, ok := binderposStoreNames[shopName]
	return ok
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
