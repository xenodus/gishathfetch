package handler

import (
	"fmt"
	"strings"

	"mtg-price-checker-sg/pkg/alert"
)

var sendJobAlert = alert.SendJobAlert

func formatCKPriceRefreshSuccess(refreshedCount, topCount, bottomCount int, generatedAt, transportOrder string) string {
	if strings.TrimSpace(transportOrder) == "" {
		transportOrder = "unknown"
	}
	return fmt.Sprintf(
		"CK price refresh finished: refreshed=%d, top=%d, bottom=%d, generatedAt=%s, transportOrder=%s",
		refreshedCount,
		topCount,
		bottomCount,
		generatedAt,
		transportOrder,
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
