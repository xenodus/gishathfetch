package controller

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"mtg-price-checker-sg/gateway"
	"mtg-price-checker-sg/gateway/agora"
	"mtg-price-checker-sg/gateway/arcanesanctum"
	"mtg-price-checker-sg/gateway/cardaffinity"
	"mtg-price-checker-sg/gateway/cardsandcollection"
	"mtg-price-checker-sg/gateway/cardscentral"
	"mtg-price-checker-sg/gateway/cardscitadel"
	"mtg-price-checker-sg/gateway/duellerpoint"
	"mtg-price-checker-sg/gateway/fivemana"
	"mtg-price-checker-sg/gateway/flagship"
	"mtg-price-checker-sg/gateway/fyendalhobby"
	"mtg-price-checker-sg/gateway/gameshaven"
	"mtg-price-checker-sg/gateway/gog"
	"mtg-price-checker-sg/gateway/hideout"
	"mtg-price-checker-sg/gateway/hideyoshi"
	"mtg-price-checker-sg/gateway/manapro"
	"mtg-price-checker-sg/gateway/moxandlotus"
	"mtg-price-checker-sg/gateway/mtgasia"
	"mtg-price-checker-sg/gateway/onemtg"
	"mtg-price-checker-sg/gateway/tcgmarketplace"
	"mtg-price-checker-sg/gateway/tefuda"
	"mtg-price-checker-sg/gateway/util"
	"mtg-price-checker-sg/pkg/alert"
	"mtg-price-checker-sg/pkg/config"
	"mtg-price-checker-sg/pkg/logger"
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
	Store      string `json:"store"`
	Error      string `json:"error"`
	StatusCode int    `json:"statusCode,omitempty"`
}

// StoreStat reports per-store search timing and how many in-stock items were returned.
type StoreStat struct {
	Store      string `json:"store"`
	ItemCount  int    `json:"itemCount"`
	DurationMs int64  `json:"durationMs"`
}

var sendAlert = alert.SendAlert

type shopSpec struct {
	name        string
	newLGS      func() gateway.LGS
	isBinderpos bool
}

