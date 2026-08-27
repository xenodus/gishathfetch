package handler

import (
	"encoding/json"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

const apiGatewayHTTPAPIVersion = "2.0"

func parseAPIRequest(event []byte) (events.APIGatewayProxyRequest, error) {
	var versionProbe struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(event, &versionProbe); err != nil {
		return events.APIGatewayProxyRequest{}, err
	}

	if versionProbe.Version == apiGatewayHTTPAPIVersion {
		var req events.APIGatewayV2HTTPRequest
		if err := json.Unmarshal(event, &req); err != nil {
			return events.APIGatewayProxyRequest{}, err
		}
		return apigwV2ToProxyRequest(req), nil
	}

	var req events.APIGatewayProxyRequest
	if err := json.Unmarshal(event, &req); err != nil {
		return events.APIGatewayProxyRequest{}, err
	}
	return req, nil
}

func apigwV2ToProxyRequest(req events.APIGatewayV2HTTPRequest) events.APIGatewayProxyRequest {
	path := strings.TrimSpace(req.RawPath)
	if path == "" {
		path = strings.TrimSpace(req.RequestContext.HTTP.Path)
	}

	query := strings.TrimSpace(req.RawQueryString)
	if query != "" && req.QueryStringParameters == nil {
		req.QueryStringParameters = map[string]string{}
		for part := range strings.SplitSeq(query, "&") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			key, value, ok := strings.Cut(part, "=")
			if !ok {
				req.QueryStringParameters[key] = ""
				continue
			}
			req.QueryStringParameters[key] = value
		}
	}

	return events.APIGatewayProxyRequest{
		Resource:                        req.RouteKey,
		Path:                            path,
		HTTPMethod:                      req.RequestContext.HTTP.Method,
		Headers:                         req.Headers,
		QueryStringParameters:           req.QueryStringParameters,
		PathParameters:                  req.PathParameters,
		StageVariables:                  req.StageVariables,
		Body:                            req.Body,
		IsBase64Encoded:                 req.IsBase64Encoded,
		MultiValueHeaders:               map[string][]string{},
		MultiValueQueryStringParameters: map[string][]string{},
		RequestContext: events.APIGatewayProxyRequestContext{
			Path: path,
		},
	}
}
