package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"mtg-price-checker-sg/gateway/cardkingdom"
	"mtg-price-checker-sg/pkg/config"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/require"
)

func TestTelegramCK_Success(t *testing.T) {
	originalLookupCKPriceFunc := lookupCKPriceFunc
	defer func() {
		lookupCKPriceFunc = originalLookupCKPriceFunc
	}()

	inStock := true
	lookupCKPriceFunc = func(_ context.Context, query string) (*cardkingdom.Listing, error) {
		require.Equal(t, "lightning bolt", query)
		return &cardkingdom.Listing{
			CardName: "Lightning Bolt",
			Edition:  "Fourth Edition",
			PriceUsd: 0.49,
			URL:      "https://www.cardkingdom.com/mtg/fourth-edition/lightning-bolt",
			InStock:  &inStock,
		}, nil
	}

	require.NoError(t, os.Setenv("ENV", config.EnvProd))
	require.NoError(t, os.Setenv(config.APITelegramBotTokenEnv, "bot-token"))
	t.Cleanup(func() {
		_ = os.Unsetenv(config.APITelegramBotTokenEnv)
	})

	result, err := TelegramCK(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod:            http.MethodGet,
		QueryStringParameters: map[string]string{"s": "lightning bolt"},
		Headers:               map[string]string{"authorization": "Bearer bot-token"},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, result.StatusCode)

	var body TelegramCKResponse
	require.NoError(t, json.Unmarshal([]byte(result.Body), &body))
	require.NotNil(t, body.Listing)
	require.Equal(t, "Lightning Bolt", body.Listing.CardName)
	require.Equal(t, 0.49, body.Listing.PriceUsd)
	require.GreaterOrEqual(t, body.DurationMs, int64(0))
}

func TestTelegramCK_NoListing(t *testing.T) {
	originalLookupCKPriceFunc := lookupCKPriceFunc
	defer func() {
		lookupCKPriceFunc = originalLookupCKPriceFunc
	}()

	lookupCKPriceFunc = func(_ context.Context, _ string) (*cardkingdom.Listing, error) {
		return nil, nil
	}

	require.NoError(t, os.Setenv("ENV", config.EnvProd))
	require.NoError(t, os.Setenv(config.APITelegramBotTokenEnv, "bot-token"))
	t.Cleanup(func() {
		_ = os.Unsetenv(config.APITelegramBotTokenEnv)
	})

	result, err := TelegramCK(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod:            http.MethodGet,
		QueryStringParameters: map[string]string{"s": "nothing"},
		Headers:               map[string]string{"authorization": "Bearer bot-token"},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, result.StatusCode)

	var body TelegramCKResponse
	require.NoError(t, json.Unmarshal([]byte(result.Body), &body))
	require.Nil(t, body.Listing)
}

func TestTelegramCK_AccessControl(t *testing.T) {
	originalLookupCKPriceFunc := lookupCKPriceFunc
	defer func() {
		lookupCKPriceFunc = originalLookupCKPriceFunc
	}()

	lookupCKPriceFunc = func(_ context.Context, _ string) (*cardkingdom.Listing, error) {
		return nil, nil
	}

	require.NoError(t, os.Setenv("ENV", config.EnvProd))
	require.NoError(t, os.Setenv(config.APIOriginVerifySecretEnv, "verify-secret"))
	require.NoError(t, os.Setenv(config.APITelegramBotTokenEnv, "bot-token"))
	t.Cleanup(func() {
		_ = os.Unsetenv(config.APIOriginVerifySecretEnv)
		_ = os.Unsetenv(config.APITelegramBotTokenEnv)
	})

	t.Run("missing bot token", func(t *testing.T) {
		result, err := TelegramCK(context.Background(), events.APIGatewayProxyRequest{
			HTTPMethod:            http.MethodGet,
			QueryStringParameters: map[string]string{"s": "bolt"},
			Headers: map[string]string{
				"x-origin-verify": "verify-secret",
			},
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusForbidden, result.StatusCode)
	})

	t.Run("wrong bot token", func(t *testing.T) {
		result, err := TelegramCK(context.Background(), events.APIGatewayProxyRequest{
			HTTPMethod:            http.MethodGet,
			QueryStringParameters: map[string]string{"s": "bolt"},
			Headers: map[string]string{
				"x-origin-verify": "verify-secret",
				"authorization":   "Bearer wrong",
			},
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusForbidden, result.StatusCode)
	})

	t.Run("not configured", func(t *testing.T) {
		_ = os.Unsetenv(config.APITelegramBotTokenEnv)
		result, err := TelegramCK(context.Background(), events.APIGatewayProxyRequest{
			HTTPMethod:            http.MethodGet,
			QueryStringParameters: map[string]string{"s": "bolt"},
			Headers: map[string]string{
				"x-origin-verify": "verify-secret",
				"authorization":   "Bearer bot-token",
			},
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusServiceUnavailable, result.StatusCode)
	})
}
