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

type mockCKPriceStore struct {
	listing *cardkingdom.Listing
	getErr  error
}

func (m *mockCKPriceStore) GetByNameKey(_ context.Context, _ string) (*cardkingdom.Listing, error) {
	return m.listing, m.getErr
}

func (m *mockCKPriceStore) PutAll(_ context.Context, _ map[string]cardkingdom.Listing) error {
	return nil
}

func TestCKPrice_Success(t *testing.T) {
	originalStoreFunc := newCKPriceStoreFunc
	originalGetFunc := getLatestCKPriceFunc
	defer func() {
		newCKPriceStoreFunc = originalStoreFunc
		getLatestCKPriceFunc = originalGetFunc
	}()

	newCKPriceStoreFunc = func(_ context.Context) (ckprices.Store, error) {
		return &mockCKPriceStore{}, nil
	}
	getLatestCKPriceFunc = func(_ context.Context, _ ckprices.Store, _ string) (*cardkingdom.Listing, error) {
		return &cardkingdom.Listing{
			CardName: "Lightning Bolt",
			Edition:  "Fourth Edition",
			PriceUsd: 1.49,
			URL:      "https://www.cardkingdom.com/mtg/fourth-edition/lightning-bolt",
			Quantity: 12,
			IsFoil:   false,
		}, nil
	}

	require.NoError(t, os.Setenv("ENV", config.EnvProd))

	result, err := CKPrice(context.Background(), events.APIGatewayProxyRequest{
		QueryStringParameters: map[string]string{"s": "Lightning Bolt"},
		Headers:               map[string]string{"origin": "http://localhost:5173"},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, result.StatusCode)

	var response struct {
		Data cardkingdom.Listing `json:"data"`
	}
	require.NoError(t, json.Unmarshal([]byte(result.Body), &response))
	require.Equal(t, "Lightning Bolt", response.Data.CardName)
	require.InDelta(t, 1.49, response.Data.PriceUsd, 0.001)
}

func TestCKPrice_NotFound(t *testing.T) {
	originalStoreFunc := newCKPriceStoreFunc
	originalGetFunc := getLatestCKPriceFunc
	defer func() {
		newCKPriceStoreFunc = originalStoreFunc
		getLatestCKPriceFunc = originalGetFunc
	}()

	newCKPriceStoreFunc = func(_ context.Context) (ckprices.Store, error) {
		return &mockCKPriceStore{}, nil
	}
	getLatestCKPriceFunc = func(_ context.Context, _ ckprices.Store, _ string) (*cardkingdom.Listing, error) {
		return nil, nil
	}

	require.NoError(t, os.Setenv("ENV", config.EnvProd))

	result, err := CKPrice(context.Background(), events.APIGatewayProxyRequest{
		QueryStringParameters: map[string]string{"s": "Not A Real Card"},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, result.StatusCode)
	require.Contains(t, result.Body, `"data":null`)
}

func TestHandle_RoutesCKPrice(t *testing.T) {
	originalStoreFunc := newCKPriceStoreFunc
	originalGetFunc := getLatestCKPriceFunc
	defer func() {
		newCKPriceStoreFunc = originalStoreFunc
		getLatestCKPriceFunc = originalGetFunc
	}()

	newCKPriceStoreFunc = func(_ context.Context) (ckprices.Store, error) {
		return &mockCKPriceStore{}, nil
	}
	getLatestCKPriceFunc = func(_ context.Context, _ ckprices.Store, _ string) (*cardkingdom.Listing, error) {
		return nil, nil
	}

	require.NoError(t, os.Setenv("ENV", config.EnvProd))

	event, err := json.Marshal(events.APIGatewayProxyRequest{
		Path:                  "/ck-price",
		QueryStringParameters: map[string]string{"s": "Lightning Bolt"},
	})
	require.NoError(t, err)

	response, err := Handle(context.Background(), event)
	require.NoError(t, err)

	apiResponse, ok := response.(events.APIGatewayProxyResponse)
	require.True(t, ok)
	require.Equal(t, http.StatusOK, apiResponse.StatusCode)
}
