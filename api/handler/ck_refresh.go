package handler

import (
	"context"
	"crypto/subtle"
	"log"
	"net/http"
	"os"
	"strings"

	"mtg-price-checker-sg/controller/ckprice"
	"mtg-price-checker-sg/pkg/config"
	"mtg-price-checker-sg/store/ckprices"

	"github.com/aws/aws-lambda-go/events"
)

type CKRefreshResponse struct {
	Refreshed int `json:"refreshed"`
}

var (
	newCKRefreshStoreFunc = func(ctx context.Context) (ckprices.Store, error) {
		return ckprices.NewDynamoDBStore(ctx)
	}
	refreshCKPricesFunc = ckprice.RefreshPrices
)

func CKPriceRefresh(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var apiRes events.APIGatewayProxyResponse

	if request.HTTPMethod == http.MethodOptions {
		return optionsResponse(request.Headers["origin"])
	}

	if request.HTTPMethod != http.MethodPost {
		return errorResponse(apiRes, "", "method not allowed", http.StatusMethodNotAllowed)
	}

	if !isValidRefreshAPIKey(request.Headers["x-api-key"]) {
		return errorResponse(apiRes, "", "unauthorized", http.StatusUnauthorized)
	}

	store, err := newCKRefreshStoreFunc(ctx)
	if err != nil {
		return errorResponse(apiRes, "", "card kingdom refresh is unavailable", http.StatusInternalServerError)
	}

	count, err := refreshCKPricesFunc(ctx, store)
	if err != nil {
		log.Printf("ck price refresh failed: %v", err)
		return errorResponse(apiRes, "", "err refreshing card kingdom prices", http.StatusInternalServerError)
	}

	log.Printf("refreshed %d card kingdom prices", count)
	return jsonResponse(apiRes, "", http.StatusOK, CKRefreshResponse{Refreshed: count})
}

func isValidRefreshAPIKey(provided string) bool {
	expected := strings.TrimSpace(os.Getenv(config.CKRefreshAPIKeyEnv))
	if expected == "" {
		return os.Getenv("ENV") != config.EnvProd
	}
	return subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) == 1
}
