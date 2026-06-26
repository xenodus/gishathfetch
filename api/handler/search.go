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

	"mtg-price-checker-sg/controller"
	"mtg-price-checker-sg/pkg/config"

	"github.com/aws/aws-lambda-go/events"
)

type WebResponse struct {
	Data   []controller.Card       `json:"data"`
	Errors []controller.StoreError `json:"errors"`
}

type ErrorResponse struct {
	Error      string `json:"error"`
	StatusCode int    `json:"statusCode"`
}

var searchFunc = controller.Search

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

	inStockCards, storeErrors, err := searchFunc(ctx, controller.SearchInput{
		SearchString: searchString,
		Lgs:          lgs,
	})
	if err != nil {
		return errorResponse(apiRes, origin, "err searching for cards", http.StatusInternalServerError)
	}

	apiRes.StatusCode = http.StatusOK
	webRes.Data = inStockCards
	if storeErrors == nil {
		webRes.Errors = []controller.StoreError{}
	} else {
		webRes.Errors = storeErrors
	}

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
