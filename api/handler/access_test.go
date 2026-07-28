package handler

import (
	"context"
	"net/http"
	"os"
	"testing"

	"mtg-price-checker-sg/controller"
	"mtg-price-checker-sg/gateway/cardkingdom"
	"mtg-price-checker-sg/pkg/config"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/require"
)

func Test_normalizeAPIPath(t *testing.T) {
	t.Parallel()
	req := events.APIGatewayProxyRequest{Path: "/api/session"}
	require.Equal(t, "session", normalizeAPIPath(req))

	req = events.APIGatewayProxyRequest{Path: "/api/search"}
	require.Equal(t, "search", normalizeAPIPath(req))

	req = events.APIGatewayProxyRequest{Path: "/prod/api"}
	require.Equal(t, "", normalizeAPIPath(req))
}

func Test_Search_AccessControl(t *testing.T) {
	originalSearchFunc := searchFunc
	originalLookupCKPriceFunc := lookupCKPriceFunc
	defer func() {
		searchFunc = originalSearchFunc
		lookupCKPriceFunc = originalLookupCKPriceFunc
	}()
	searchFunc = func(_ context.Context, _ controller.SearchInput) ([]controller.Card, []controller.StoreError, error) {
		return nil, nil, nil
	}
	lookupCKPriceFunc = func(_ context.Context, _ string) (*cardkingdom.Listing, error) {
		return nil, nil
	}

	require.NoError(t, os.Setenv("ENV", config.EnvProd))
	require.NoError(t, os.Setenv(config.APIOriginVerifySecretEnv, "verify-secret"))
	require.NoError(t, os.Setenv(config.APISessionSecretEnv, "session-secret"))
	t.Cleanup(func() {
		_ = os.Unsetenv(config.APIOriginVerifySecretEnv)
		_ = os.Unsetenv(config.APISessionSecretEnv)
	})

	req := events.APIGatewayProxyRequest{
		QueryStringParameters: map[string]string{"s": "bolt"},
	}
	result, err := Search(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, http.StatusForbidden, result.StatusCode)
}
