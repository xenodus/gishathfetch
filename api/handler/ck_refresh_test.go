package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"mtg-price-checker-sg/gateway/cardkingdom"
	"mtg-price-checker-sg/pkg/config"
	"mtg-price-checker-sg/store/ckprices"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/require"
)

type mockCKRefreshStore struct{}

func (m *mockCKRefreshStore) GetByNameKey(_ context.Context, _ string) (*cardkingdom.Listing, error) {
	return nil, nil
}

func (m *mockCKRefreshStore) PutAll(_ context.Context, _ map[string]cardkingdom.Listing) error {
	return nil
}

func TestCKPriceRefresh_Accepted(t *testing.T) {
	originalEnqueueFunc := enqueueCKPriceRefreshFunc
	defer func() { enqueueCKPriceRefreshFunc = originalEnqueueFunc }()

	enqueueCKPriceRefreshFunc = func(_ context.Context) error {
		return nil
	}

	require.NoError(t, os.Setenv(config.CKRefreshAPIKeyEnv, "test-secret"))
	require.NoError(t, os.Setenv("ENV", config.EnvProd))

	result, err := CKPriceRefresh(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodPost,
		Headers:    map[string]string{"x-api-key": "test-secret"},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusAccepted, result.StatusCode)

	var response CKRefreshAcceptedResponse
	require.NoError(t, json.Unmarshal([]byte(result.Body), &response))
	require.Equal(t, "accepted", response.Status)
}

func TestCKPriceRefresh_Unauthorized(t *testing.T) {
	require.NoError(t, os.Setenv(config.CKRefreshAPIKeyEnv, "test-secret"))
	require.NoError(t, os.Setenv("ENV", config.EnvProd))

	result, err := CKPriceRefresh(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodPost,
		Headers:    map[string]string{"x-api-key": "wrong"},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, result.StatusCode)
}

func TestHandle_RoutesCKPriceRefreshRun(t *testing.T) {
	originalStoreFunc := newCKRefreshStoreFunc
	originalRefreshFunc := refreshCKPricesFunc
	defer func() {
		newCKRefreshStoreFunc = originalStoreFunc
		refreshCKPricesFunc = originalRefreshFunc
	}()

	newCKRefreshStoreFunc = func(_ context.Context) (ckprices.Store, error) {
		return &mockCKRefreshStore{}, nil
	}
	refreshCKPricesFunc = func(_ context.Context, _ ckprices.Store) (int, error) {
		return 1, nil
	}

	event, err := json.Marshal(map[string]string{"action": ckPriceRefreshRunAction})
	require.NoError(t, err)

	_, err = Handle(context.Background(), event)
	require.NoError(t, err)
}

func TestHandle_RoutesCKPriceRefresh(t *testing.T) {
	originalEnqueueFunc := enqueueCKPriceRefreshFunc
	defer func() { enqueueCKPriceRefreshFunc = originalEnqueueFunc }()

	enqueueCKPriceRefreshFunc = func(_ context.Context) error {
		return nil
	}

	require.NoError(t, os.Setenv(config.CKRefreshAPIKeyEnv, "test-secret"))
	require.NoError(t, os.Setenv("ENV", config.EnvProd))

	event, err := json.Marshal(events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodPost,
		Path:       "/ck-price/refresh",
		Headers:    map[string]string{"x-api-key": "test-secret"},
	})
	require.NoError(t, err)

	response, err := Handle(context.Background(), event)
	require.NoError(t, err)

	apiResponse, ok := response.(events.APIGatewayProxyResponse)
	require.True(t, ok)
	require.Equal(t, http.StatusAccepted, apiResponse.StatusCode)
}
