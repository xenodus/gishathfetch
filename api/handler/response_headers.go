package handler

import (
	"github.com/aws/aws-lambda-go/events"
)

// ensureResponseHeaders returns a non-nil Headers map on apiResponse.
func ensureResponseHeaders(apiResponse *events.APIGatewayProxyResponse) map[string]string {
	if apiResponse.Headers == nil {
		apiResponse.Headers = map[string]string{}
	}
	return apiResponse.Headers
}
