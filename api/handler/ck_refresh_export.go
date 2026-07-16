package handler

import (
	"context"
	"errors"
	"log"
	"time"

	"mtg-price-checker-sg/store/ckpricereport"
	"mtg-price-checker-sg/store/ckprices"
)

type latestReportReader interface {
	ReadLatest(context.Context) (*ckpricereport.Report, error)
}

func selectPriceChanges(pre, post *ckprices.TopBottomPriceChanges) *ckprices.TopBottomPriceChanges {
	if changesHaveMovers(post) {
		return post
	}
	if changesHaveMovers(pre) {
		return pre
	}
	if post != nil {
		return post
	}
	return &ckprices.TopBottomPriceChanges{}
}

func changesHaveMovers(changes *ckprices.TopBottomPriceChanges) bool {
	return ckpricereport.HasMovers(ckpricereport.NewReport(changes, time.Time{}))
}

func writePriceChangeReport(
	ctx context.Context,
	writer ckpricereport.Writer,
	report *ckpricereport.Report,
) (preserved bool, existing *ckpricereport.Report, err error) {
	if ckpricereport.HasMovers(report) {
		return false, nil, writer.Write(ctx, report)
	}

	reader, ok := writer.(latestReportReader)
	if !ok {
		return false, nil, writer.Write(ctx, report)
	}

	existing, err = reader.ReadLatest(ctx)
	if err != nil {
		if errors.Is(err, ckpricereport.ErrLatestReportNotFound) {
			return false, nil, writer.Write(ctx, report)
		}
		return false, nil, err
	}
	if ckpricereport.HasMovers(existing) {
		return true, existing, nil
	}
	return false, nil, writer.Write(ctx, report)
}

func logPriceChangeExport(
	pre, post, selected *ckprices.TopBottomPriceChanges,
	preserved bool,
	existing *ckpricereport.Report,
	report *ckpricereport.Report,
) {
	switch {
	case preserved && existing != nil:
		log.Printf(
			"ck price refresh: preserved existing price change export top=%d bottom=%d generatedAt=%s",
			len(existing.Top),
			len(existing.Bottom),
			existing.GeneratedAt,
		)
	case changesHaveMovers(post):
		log.Printf("ck price refresh: exported post-refresh price changes top=%d bottom=%d", len(report.Top), len(report.Bottom))
	case changesHaveMovers(pre):
		log.Printf(
			"ck price refresh: exported pre-refresh price changes top=%d bottom=%d (post-refresh had none)",
			len(report.Top),
			len(report.Bottom),
		)
	default:
		log.Printf("ck price refresh: exported price changes top=%d bottom=%d generatedAt=%s", len(report.Top), len(report.Bottom), report.GeneratedAt)
	}

	if changesHaveMovers(pre) && !changesHaveMovers(post) && changesHaveMovers(selected) && selected == pre {
		log.Printf("ck price refresh: using pre-refresh rankings after same-day re-run")
	}
}

func exportCounts(preserved bool, existing, report *ckpricereport.Report) (topCount, bottomCount int, generatedAt string) {
	if preserved && existing != nil {
		return len(existing.Top), len(existing.Bottom), existing.GeneratedAt
	}
	return len(report.Top), len(report.Bottom), report.GeneratedAt
}
