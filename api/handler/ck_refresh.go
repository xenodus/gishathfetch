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

func runCKPriceRefresh(ctx context.Context) (err error) {
	log.Printf("ck price refresh: started")
	var refreshedCount int
	var topCount int
	var bottomCount int
	var generatedAt string

	defer func() {
		if err != nil {
			sendJobAlert(formatCKPriceRefreshFailure(err))
			return
		}
		sendJobAlert(formatCKPriceRefreshSuccess(refreshedCount, topCount, bottomCount, generatedAt))
	}()

	store, err := newCKRefreshStoreFunc(ctx)
	if err != nil {
		log.Printf("ck price refresh: failed opening dynamodb store: %v", err)
		return err
	}

	refreshedCount, err = refreshCKPricesFunc(ctx, store)
	if err != nil {
		log.Printf("ck price refresh: failed: %v", err)
		return err
	}

	log.Printf("ck price refresh: finished refreshed=%d", refreshedCount)

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

	topCount = len(report.Top)
	bottomCount = len(report.Bottom)
	generatedAt = report.GeneratedAt
	log.Printf("ck price refresh: exported price changes top=%d bottom=%d generatedAt=%s", topCount, bottomCount, generatedAt)
	return nil
}
