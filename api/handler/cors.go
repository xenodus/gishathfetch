package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"slices"

	"mtg-price-checker-sg/pkg/config"

	"github.com/aws/aws-lambda-go/events"
)

func applyCORSHeaders(apiResponse *events.APIGatewayProxyResponse, origin string) {
	applyCORSHeadersWithMethods(apiResponse, origin, "GET, OPTIONS")
}

func applyCORSHeadersWithMethods(apiResponse *events.APIGatewayProxyResponse, origin string, methods string) {
	if !slices.Contains(config.GetAllowedOrigins(), origin) {
		return
	}
	apiResponse.Headers = map[string]string{
		"Access-Control-Allow-Origin":  origin,
		"Access-Control-Allow-Methods": methods,
		"Access-Control-Allow-Headers": "Content-Type, Authorization, X-Admin-Api-Key",
		"Vary":                         "Origin",
	}
}

func jsonResponse(
	apiResponse events.APIGatewayProxyResponse,
	origin string,
	statusCode int,
	payload any,
) (events.APIGatewayProxyResponse, error) {
	return jsonResponseWithMethods(apiResponse, origin, statusCode, payload, "GET, OPTIONS")
}

func jsonResponseWithMethods(
	apiResponse events.APIGatewayProxyResponse,
	origin string,
	statusCode int,
	payload any,
	methods string,
) (events.APIGatewayProxyResponse, error) {
	apiResponse.StatusCode = statusCode
	applyCORSHeadersWithMethods(&apiResponse, origin, methods)

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(payload); err != nil {
		return errorResponseWithMethods(apiResponse, origin, "err marshalling response", http.StatusInternalServerError, methods)
	}

	apiResponse.Body = buf.String()
	return apiResponse, nil
}

func errorResponse(
	apiResponse events.APIGatewayProxyResponse,
	origin string,
	message string,
	statusCode int,
) (events.APIGatewayProxyResponse, error) {
	return errorResponseWithMethods(apiResponse, origin, message, statusCode, "GET, OPTIONS")
}

func errorResponseWithMethods(
	apiResponse events.APIGatewayProxyResponse,
	origin string,
	message string,
	statusCode int,
	methods string,
) (events.APIGatewayProxyResponse, error) {
	apiResponse.StatusCode = statusCode
	applyCORSHeadersWithMethods(&apiResponse, origin, methods)

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(ErrorResponse{
		Error:      message,
		StatusCode: statusCode,
	}); err != nil {
		apiResponse.StatusCode = http.StatusInternalServerError
		apiResponse.Body = `{"error":"err marshalling error response","statusCode":500}` + "\n"
		return apiResponse, nil
	}

	apiResponse.Body = buf.String()
	return apiResponse, nil
}

func optionsResponse(origin string) (events.APIGatewayProxyResponse, error) {
	return optionsResponseWithMethods(origin, "GET, OPTIONS")
}

func optionsResponseWithMethods(origin string, methods string) (events.APIGatewayProxyResponse, error) {
	apiResponse := events.APIGatewayProxyResponse{StatusCode: http.StatusNoContent}
	applyCORSHeadersWithMethods(&apiResponse, origin, methods)
	return apiResponse, nil
}
