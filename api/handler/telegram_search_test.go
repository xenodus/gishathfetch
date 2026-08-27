package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"mtg-price-checker-sg/controller"
	"mtg-price-checker-sg/pkg/config"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/require"
)

func TestTelegramSearch_Success(t *testing.T) {
	originalSearchFunc := searchFunc
	defer func() {
		searchFunc = originalSearchFunc
	}()
	searchFunc = func(_ context.Context, input controller.SearchInput) ([]controller.Card, []controller.StoreError, []controller.StoreStat, error) {
		require.Equal(t, "abrade", input.SearchString)
		return []controller.Card{
			{Name: "Abrade", Price: 1.5, Source: "Flagship Games", InStock: true, Url: "https://shop.example/abrade"},
			{Name: "Abrade", Price: 2.0, Source: "Hideout", InStock: true},
		}, []controller.StoreError{}, []controller.StoreStat{
			{Store: "Flagship Games", ItemCount: 1, DurationMs: 100},
		}, nil
	}

	require.NoError(t, os.Setenv("ENV", config.EnvProd))
	require.NoError(t, os.Setenv(config.APITelegramBotTokenEnv, "bot-token"))
	t.Cleanup(func() {
		_ = os.Unsetenv(config.APITelegramBotTokenEnv)
	})

	req := events.APIGatewayProxyRequest{
		HTTPMethod:            http.MethodGet,
		QueryStringParameters: map[string]string{"s": "abrade"},
		Headers: map[string]string{
			"authorization": "Bearer bot-token",
		},
	}

	result, err := TelegramSearch(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, result.StatusCode)

	var body TelegramSearchResponse
	require.NoError(t, json.Unmarshal([]byte(result.Body), &body))
	require.Equal(t, 2, body.ResultCount)
	require.NotNil(t, body.Cheapest)
	require.Equal(t, "Abrade", body.Cheapest.Name)
	require.Equal(t, 1.5, body.Cheapest.Price)
	require.Equal(t, "Flagship Games", body.Cheapest.Source)
	require.Contains(t, body.WebsiteURL, "s=abrade")
	require.Contains(t, body.WebsiteURL, "utm_source=telegram")
	require.GreaterOrEqual(t, body.TotalDurationMs, int64(0))
}

func TestTelegramSearch_NoResults(t *testing.T) {
	originalSearchFunc := searchFunc
	defer func() {
		searchFunc = originalSearchFunc
	}()
	searchFunc = func(_ context.Context, _ controller.SearchInput) ([]controller.Card, []controller.StoreError, []controller.StoreStat, error) {
		return nil, nil, nil, nil
	}

	require.NoError(t, os.Setenv("ENV", config.EnvProd))
	require.NoError(t, os.Setenv(config.APITelegramBotTokenEnv, "bot-token"))
	t.Cleanup(func() {
		_ = os.Unsetenv(config.APITelegramBotTokenEnv)
	})

	result, err := TelegramSearch(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod:            http.MethodGet,
		QueryStringParameters: map[string]string{"s": "nothing"},
		Headers:               map[string]string{"authorization": "Bearer bot-token"},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, result.StatusCode)

	var body TelegramSearchResponse
	require.NoError(t, json.Unmarshal([]byte(result.Body), &body))
	require.Equal(t, 0, body.ResultCount)
	require.Nil(t, body.Cheapest)
}

func TestTelegramSearch_AccessControl(t *testing.T) {
	originalSearchFunc := searchFunc
	defer func() {
		searchFunc = originalSearchFunc
	}()
	searchFunc = func(_ context.Context, _ controller.SearchInput) ([]controller.Card, []controller.StoreError, []controller.StoreStat, error) {
		return nil, nil, nil, nil
	}

	require.NoError(t, os.Setenv("ENV", config.EnvProd))
	require.NoError(t, os.Setenv(config.APIOriginVerifySecretEnv, "verify-secret"))
	require.NoError(t, os.Setenv(config.APITelegramBotTokenEnv, "bot-token"))
	t.Cleanup(func() {
		_ = os.Unsetenv(config.APIOriginVerifySecretEnv)
		_ = os.Unsetenv(config.APITelegramBotTokenEnv)
	})

	t.Run("missing bot token", func(t *testing.T) {
		result, err := TelegramSearch(context.Background(), events.APIGatewayProxyRequest{
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
		result, err := TelegramSearch(context.Background(), events.APIGatewayProxyRequest{
			HTTPMethod:            http.MethodGet,
			QueryStringParameters: map[string]string{"s": "bolt"},
			Headers: map[string]string{
				"x-origin-verify": "verify-secret",
				"authorization":                "Bearer wrong",
			},
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusForbidden, result.StatusCode)
	})

	t.Run("not configured", func(t *testing.T) {
		_ = os.Unsetenv(config.APITelegramBotTokenEnv)
		result, err := TelegramSearch(context.Background(), events.APIGatewayProxyRequest{
			HTTPMethod:            http.MethodGet,
			QueryStringParameters: map[string]string{"s": "bolt"},
			Headers: map[string]string{
				"x-origin-verify": "verify-secret",
				"authorization":                "Bearer bot-token",
			},
		})
		require.NoError(t, err)
		require.Equal(t, http.StatusServiceUnavailable, result.StatusCode)
	})
}

func Test_buildWebsiteSearchURL(t *testing.T) {
	url := buildWebsiteSearchURL("Lightning Bolt", []string{"Flagship Games", "Hideout"})
	require.Contains(t, url, "s=Lightning+Bolt")
	require.Contains(t, url, "lgs=Flagship+Games")
	require.Contains(t, url, "utm_campaign=price_lookup")
}
