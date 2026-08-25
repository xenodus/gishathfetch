package handler

import (
	"net/http"

	"mtg-price-checker-sg/pkg/config"

	"github.com/aws/aws-lambda-go/events"
)

const (
	maintenanceModeHeader    = "X-Maintenance-Mode"
	maintenanceMessageHeader = "X-Maintenance-Message"
	noticeMessageHeader      = "X-Notice-Message"
)

// SiteStatusResponse is returned in GET /session JSON so browsers can read site
// status without relying on Access-Control-Expose-Headers for custom headers.
type SiteStatusResponse struct {
	MaintenanceMode    bool   `json:"maintenanceMode"`
	MaintenanceMessage string `json:"maintenanceMessage,omitempty"`
	NoticeMessage      string `json:"noticeMessage,omitempty"`
}

func buildSiteStatusResponse() SiteStatusResponse {
	status := SiteStatusResponse{}
	if config.APIMaintenanceMode() {
		status.MaintenanceMode = true
		status.MaintenanceMessage = config.APIMaintenanceMessage()
	}
	if noticeMessage := config.APINoticeMessage(); noticeMessage != "" {
		status.NoticeMessage = noticeMessage
	}
	return status
}

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
	headers := ensureResponseHeaders(apiRes)
	if config.APIMaintenanceMode() {
		headers[maintenanceModeHeader] = "1"
		headers[maintenanceMessageHeader] = config.APIMaintenanceMessage()
	}
	if noticeMessage := config.APINoticeMessage(); noticeMessage != "" {
		headers[noticeMessageHeader] = noticeMessage
	}
}
