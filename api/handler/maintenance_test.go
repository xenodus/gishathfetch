package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"mtg-price-checker-sg/pkg/config"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/require"
)

func TestSearch_MaintenanceMode(t *testing.T) {
	t.Setenv(config.APIMaintenanceModeEnv, "true")
	t.Setenv(config.APIMaintenanceMessageEnv, "Scheduled maintenance in progress.")

	req := events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodGet,
		QueryStringParameters: map[string]string{
			"s": "abrade",
		},
	}

	result, err := Search(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, http.StatusServiceUnavailable, result.StatusCode)

	var body ErrorResponse
	require.NoError(t, json.Unmarshal([]byte(result.Body), &body))
	require.Equal(t, "Scheduled maintenance in progress.", body.Error)
	require.Equal(t, http.StatusServiceUnavailable, body.StatusCode)
}

func TestSession_AdvertisesMaintenanceHeaders(t *testing.T) {
	t.Setenv(config.APISessionSecretEnv, "test-session-secret")
	t.Setenv(config.APIMaintenanceModeEnv, "true")
	t.Setenv(config.APIMaintenanceMessageEnv, "Back soon.")

	req := events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodGet,
		Headers: map[string]string{
			"origin": "http://localhost:5173",
		},
	}

	result, err := Session(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, result.StatusCode)
	require.Equal(t, "1", result.Headers[maintenanceModeHeader])
	require.Equal(t, "Back soon.", result.Headers[maintenanceMessageHeader])

	var body SiteStatusResponse
	require.NoError(t, json.Unmarshal([]byte(result.Body), &body))
	require.True(t, body.MaintenanceMode)
	require.Equal(t, "Back soon.", body.MaintenanceMessage)
}

func TestSession_AdvertisesNoticeHeader(t *testing.T) {
	t.Setenv(config.APISessionSecretEnv, "test-session-secret")
	t.Setenv(config.APINoticeMessageEnv, "Card Kingdom prices may be delayed today.")

	req := events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodGet,
		Headers: map[string]string{
			"origin": "http://localhost:5173",
		},
	}

	result, err := Session(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, result.StatusCode)
	require.Equal(t, "Card Kingdom prices may be delayed today.", result.Headers[noticeMessageHeader])
	require.Empty(t, result.Headers[maintenanceModeHeader])

	var body SiteStatusResponse
	require.NoError(t, json.Unmarshal([]byte(result.Body), &body))
	require.False(t, body.MaintenanceMode)
	require.Equal(t, "Card Kingdom prices may be delayed today.", body.NoticeMessage)
}
