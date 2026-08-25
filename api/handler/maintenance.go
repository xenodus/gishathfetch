package handler

import (
	"net/http"

	"mtg-price-checker-sg/pkg/config"

	"github.com/aws/aws-lambda-go/events"
)

const (
	maintenanceModeHeader   = "X-Maintenance-Mode"
	maintenanceMessageHeader = "X-Maintenance-Message"
)

func maintenanceActiveResponse(
	apiRes events.APIGatewayProxyResponse,
	origin string,
) (events.APIGatewayProxyResponse, error) {
	return errorResponse(
		apiRes,
		origin,
		config.APIMaintenanceMessage(),
		http.StatusServiceUnavailable,
	)
}

func applyMaintenanceHeaders(apiRes *events.APIGatewayProxyResponse) {
	if !config.APIMaintenanceMode() {
		return
	}
	headers := ensureResponseHeaders(apiRes)
	headers[maintenanceModeHeader] = "1"
	headers[maintenanceMessageHeader] = config.APIMaintenanceMessage()
}
