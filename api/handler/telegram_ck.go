package handler

import (
	"context"
	"net/http"
	"time"

	"mtg-price-checker-sg/gateway/cardkingdom"
	"mtg-price-checker-sg/pkg/config"

	"github.com/aws/aws-lambda-go/events"
)

// TelegramCKResponse is the payload for Telegram Card Kingdom lookups.
type TelegramCKResponse struct {
	Listing    *cardkingdom.Listing `json:"listing"`
	DurationMs int64                `json:"durationMs"`
}

// TelegramCK returns the cheapest fresh Card Kingdom listing for a card name.
func TelegramCK(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var apiRes events.APIGatewayProxyResponse
	origin := request.Headers["origin"]

	if request.HTTPMethod == "OPTIONS" {
		return optionsResponse(origin)
	}

	if request.HTTPMethod != http.MethodGet {
		return errorResponse(apiRes, origin, "method not allowed", http.StatusMethodNotAllowed)
	}

	if res, ok := enforceOriginVerify(apiRes, origin, request.Headers); !ok {
		return res, nil
	}
	if res, ok := enforceTelegramBotToken(apiRes, origin, request.Headers); !ok {
		return res, nil
	}

	if config.APIMaintenanceMode() {
		return maintenanceActiveResponse(apiRes, origin)
	}

	query := parseSearchQuery(request)
	if message, statusCode := validateSearchString(query.searchString); statusCode != 0 {
		return errorResponse(apiRes, origin, message, statusCode)
	}

	started := time.Now()
	listing, err := lookupCKPriceFunc(ctx, query.searchString)
	if err != nil {
		return errorResponse(apiRes, origin, "err looking up card kingdom price", http.StatusInternalServerError)
	}

	body := TelegramCKResponse{
		Listing:    listing,
		DurationMs: time.Since(started).Milliseconds(),
	}
	return jsonResponse(apiRes, origin, http.StatusOK, body)
}
