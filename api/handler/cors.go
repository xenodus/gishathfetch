package handler

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

func applyCORSHeaders(apiResponse *events.APIGatewayProxyResponse, origin string) {
	if !isAllowedOrigin(origin) {
		return
	}
	apiResponse.Headers = map[string]string{
		"Access-Control-Allow-Origin":  origin,
		"Access-Control-Allow-Methods": "GET, OPTIONS",
		"Access-Control-Allow-Headers": "Content-Type",
		"Vary":                         "Origin",
	}
}

func jsonResponse(
	apiResponse events.APIGatewayProxyResponse,
	origin string,
	statusCode int,
	payload any,
) (events.APIGatewayProxyResponse, error) {
	apiResponse.StatusCode = statusCode
	applyCORSHeaders(&apiResponse, origin)

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(payload); err != nil {
		return errorResponse(apiResponse, origin, "err marshalling response", http.StatusInternalServerError)
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
	apiResponse.StatusCode = statusCode
	applyCORSHeaders(&apiResponse, origin)

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
	apiResponse := events.APIGatewayProxyResponse{StatusCode: http.StatusNoContent}
	applyCORSHeaders(&apiResponse, origin)
	return apiResponse, nil
}
