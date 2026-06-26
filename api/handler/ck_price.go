package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"mtg-price-checker-sg/controller/ckprice"
	"mtg-price-checker-sg/pkg/config"
	"mtg-price-checker-sg/store/ckprices"

	"github.com/aws/aws-lambda-go/events"
)

type CKPriceResponse struct {
	Data any `json:"data"`
}

var (
	newCKPriceStoreFunc = func(ctx context.Context) (ckprices.Store, error) {
		return ckprices.NewDynamoDBStore(ctx)
	}
	getLatestCKPriceFunc = ckprice.GetLatestPrice
)

func CKPrice(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var apiRes events.APIGatewayProxyResponse
	origin := request.Headers["origin"]

	if request.HTTPMethod == http.MethodOptions {
		return optionsResponse(origin)
	}

	searchString, err := url.QueryUnescape(strings.TrimSpace(request.QueryStringParameters["s"]))
	if err != nil {
		searchString = ""
	}

	if searchString == "" || len(searchString) < config.MinSearchStringLength {
		return errorResponse(
			apiRes,
			origin,
			fmt.Sprintf("enter at least %d characters to search", config.MinSearchStringLength),
			http.StatusBadRequest,
		)
	}

	if len(searchString) > config.MaxSearchStringLength {
		return errorResponse(
			apiRes,
			origin,
			fmt.Sprintf("card name is too long (maximum %d characters)", config.MaxSearchStringLength),
			http.StatusBadRequest,
		)
	}

	store, err := newCKPriceStoreFunc(ctx)
	if err != nil {
		return errorResponse(apiRes, origin, "card kingdom price lookup is unavailable", http.StatusInternalServerError)
	}

	listing, err := getLatestCKPriceFunc(ctx, store, searchString)
	if err != nil {
		return errorResponse(apiRes, origin, "err looking up card kingdom price", http.StatusInternalServerError)
	}

	return jsonResponse(apiRes, origin, http.StatusOK, CKPriceResponse{Data: listing})
}
