package handler

import (
	"context"
	"log"
	"os"
	"strings"

	"mtg-price-checker-sg/controller/analyticskeywords"
	"mtg-price-checker-sg/gateway/ga4"
	"mtg-price-checker-sg/pkg/config"
	"mtg-price-checker-sg/store/analyticsreport"
)

const analyticsKeywordsExportRunAction = "analytics-keywords-export-run"

var (
	newGA4ReporterFunc = func(ctx context.Context) (ga4.Reporter, error) {
		return ga4.NewClient(ctx)
	}
	buildAnalyticsKeywordsReportFunc = analyticskeywords.BuildReport
	newAnalyticsReportWriterFunc = func(ctx context.Context) (analyticsreport.Writer, error) {
		return analyticsreport.NewS3Writer(ctx)
	}
)

func runAnalyticsKeywordsExport(ctx context.Context) (err error) {
	log.Printf("analytics keywords export: started")
	var generatedAt string

	defer func() {
		if err != nil {
			sendJobSlackAlert(formatAnalyticsKeywordsExportFailure(err))
			return
		}
		sendJobSlackAlert(formatAnalyticsKeywordsExportSuccess(generatedAt))
	}()

	reporter, err := newGA4ReporterFunc(ctx)
	if err != nil {
		log.Printf("analytics keywords export: failed opening ga4 client: %v", err)
		return err
	}

	propertyID := strings.TrimSpace(os.Getenv(config.GA4PropertyIDEnv))
	report, err := buildAnalyticsKeywordsReportFunc(ctx, reporter, propertyID, analyticskeywords.TopKeywordLimit)
	if err != nil {
		log.Printf("analytics keywords export: failed building report: %v", err)
		return err
	}

	writer, err := newAnalyticsReportWriterFunc(ctx)
	if err != nil {
		log.Printf("analytics keywords export: failed opening s3 writer: %v", err)
		return err
	}

	if err = writer.Write(ctx, report); err != nil {
		log.Printf("analytics keywords export: failed writing report: %v", err)
		return err
	}

	generatedAt = report.GeneratedAt
	log.Printf("analytics keywords export: finished generatedAt=%s", generatedAt)
	return nil
}
