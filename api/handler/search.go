package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"mtg-price-checker-sg/controller"
	"mtg-price-checker-sg/controller/ckprice"
	"mtg-price-checker-sg/gateway/cardkingdom"
	"mtg-price-checker-sg/pkg/config"
	"mtg-price-checker-sg/pkg/logger"
	"mtg-price-checker-sg/store/ckprices"

	"github.com/aws/aws-lambda-go/events"
)

type WebResponse struct {
	Data             []controller.Card       `json:"data"`
	Errors           []controller.StoreError `json:"errors"`
	Stats            []controller.StoreStat  `json:"stats"`
	// TotalDurationMs is wall-clock time until the search response is ready,
	// including store fan-out, the minimum response pad, and any Card Kingdom
	// enrichment wait. Per-store Stats[].DurationMs only cover each store's
	// own work and can be much lower than this when enrichment is slow.
	TotalDurationMs  int64                   `json:"totalDurationMs"`
	CardKingdomPrice *cardkingdom.Listing    `json:"cardKingdomPrice,omitempty"`
}

type ErrorResponse struct {
	Error      string `json:"error"`
	StatusCode int    `json:"statusCode"`
}

var searchFunc = controller.Search

var lookupCKPriceFunc = func(ctx context.Context, query string) (*cardkingdom.Listing, error) {
	store, err := ckprices.NewDynamoDBStore(ctx)
	if err != nil {
		return nil, err
	}
	return ckprice.GetLatestPrice(ctx, store, query)
}

func Search(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var apiRes events.APIGatewayProxyResponse
	var webRes WebResponse
	var lgs []string

	// Determine allowed origin for CORS.
	// AWS Lambda proxy integration normalises all headers to lowercase,
	// so only the lowercase "origin" key is needed.
	origin := request.Headers["origin"]

	if request.HTTPMethod == "OPTIONS" {
		return optionsResponse(origin)
	}

	if res, ok := enforceOriginVerify(apiRes, origin, request.Headers); !ok {
		return res, nil
	}
	if res, ok := enforceSession(apiRes, origin, request.Headers); !ok {
		return res, nil
	}

	if config.APIMaintenanceMode() {
		return maintenanceActiveResponse(apiRes, origin)
	}

	searchString, err := url.QueryUnescape(strings.TrimSpace(request.QueryStringParameters["s"]))
	if err != nil {
		searchString = ""
	}
	lgsString, err := url.QueryUnescape(strings.TrimSpace(request.QueryStringParameters["lgs"]))
	if err != nil {
		lgsString = ""
	}

	if os.Getenv("ENV") != config.EnvProd {
		searchString = "Opt"
		lgsString, _ = url.QueryUnescape("Flagship%20Games%2CGames%20Haven%2CGrey%20Ogre%20Games%2CHideout%2CMana%20Pro%2CMox%20%26%20Lotus%2COneMtg%2CSanctuary%20Gaming")
	}

	if searchString == "" || len(searchString) < config.MinSearchStringLength {
		return errorResponse(
			apiRes,
			origin,
			fmt.Sprintf(
				"enter at least %d characters to search",
				config.MinSearchStringLength,
			),
			http.StatusBadRequest,
		)
	}

	if len(searchString) > config.MaxSearchStringLength {
		return errorResponse(
			apiRes,
			origin,
			fmt.Sprintf(
				"card name is too long (maximum %d characters)",
				config.MaxSearchStringLength,
			),
			http.StatusBadRequest,
		)
	}

	if lgsString != "" {
		lgs = strings.Split(lgsString, ",")
	}

	var (
		inStockCards []controller.Card
		storeErrors  []controller.StoreError
		storeStats   []controller.StoreStat
		ckPrice      *cardkingdom.Listing
		searchErr    error
	)

	requestStart := time.Now()
	var wg sync.WaitGroup

	wg.Go(func() {
		inStockCards, storeErrors, storeStats, searchErr = searchFunc(ctx, controller.SearchInput{
			SearchString: searchString,
			Lgs:          lgs,
		})
	})

	if config.CKPriceLookupEnabled() {
		wg.Go(func() {
			ckCtx, cancel := context.WithTimeout(ctx, config.CKPriceLookupTimeout)
			defer cancel()

			price, err := lookupCKPriceFunc(ckCtx, searchString)
			if err != nil {
				logger.From(ckCtx).WarnContext(ckCtx, "ck price lookup failed", "search", searchString, "err", err)
				return
			}
			ckPrice = price
		})
	}

	wg.Wait()
	totalDurationMs := time.Since(requestStart).Milliseconds()

	if searchErr != nil {
		return errorResponse(apiRes, origin, "err searching for cards", http.StatusInternalServerError)
	}

	apiRes.StatusCode = http.StatusOK
	webRes.Data = inStockCards
	if storeErrors == nil {
		webRes.Errors = []controller.StoreError{}
	} else {
		webRes.Errors = storeErrors
	}
	if storeStats == nil {
		webRes.Stats = []controller.StoreStat{}
	} else {
		webRes.Stats = storeStats
	}
	webRes.TotalDurationMs = totalDurationMs
	webRes.CardKingdomPrice = ckPrice

	return searchSuccessResponse(apiRes, webRes, origin)
}

func searchSuccessResponse(apiResponse events.APIGatewayProxyResponse, webResponse WebResponse, origin string) (events.APIGatewayProxyResponse, error) {
	applyCORSHeaders(&apiResponse, origin)
	if apiResponse.StatusCode != http.StatusOK {
		return apiResponse, nil
	}

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(webResponse); err != nil {
		return errorResponse(apiResponse, origin, "err marshalling to json result", http.StatusInternalServerError)
	}

	apiResponse.Body = buf.String()
	return apiResponse, nil
}
