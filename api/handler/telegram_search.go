package handler

import (
	"context"
	"net/http"
	"sync"
	"time"

	"mtg-price-checker-sg/controller"
	"mtg-price-checker-sg/gateway/ga4"
	"mtg-price-checker-sg/pkg/apiauth"
	"mtg-price-checker-sg/pkg/config"

	"github.com/aws/aws-lambda-go/events"
)

var trackTelegramSearchEventFunc = ga4.TrySendSearchEvent

// TelegramSearchResponse is the minimal payload for Telegram bot searches.
type TelegramSearchResponse struct {
	Cheapest        *controller.Card        `json:"cheapest"`
	PhotoURL        string                  `json:"photoUrl,omitempty"`
	ResultCount     int                     `json:"resultCount"`
	WebsiteURL      string                  `json:"websiteUrl"`
	Errors          []controller.StoreError `json:"errors"`
	TotalDurationMs int64                   `json:"totalDurationMs"`
}

// TelegramSearch returns the cheapest in-stock match and total result count.
func TelegramSearch(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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

	requestStart := time.Now()
	var (
		inStockCards []controller.Card
		storeErrors  []controller.StoreError
		searchErr    error
	)
	var wg sync.WaitGroup
	wg.Go(func() {
		trackTelegramSearchEventFunc(ctx, query.searchString)
	})
	wg.Go(func() {
		inStockCards, storeErrors, _, searchErr = searchFunc(ctx, controller.SearchInput{
			SearchString: query.searchString,
			Lgs:          query.lgs,
		})
	})
	wg.Wait()
	totalDurationMs := time.Since(requestStart).Milliseconds()

	if searchErr != nil {
		return errorResponse(apiRes, origin, "err searching for cards", http.StatusInternalServerError)
	}

	var cheapest *controller.Card
	if len(inStockCards) > 0 {
		cheapest = &inStockCards[0]
	}

	if storeErrors == nil {
		storeErrors = []controller.StoreError{}
	}

	body := TelegramSearchResponse{
		Cheapest:        cheapest,
		PhotoURL:        selectTelegramPhotoURL(ctx, inStockCards),
		ResultCount:     len(inStockCards),
		WebsiteURL:      buildWebsiteSearchURL(query.searchString, query.lgs),
		Errors:          storeErrors,
		TotalDurationMs: totalDurationMs,
	}

	return jsonResponse(apiRes, origin, http.StatusOK, body)
}

func enforceTelegramBotToken(
	apiRes events.APIGatewayProxyResponse,
	origin string,
	headers map[string]string,
) (events.APIGatewayProxyResponse, bool) {
	if config.APITelegramBotToken() == "" {
		return mustErrorResponse(apiRes, origin, "telegram search not configured", http.StatusServiceUnavailable), false
	}

	if err := apiauth.VerifyTelegramBotToken(headers); err != nil {
		return accessDeniedResponse(apiRes, origin, "forbidden"), false
	}

	return apiRes, true
}
