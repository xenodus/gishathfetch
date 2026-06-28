package handler

import (
	"context"
	"log"

	"mtg-price-checker-sg/controller/ckprice"
	"mtg-price-checker-sg/store/ckprices"
)

const ckPriceRefreshRunAction = "ck-price-refresh-run"

var (
	newCKRefreshStoreFunc = func(ctx context.Context) (ckprices.Store, error) {
		return ckprices.NewDynamoDBStore(ctx)
	}
	refreshCKPricesFunc = ckprice.RefreshPrices
)

func runCKPriceRefresh(ctx context.Context) error {
	log.Printf("ck price refresh: started")
	store, err := newCKRefreshStoreFunc(ctx)
	if err != nil {
		log.Printf("ck price refresh: failed opening dynamodb store: %v", err)
		return err
	}

	count, err := refreshCKPricesFunc(ctx, store)
	if err != nil {
		log.Printf("ck price refresh: failed: %v", err)
		return err
	}

	log.Printf("ck price refresh: finished refreshed=%d", count)
	return nil
}
