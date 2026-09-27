package ga4

import (
	"context"
	"strings"
	"time"

	"mtg-price-checker-sg/controller"
	"mtg-price-checker-sg/pkg/config"
	"mtg-price-checker-sg/pkg/logger"
)

const (
	// LgsCardNotFoundEventName is emitted when an exact card search returns no
	// in-stock hits at a given store.
	LgsCardNotFoundEventName = "lgs_card_not_found"
	lgsParam                 = "lgs"
	searchSourceWebsite      = "website"
	websiteSearchClientID    = "gishath-website-search"
)

// StoresWithNoResults lists stores that completed search with zero in-stock items.
func StoresWithNoResults(
	searchedStores []string,
	stats []controller.StoreStat,
	storeErrors []controller.StoreError,
) []string {
	if len(searchedStores) == 0 {
		return nil
	}

	errored := make(map[string]struct{}, len(storeErrors))
	for _, storeErr := range storeErrors {
		store := strings.TrimSpace(storeErr.Store)
		if store != "" {
			errored[store] = struct{}{}
		}
	}

	itemCount := make(map[string]int, len(stats))
	for _, stat := range stats {
		store := strings.TrimSpace(stat.Store)
		if store == "" {
			continue
		}
		itemCount[store] = stat.ItemCount
	}

	var withoutStock []string
	for _, store := range searchedStores {
		store = strings.TrimSpace(store)
		if store == "" {
			continue
		}
		if _, failed := errored[store]; failed {
			continue
		}
		count, ok := itemCount[store]
		if !ok || count != 0 {
			continue
		}
		withoutStock = append(withoutStock, store)
	}
	return withoutStock
}

// TryRecordLgsCardNotFoundEvents sends lgs_card_not_found Measurement Protocol
// events when CK price lookup verified the card name and a store returned zero
// in-stock hits.
func TryRecordLgsCardNotFoundEvents(
	ctx context.Context,
	clientID string,
	verifiedCardName string,
	searchedStores []string,
	stats []controller.StoreStat,
	storeErrors []controller.StoreError,
) {
	if !config.GA4MeasurementConfigured() {
		return
	}

	canonical := strings.TrimSpace(verifiedCardName)
	if canonical == "" {
		return
	}

	trackCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	stores := StoresWithNoResults(searchedStores, stats, storeErrors)
	if len(stores) == 0 {
		return
	}

	clientID = strings.TrimSpace(clientID)
	if clientID == "" {
		clientID = websiteSearchClientID
	}

	sender, err := NewMeasurementSender()
	if err != nil {
		logger.From(ctx).WarnContext(ctx, "ga4 lgs_card_not_found skipped", "err", err)
		return
	}

	for _, store := range stores {
		if err := sender.sendMeasurementEvent(trackCtx, clientID, LgsCardNotFoundEventName, map[string]any{
			"search_term":          canonical,
			lgsParam:                 store,
			searchSourceParam:        searchSourceWebsite,
			"engagement_time_msec": 1,
		}); err != nil {
			logger.From(ctx).WarnContext(
				ctx,
				"ga4 lgs_card_not_found failed",
				"searchTerm", canonical,
				"lgs", store,
				"err", err,
			)
			return
		}
	}

	logger.From(ctx).InfoContext(
		ctx,
		"ga4 lgs_card_not_found events sent",
		"searchTerm", canonical,
		"storeCount", len(stores),
	)
}
