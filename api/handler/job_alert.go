package handler

import (
	"fmt"

	"mtg-price-checker-sg/pkg/alert"
)

var sendJobSlackAlert = alert.SendSlackAlert

func formatCKPriceRefreshSuccess(refreshedCount, topCount, bottomCount int, generatedAt string) string {
	return fmt.Sprintf(
		"CK price refresh finished: refreshed=%d, top=%d, bottom=%d, generatedAt=%s",
		refreshedCount,
		topCount,
		bottomCount,
		generatedAt,
	)
}

func formatCKPriceRefreshFailure(err error) string {
	return fmt.Sprintf("CK price refresh failed: %v", err)
}

func formatAnalyticsKeywordsExportSuccess(generatedAt string) string {
	return fmt.Sprintf("Analytics keywords export finished: generatedAt=%s", generatedAt)
}

func formatAnalyticsKeywordsExportFailure(err error) string {
	return fmt.Sprintf("Analytics keywords export failed: %v", err)
}