var shopRegistry = []shopSpec{
	{name: agora.StoreName, newLGS: agora.NewLGS},
	{name: arcanesanctum.StoreName, newLGS: arcanesanctum.NewLGS, isBinderpos: true},
	{name: cardaffinity.StoreName, newLGS: cardaffinity.NewLGS, isBinderpos: true},
	{name: cardscentral.StoreName, newLGS: cardscentral.NewLGS},
	{name: cardscitadel.StoreName, newLGS: cardscitadel.NewLGS, isBinderpos: true},
	{name: cardsandcollection.StoreName, newLGS: cardsandcollection.NewLGS},
	{name: duellerpoint.StoreName, newLGS: duellerpoint.NewLGS},
	{name: fivemana.StoreName, newLGS: fivemana.NewLGS},
	{name: flagship.StoreName, newLGS: flagship.NewLGS, isBinderpos: true},
	{name: fyendalhobby.StoreName, newLGS: fyendalhobby.NewLGS, isBinderpos: true},
	{name: gameshaven.StoreName, newLGS: gameshaven.NewLGS, isBinderpos: true},
	{name: gog.StoreName, newLGS: gog.NewLGS, isBinderpos: true},
	{name: hideout.StoreName, newLGS: hideout.NewLGS, isBinderpos: true},
	{name: hideyoshi.StoreName, newLGS: hideyoshi.NewLGS, isBinderpos: true},
	{name: manapro.StoreName, newLGS: manapro.NewLGS, isBinderpos: true},
	{name: moxandlotus.StoreName, newLGS: moxandlotus.NewLGS},
	{name: mtgasia.StoreName, newLGS: mtgasia.NewLGS, isBinderpos: true},
	{name: onemtg.StoreName, newLGS: onemtg.NewLGS, isBinderpos: true},
	{name: tcgmarketplace.StoreName, newLGS: tcgmarketplace.NewLGS},
	{name: tefuda.StoreName, newLGS: tefuda.NewLGS},
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

func Search(ctx context.Context, input SearchInput) ([]Card, []StoreError, []StoreStat, error) {
	shopNameToLGSMap := initAndMapShops(input.Lgs)
	return searchShops(ctx, input, shopNameToLGSMap)
}

func searchShops(ctx context.Context, input SearchInput, shopNameToLGSMap map[string]gateway.LGS) ([]Card, []StoreError, []StoreStat, error) {
	if len(shopNameToLGSMap) == 0 {
		return nil, []StoreError{}, []StoreStat{}, nil
	}

	realStart := time.Now()
	responseThreshold := 1 * time.Second

	// 1. Fetch concurrently (each store may search original + stripped forms)
	cards, siteErrors, shopDurations := fetchCardsConcurrently(ctx, input.SearchString, shopNameToLGSMap)
	_ = siteErrors // available for future use (e.g. partial-failure UX)

	// 2. Filter and Sort
	var inStockCards []Card
	if len(cards) > 0 {
		inStockCards = filterAndSortCards(cards, input.SearchString)
	}

	// 3. Ensure request takes at least the threshold
	if time.Since(realStart) < responseThreshold {
		sleepDuration := responseThreshold - time.Since(realStart)
		time.Sleep(sleepDuration)
		logger.From(ctx).InfoContext(ctx, "Sleeping for minimum response threshold", "duration", sleepDuration)
	}

	return inStockCards, buildStoreErrors(siteErrors), buildStoreStats(shopDurations, inStockCards), nil
}

const maxConcurrentStoreSearches = 9

type shopSearchJob struct {
	name string
	lgs  gateway.LGS
}

func fetchCardsConcurrently(ctx context.Context, searchString string, shops map[string]gateway.LGS) ([]gateway.Card, map[string]error, []shopSearchDuration) {
	var wg sync.WaitGroup
	aggregator := newFetchResultAggregator(len(shops))

	start := time.Now()

	jobs := make(chan shopSearchJob, len(shops))
	for shopName, lgs := range shops {
		jobs <- shopSearchJob{name: shopName, lgs: lgs}
	}
	close(jobs)

	workerCount := min(len(shops), maxConcurrentStoreSearches)

	for range workerCount {
		wg.Go(func() {
			for job := range jobs {
				searchShop(ctx, searchString, job.name, job.lgs, aggregator)
			}
		})
	}

	wg.Wait()
	cards, siteErrors, alertErrorMessages := aggregator.snapshot()
	shopDurations := aggregator.shopDurationSnapshot()
	if len(alertErrorMessages) > 0 {
		// Send synchronously so Lambda does not freeze the runtime before the
		// Slack webhook POST completes (fire-and-forget goroutines are dropped).
		sendAlert(formatAlertErrorSummary(searchString, alertErrorMessages))
	}
	if len(siteErrors) > 0 {
		logger.From(ctx).InfoContext(ctx, "Shops with errors", "search", searchString, "count", len(siteErrors))
	}
	logger.From(ctx).InfoContext(ctx, formatShopSearchSummary(searchString, time.Since(start), shopDurations))
	return cards, siteErrors, shopDurations
}

type shopSearchDuration struct {
	name     string
	duration time.Duration
}

type fetchResultAggregator struct {
	mu                 sync.Mutex
	cards              []gateway.Card
	siteErrors         map[string]error
	alertErrorMessages []string
	shopDurations      []shopSearchDuration
}

func newFetchResultAggregator(shopCount int) *fetchResultAggregator {
	return &fetchResultAggregator{
		cards:              []gateway.Card{},
		siteErrors:         make(map[string]error, shopCount),
		alertErrorMessages: make([]string, 0, shopCount),
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

func (f *fetchResultAggregator) addAlertErrorMessage(message string) {
	f.mu.Lock()
	f.alertErrorMessages = append(f.alertErrorMessages, message)
	f.mu.Unlock()
}

func (f *fetchResultAggregator) addShopDuration(shopName string, duration time.Duration) {
	f.mu.Lock()
	f.shopDurations = append(f.shopDurations, shopSearchDuration{name: shopName, duration: duration})
	f.mu.Unlock()
}

func (f *fetchResultAggregator) clearShopFailure(shopName string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	delete(f.siteErrors, shopName)

	if len(f.alertErrorMessages) == 0 {
		return
	}

	filtered := f.alertErrorMessages[:0]
	for _, message := range f.alertErrorMessages {
		if isShopFailureAlertMessage(message, shopName) {
			continue
		}
		filtered = append(filtered, message)
	}
	f.alertErrorMessages = filtered
}

func isShopFailureAlertMessage(message, shopName string) bool {
	if strings.HasPrefix(message, fmt.Sprintf("Recovered from panic in shop [%s]:", shopName)) {
		return true
	}
	return strings.HasPrefix(message, fmt.Sprintf("Error encountered searching [%s] for [", shopName))
}

func (f *fetchResultAggregator) shopDurationSnapshot() []shopSearchDuration {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]shopSearchDuration(nil), f.shopDurations...)
}

func (f *fetchResultAggregator) snapshot() ([]gateway.Card, map[string]error, []string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	cards := append([]gateway.Card(nil), f.cards...)
	siteErrors := make(map[string]error, len(f.siteErrors))
	maps.Copy(siteErrors, f.siteErrors)
	alertErrorMessages := append([]string(nil), f.alertErrorMessages...)

	return cards, siteErrors, alertErrorMessages
}

func searchShop(
	ctx context.Context,
	searchString string,
	shopName string,
	lgs gateway.LGS,
	aggregator *fetchResultAggregator,
) {
	defer recoverShopPanic(shopName, aggregator)
	start := time.Now()
	defer func() {
		aggregator.addShopDuration(shopName, time.Since(start))
	}()

	searchCtx := ctx
	if config.UseProxy {
		if proxyURLs := util.GetDedicatedProxyURLs(); len(proxyURLs) > 0 {
			// Wait for a proxy slot on the request context so queue time does not
			// consume the per-store search budget (see PerSiteTimeout below).
			releaseSearchSlot, slotErr := gateway.AcquireDedicatedProxySearchSlot(ctx)
			if slotErr != nil {
				recordShopSearchError(searchString, shopName, slotErr, aggregator)
				return
			}
			defer releaseSearchSlot()

			if proxyURL, release, err := gateway.LeaseDedicatedProxyURL(ctx, proxyURLs); err == nil {
				defer release()
				searchCtx = gateway.WithRequestDedicatedProxy(ctx, proxyURL)
			}
		}
	}

	shopCtx, cancel := context.WithTimeout(searchCtx, config.PerSiteTimeout)
	defer cancel()

	queries := util.StoreSearchQueries(searchString)
	if len(queries) == 0 {
		return
	}

	if len(queries) == 1 {
		cards, err := lgs.Search(shopCtx, queries[0])
		if err != nil {
			recordShopSearchError(searchString, shopName, err, aggregator)
		}
		aggregator.addCards(cards)
		return
	}

	var (
		wg                sync.WaitGroup
		mu                sync.Mutex
		allCards          []gateway.Card
		errs              []error
		anyQueryCompleted bool
	)

	for _, query := range queries {
		wg.Add(1)
		go func(query string) {
			defer wg.Done()

			var (
				queryCards    []gateway.Card
				queryErr      error
				queryFinished bool
			)
			func() {
				defer recoverShopPanic(shopName, aggregator)
				queryCards, queryErr = lgs.Search(shopCtx, query)
				queryFinished = true
			}()

			mu.Lock()
			defer mu.Unlock()
			if queryErr != nil {
				errs = append(errs, queryErr)
				return
			}
			if queryFinished {
				anyQueryCompleted = true
			}
			allCards = append(allCards, queryCards...)
		}(query)
	}
	wg.Wait()

	allCards = dedupeStoreCards(allCards)
	if len(allCards) > 0 || anyQueryCompleted {
		// A sibling query may have panicked or errored while another variant still
		// completed successfully (with or without inventory).
		aggregator.clearShopFailure(shopName)
	} else if len(errs) > 0 {
		recordShopSearchError(searchString, shopName, errors.Join(errs...), aggregator)
	}
	aggregator.addCards(allCards)
}

func dedupeStoreCards(cards []gateway.Card) []gateway.Card {
	if len(cards) <= 1 {
		return cards
	}

	seen := make(map[string]struct{}, len(cards))
	deduped := make([]gateway.Card, 0, len(cards))
	for _, card := range cards {
		key := storeCardKey(card)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		deduped = append(deduped, card)
	}
	return deduped
}

func storeCardKey(card gateway.Card) string {
	if url := strings.TrimSpace(card.Url); url != "" {
		return url
	}
	return fmt.Sprintf(
		"%s|%s|%v|%0.2f",
		card.Name,
		card.Quality,
		card.IsFoil,
		card.Price,
	)
}

func recoverShopPanic(shopName string, aggregator *fetchResultAggregator) {
	if r := recover(); r != nil {
		errMsg := fmt.Sprintf("Recovered from panic in shop [%s]: %v", shopName, r)
		slog.Error(errMsg, "shop", shopName, "panic", r)
		aggregator.addSiteError(shopName, fmt.Errorf("panic: %v", r))
		aggregator.addAlertErrorMessage(errMsg)
	}
}

func recordShopSearchError(searchString, shopName string, err error, aggregator *fetchResultAggregator) {
	if !errors.Is(err, context.Canceled) {
		errMsg := fmt.Sprintf(
			"Error encountered searching [%s] for [%s]: %s",
			shopName,
			searchString,
			gateway.EnsureHTTPStatusInErrorMessage(err.Error()),
		)
		slog.Warn(errMsg, "shop", shopName, "search", searchString, "err", err)
		aggregator.addAlertErrorMessage(errMsg)
	}
	aggregator.addSiteError(shopName, err)
}

func formatShopSearchSummary(searchString string, totalDuration time.Duration, shopDurations []shopSearchDuration) string {
	sortedDurations := append([]shopSearchDuration(nil), shopDurations...)
	sort.Slice(sortedDurations, func(i, j int) bool {
		return sortedDurations[i].name < sortedDurations[j].name
	})

	shopSummaries := make([]string, 0, len(sortedDurations))
	for _, shopDuration := range sortedDurations {
		shopSummaries = append(shopSummaries, fmt.Sprintf("[%s] %s", shopDuration.name, shopDuration.duration))
	}

	return fmt.Sprintf(
		"Checked %d shops for [%s] in %s: %s",
		len(sortedDurations),
		searchString,
		totalDuration,
		strings.Join(shopSummaries, ", "),
	)
}

func formatAlertErrorSummary(searchString string, errorMessages []string) string {
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
		enrichedError := gateway.EnsureHTTPStatusInErrorMessage(err.Error())
		storeErrors = append(storeErrors, StoreError{
			Store:      storeName,
			Error:      enrichedError,
			StatusCode: gateway.ExtractHTTPStatusCode(enrichedError),
		})
	}

	if len(storeErrors) == 0 {
		return []StoreError{}
	}

	return storeErrors
}

