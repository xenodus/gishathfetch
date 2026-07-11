package handler

import (
	"context"
	"log"
	"time"

	"mtg-price-checker-sg/controller/ckprice"
	"mtg-price-checker-sg/store/ckpricereport"
	"mtg-price-checker-sg/store/ckprices"
)

const ckPriceRefreshRunAction = "ck-price-refresh-run"

var (
	newCKRefreshStoreFunc = func(ctx context.Context) (ckprices.Store, error) {
		return ckprices.NewDynamoDBStore(ctx)
	}
	refreshCKPricesFunc = ckprice.RefreshPrices
	newCKPriceReportWriterFunc = func(ctx context.Context) (ckpricereport.Writer, error) {
		return ckpricereport.NewS3Writer(ctx)
	}
	ckPriceReportNowFunc = time.Now
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

	changes, err := store.GetTopBottomPriceChanges(ctx)
	if err != nil {
		log.Printf("ck price refresh: failed reading price changes: %v", err)
		return err
	}

	writer, err := newCKPriceReportWriterFunc(ctx)
	if err != nil {
		log.Printf("ck price refresh: failed opening s3 writer: %v", err)
		return err
	}

	report := ckpricereport.NewReport(changes, ckPriceReportNowFunc())
	if err = writer.Write(ctx, report); err != nil {
		log.Printf("ck price refresh: failed writing price change report: %v", err)
		return err
	}

	log.Printf("ck price refresh: exported price changes top=%d bottom=%d generatedAt=%s", len(report.Top), len(report.Bottom), report.GeneratedAt)
	return nil
}
