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

func TestCKPriceRefresh_Success(t *testing.T) {
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
		return 42, nil
	}

	require.NoError(t, os.Setenv(config.CKRefreshAPIKeyEnv, "test-secret"))
	require.NoError(t, os.Setenv("ENV", config.EnvProd))

	result, err := CKPriceRefresh(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodPost,
		Headers:    map[string]string{"x-api-key": "test-secret"},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, result.StatusCode)

	var response CKRefreshResponse
	require.NoError(t, json.Unmarshal([]byte(result.Body), &response))
	require.Equal(t, 42, response.Refreshed)
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

func TestHandle_RoutesCKPriceRefresh(t *testing.T) {
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
	require.Equal(t, http.StatusOK, apiResponse.StatusCode)
}
