package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"

	"mtg-price-checker-sg/controller"
	"mtg-price-checker-sg/pkg/config"

	"github.com/aws/aws-lambda-go/events"
)

type WebResponse struct {
	Data   []controller.Card       `json:"data"`
	Errors []controller.StoreError `json:"errors"`
}

type ValidationErrorResponse struct {
	Error string `json:"error"`
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
		apiRes.StatusCode = http.StatusNoContent
		return lambdaApiResponse(apiRes, webRes, origin)
	}

	searchString, err := url.QueryUnescape(strings.TrimSpace(request.QueryStringParameters["s"]))
	if err != nil {
		searchString = ""
	}
	lgsString, err := url.QueryUnescape(strings.TrimSpace(request.QueryStringParameters["lgs"]))
	if err != nil {
		lgsString = ""
	}

	if os.Getenv("ENV") != config.EnvProd && os.Getenv("ENV") != config.EnvStaging {
		searchString = "Opt"
		lgsString, _ = url.QueryUnescape("Flagship%20Games%2CGames%20Haven%2CGrey%20Ogre%20Games%2CHideout%2CMana%20Pro%2CMox%20%26%20Lotus%2COneMtg%2CSanctuary%20Gaming")
	}

	if searchString == "" || len(searchString) < config.MinSearchStringLength {
		apiRes.StatusCode = http.StatusBadRequest
		return lambdaApiResponse(apiRes, webRes, origin)
	}

	if len(searchString) > config.MaxSearchStringLength {
		apiRes.StatusCode = http.StatusBadRequest
		return validationErrorResponse(
			apiRes,
			origin,
			"card name is too long (maximum 64 characters)",
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
		apiRes.StatusCode = http.StatusInternalServerError
		apiRes.Body = "err searching for cards"
		return lambdaApiResponse(apiRes, webRes, origin)
	}

	apiRes.StatusCode = http.StatusOK
	webRes.Data = inStockCards
	if storeErrors == nil {
		webRes.Errors = []controller.StoreError{}
	} else {
		webRes.Errors = storeErrors
	}

	return lambdaApiResponse(apiRes, webRes, origin)
}

func validationErrorResponse(
	apiResponse events.APIGatewayProxyResponse,
	origin string,
	message string,
) (events.APIGatewayProxyResponse, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(ValidationErrorResponse{Error: message}); err != nil {
		apiResponse.StatusCode = http.StatusInternalServerError
		apiResponse.Body = "err marshalling validation error"
		return lambdaApiResponse(apiResponse, WebResponse{}, origin)
	}

	apiResponse.Body = buf.String()
	return lambdaApiResponse(apiResponse, WebResponse{}, origin)
}

func lambdaApiResponse(apiResponse events.APIGatewayProxyResponse, webResponse WebResponse, origin string) (events.APIGatewayProxyResponse, error) {
	// Set CORS headers.
	// Vary: Origin is required so CDNs/caches don't serve a response cached
	// for one origin to a different origin when the Allow-Origin is dynamic.
	if slices.Contains(config.GetAllowedOrigins(), origin) {
		apiResponse.Headers = map[string]string{
			"Access-Control-Allow-Origin":  origin,
			"Access-Control-Allow-Methods": "GET, OPTIONS",
			"Access-Control-Allow-Headers": "Content-Type",
			"Vary":                         "Origin",
		}
	}

	if apiResponse.StatusCode != http.StatusOK {
		return apiResponse, nil
	}

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(webResponse); err != nil {
		apiResponse.StatusCode = http.StatusInternalServerError
		apiResponse.Body = "err marshalling to json result"
		return apiResponse, nil
	}

	apiResponse.Body = buf.String()

	return apiResponse, nil
}
