package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"mtg-price-checker-sg/controller"
	"mtg-price-checker-sg/pkg/config"
)

type WebResponse struct {
	Data []controller.Card `json:"data"`
}

func Search(_ context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var apiRes events.APIGatewayProxyResponse
	var webRes WebResponse
	var lgs []string

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

	if searchString == "" || len(searchString) < 3 {
		apiRes.StatusCode = http.StatusBadRequest
		return lambdaApiResponse(apiRes, webRes)
	}

	if lgsString != "" {
		lgs = strings.Split(lgsString, ",")
	}

	inStockCards, _ := controller.Search(controller.SearchInput{
		SearchString: searchString,
		Lgs:          lgs,
	})

	if len(inStockCards) > 0 {
		apiRes.StatusCode = http.StatusOK
		webRes.Data = inStockCards
	}

	return lambdaApiResponse(apiRes, webRes)
}

func lambdaApiResponse(apiResponse events.APIGatewayProxyResponse, webResponse WebResponse) (events.APIGatewayProxyResponse, error) {
	bodyBytes, err := json.MarshalIndent(webResponse, "", "    ")
	if err != nil {
		apiResponse.StatusCode = http.StatusInternalServerError
		apiResponse.Body = "err marshalling to json result"
		return apiResponse, nil
	}

	apiResponse.Body = strings.Replace(string(bodyBytes), "\\u0026", "&", -1)

	return apiResponse, nil
}
