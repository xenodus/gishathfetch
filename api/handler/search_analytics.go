package handler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"mtg-price-checker-sg/controller"
	"mtg-price-checker-sg/gateway/ga4"
	"mtg-price-checker-sg/pkg/apiauth"

	"github.com/aws/aws-lambda-go/events"
)

const lgsNotFoundAnalyticsTimeout = 5 * time.Second

var recordLgsCardNotFoundEventsFunc = ga4.TryRecordLgsCardNotFoundEvents

type searchLgsNotFoundAnalyticsInput struct {
	searchQuery        string
	searchedStores     []string
	stats              []controller.StoreStat
	storeErrors        []controller.StoreError
	ckLookupPerformed  bool
	ckVerifiedCardName string
}

func scheduleSearchLgsNotFoundAnalytics(
	ctx context.Context,
	request events.APIGatewayProxyRequest,
	input searchLgsNotFoundAnalyticsInput,
) {
	clientID := measurementClientID(request)
	searchedStores := append([]string(nil), input.searchedStores...)
	stats := append([]controller.StoreStat(nil), input.stats...)
	storeErrors := append([]controller.StoreError(nil), input.storeErrors...)

	go func() {
		trackCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), lgsNotFoundAnalyticsTimeout)
		defer cancel()

		recordLgsCardNotFoundEventsFunc(
			trackCtx,
			clientID,
			input.searchQuery,
			searchedStores,
			stats,
			storeErrors,
			input.ckLookupPerformed,
			input.ckVerifiedCardName,
		)
	}()
}

func measurementClientID(request events.APIGatewayProxyRequest) string {
	token := cookieValue(headerValue(request.Headers, "cookie"), apiauth.SessionCookieName())
	if token == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:16])
}
