package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strings"

	"mtg-price-checker-sg/controller"
	"mtg-price-checker-sg/pkg/config"

	"github.com/aws/aws-lambda-go/events"
)

type WebResponse struct {
	Data []controller.Card `json:"data"`
}

var searchFunc = controller.Search

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
		lgsString, _ = url.QueryUnescape("Flagship%20Games%2CGames%20Haven%2CGrey%20Ogre%20Games%2CHideout%2CMana%20Pro%2CMox%20%26%20Lotus%2COneMtg%2CSanctuary%20Gaming%2CTefuda")
	}

	if searchString == "" || len(searchString) < 3 {
		apiRes.StatusCode = http.StatusBadRequest
		return lambdaApiResponse(apiRes, webRes)
	}

	if lgsString != "" {
		lgs = strings.Split(lgsString, ",")
	}

	inStockCards, err := searchFunc(controller.SearchInput{
		SearchString: searchString,
		Lgs:          lgs,
	})
	if err != nil {
		apiRes.StatusCode = http.StatusInternalServerError
		apiRes.Body = "err searching for cards"
		return lambdaApiResponse(apiRes, webRes)
	}

	apiRes.StatusCode = http.StatusOK
	webRes.Data = inStockCards

	return lambdaApiResponse(apiRes, webRes)
}

func lambdaApiResponse(apiResponse events.APIGatewayProxyResponse, webResponse WebResponse) (events.APIGatewayProxyResponse, error) {
	if apiResponse.StatusCode != http.StatusOK {
		return apiResponse, nil
	}

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "    ")

	if err := encoder.Encode(webResponse); err != nil {
		apiResponse.StatusCode = http.StatusInternalServerError
		apiResponse.Body = "err marshalling to json result"
		return apiResponse, nil
	}

	apiResponse.Body = buf.String()

	return apiResponse, nil
}
