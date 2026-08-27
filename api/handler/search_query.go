package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"mtg-price-checker-sg/pkg/config"

	"github.com/aws/aws-lambda-go/events"
)

type parsedSearchQuery struct {
	searchString string
	lgs          []string
}

func parseSearchQuery(request events.APIGatewayProxyRequest) parsedSearchQuery {
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

	var lgs []string
	if lgsString != "" {
		lgs = strings.Split(lgsString, ",")
	}

	return parsedSearchQuery{
		searchString: searchString,
		lgs:          lgs,
	}
}

func validateSearchString(searchString string) (string, int) {
	if searchString == "" || len(searchString) < config.MinSearchStringLength {
		return fmt.Sprintf(
			"enter at least %d characters to search",
			config.MinSearchStringLength,
		), http.StatusBadRequest
	}
	if len(searchString) > config.MaxSearchStringLength {
		return fmt.Sprintf(
			"card name is too long (maximum %d characters)",
			config.MaxSearchStringLength,
		), http.StatusBadRequest
	}
	return "", 0
}

func buildWebsiteSearchURL(searchString string, lgs []string) string {
	base := strings.TrimSuffix(config.SiteBaseURL, "/")
	params := url.Values{}
	params.Set("s", searchString)
	if len(lgs) > 0 {
		params.Set("lgs", strings.Join(lgs, ","))
	}
	params.Set("utm_source", "telegram")
	params.Set("utm_medium", "bot")
	params.Set("utm_campaign", "price_lookup")
	return base + "/?" + params.Encode()
}
