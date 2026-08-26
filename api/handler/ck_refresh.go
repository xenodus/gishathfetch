package handler

import (
	"context"
	"time"

	"mtg-price-checker-sg/controller/ckprice"
	"mtg-price-checker-sg/pkg/logger"
	"mtg-price-checker-sg/store/ckpricereport"
	"mtg-price-checker-sg/store/ckprices"
)

const ckPriceRefreshRunAction = "ck-price-refresh-run"

var (
	newCKRefreshStoreFunc = func(ctx context.Context) (ckprices.Store, error) {
		return ckprices.NewDynamoDBStore(ctx)
	}
	refreshCKPricesFunc        = ckprice.RefreshPrices
	newCKPriceReportWriterFunc = func(ctx context.Context) (ckpricereport.Writer, error) {
		return ckpricereport.NewS3Writer(ctx)
	}
	ckPriceReportNowFunc = time.Now
)

func runCKPriceRefresh(ctx context.Context) (err error) {
	logger.From(ctx).InfoContext(ctx, "ck price refresh: started")
	var refreshedCount int
	var topCount int
	var bottomCount int
	var generatedAt string
	var transportOrder string

	defer func() {
		if err != nil {
			sendJobAlert(formatCKPriceRefreshFailure(err))
			return
		}
		sendJobAlert(formatCKPriceRefreshSuccess(refreshedCount, topCount, bottomCount, generatedAt, transportOrder))
	}()

	store, err := newCKRefreshStoreFunc(ctx)
	if err != nil {
		logger.From(ctx).ErrorContext(ctx, "ck price refresh: failed opening dynamodb store", "err", err)
		return err
	}

	refreshResult, err := refreshCKPricesFunc(ctx, store)
	if err != nil {
		logger.From(ctx).ErrorContext(ctx, "ck price refresh: failed", "err", err)
		return err
	}
	refreshedCount = refreshResult.ListingCount
	transportOrder = refreshResult.TransportOrder

	logger.From(ctx).InfoContext(ctx, "ck price refresh: finished",
		"refreshed", refreshedCount,
		"transportOrder", transportOrder,
	)

	changes, err := ckprices.TopBottomPriceChangesInPricelist(ctx, store, refreshResult.Listings)
	if err != nil {
		logger.From(ctx).ErrorContext(ctx, "ck price refresh: failed reading price changes", "err", err)
		return err
	}

	writer, err := newCKPriceReportWriterFunc(ctx)
	if err != nil {
		logger.From(ctx).ErrorContext(ctx, "ck price refresh: failed opening s3 writer", "err", err)
		return err
	}

	report := ckpricereport.NewReport(changes, ckPriceReportNowFunc())
	if err = writer.Write(ctx, report); err != nil {
		logger.From(ctx).ErrorContext(ctx, "ck price refresh: failed writing price change report", "err", err)
		return err
	}

	topCount = len(report.Top)
	bottomCount = len(report.Bottom)
	generatedAt = report.GeneratedAt
	logger.From(ctx).InfoContext(ctx, "ck price refresh: exported price changes",
		"top", topCount,
		"bottom", bottomCount,
		"generatedAt", generatedAt,
	)
	return nil
}
