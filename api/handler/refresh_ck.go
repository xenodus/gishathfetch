package handler

import (
	"context"
	"log"

	"mtg-price-checker-sg/controller/ckprice"
	"mtg-price-checker-sg/store/ckprices"

	"github.com/aws/aws-lambda-go/events"
)

var refreshCKPricesFunc = ckprice.RefreshPrices

func RefreshCKPrices(ctx context.Context, _ events.CloudWatchEvent) error {
	store, err := ckprices.NewDynamoDBStore(ctx)
	if err != nil {
		return err
	}

	count, err := refreshCKPricesFunc(ctx, store)
	if err != nil {
		return err
	}

	log.Printf("refreshed %d card kingdom prices", count)
	return nil
}