func buildStoreStats(shopDurations []shopSearchDuration, cards []Card) []StoreStat {
	if len(shopDurations) == 0 {
		return []StoreStat{}
	}

	itemCounts := make(map[string]int, len(shopDurations))
	for _, card := range cards {
		if card.Source == "" {
			continue
		}
		itemCounts[card.Source]++
	}

	sortedDurations := append([]shopSearchDuration(nil), shopDurations...)
	sort.Slice(sortedDurations, func(i, j int) bool {
		if sortedDurations[i].duration != sortedDurations[j].duration {
			return sortedDurations[i].duration > sortedDurations[j].duration
		}
		return sortedDurations[i].name < sortedDurations[j].name
	})

	stats := make([]StoreStat, 0, len(sortedDurations))
	for _, shopDuration := range sortedDurations {
		stats = append(stats, StoreStat{
			Store:      shopDuration.name,
			ItemCount:  itemCounts[shopDuration.name],
			DurationMs: shopDuration.duration.Milliseconds(),
		})
	}
	return stats
}

func parseSearchErrorMessage(message, searchString string) (shopName, details string, ok bool) {
	const prefix = "Error encountered searching ["
	if !strings.HasPrefix(message, prefix) {
		return "", "", false
	}

	withoutPrefix := strings.TrimPrefix(message, prefix)
	const shopSuffix = "] for ["
	before, after, ok0 := strings.Cut(withoutPrefix, shopSuffix)
	if !ok0 {
		return "", "", false
	}

	shopName = before
	withoutShop := after

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

func filterAndSortCards(cards []gateway.Card, searchString string) []Card {
	var inStockCards, inStockExactMatchCards, inStockPartialMatchCards, inStockPrefixMatchCards []Card

	// Sort by price ASC
	sort.SliceStable(cards, func(i, j int) bool {
		return cards[i].Price < cards[j].Price
	})

	foldedSearchString := util.FoldForMatch(searchString)

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

			foldedName := util.FoldForMatch(cleanCardName)

			if !strings.Contains(foldedName, foldedSearchString) {
				// skip card if not in substring
				continue
			}

			// exact match
			if foldedName == foldedSearchString {
				inStockExactMatchCards = append(inStockExactMatchCards, card)
				continue
			}

			// prefix
			if strings.HasPrefix(foldedName, foldedSearchString) {
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

func initAndMapShops(lgs []string) map[string]gateway.LGS {
	selectedLGS := map[string]struct{}{}
	for _, storeName := range lgs {
		selectedLGS[storeName] = struct{}{}
	}

	lgsMap := make(map[string]gateway.LGS, len(shopRegistry))
	for _, shop := range shopRegistry {
		if shop.name == agora.StoreName && !config.AgoraSearchEnabled {
			continue
		}
		if len(selectedLGS) > 0 {
			if _, exists := selectedLGS[shop.name]; !exists {
				continue
			}
		}
		lgsMap[shop.name] = shop.newLGS()
	}

	return lgsMap
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

func extraInfoInnerText(info string) string {
	info = strings.TrimSpace(info)
	if len(info) >= 2 {
		switch info[0] {
		case '[':
			if info[len(info)-1] == ']' {
				return strings.TrimSpace(info[1 : len(info)-1])
			}
		case '(':
			if info[len(info)-1] == ')' {
				return strings.TrimSpace(info[1 : len(info)-1])
			}
		}
	}
	return info
}

func isNonemptyExtraInfo(info string) bool {
	return extraInfoInnerText(info) != ""
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
			info = strings.TrimSpace(info)
			if !isNonemptyExtraInfo(info) {
				continue
			}
			if !strings.HasPrefix(info, "[") && !strings.HasPrefix(info, "(") {
				extraInfoWithBrackets = append(extraInfoWithBrackets, "["+info+"]")
			} else {
				extraInfoWithBrackets = append(extraInfoWithBrackets, info)
			}
		}
	}
	return cleanCardName, extraInfoWithBrackets
}
