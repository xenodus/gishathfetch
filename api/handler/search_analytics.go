package handler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"mtg-price-checker-sg/controller"
	"mtg-price-checker-sg/gateway/ga4"
	"mtg-price-checker-sg/pkg/apiauth"

	"github.com/aws/aws-lambda-go/events"
)

var recordLgsCardNotFoundEventsFunc = ga4.TryRecordLgsCardNotFoundEvents

func recordSearchLgsNotFoundAnalytics(
	ctx context.Context,
	request events.APIGatewayProxyRequest,
	searchQuery string,
	searchedStores []string,
	stats []controller.StoreStat,
	storeErrors []controller.StoreError,
) {
	recordLgsCardNotFoundEventsFunc(
		ctx,
		measurementClientID(request),
		searchQuery,
		searchedStores,
		stats,
		storeErrors,
	)
}

func measurementClientID(request events.APIGatewayProxyRequest) string {
	token := cookieValue(headerValue(request.Headers, "cookie"), apiauth.SessionCookieName())
	if token == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:16])
}
