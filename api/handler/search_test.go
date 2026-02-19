package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"testing"

	"mtg-price-checker-sg/controller"
	"mtg-price-checker-sg/pkg/config"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/require"
)

func Test_Search_Success(t *testing.T) {
	type args struct {
		givenAPIGatewayProxyRequest events.APIGatewayProxyRequest
		mockSearchResponse          []controller.Card
		mockSearchErr               error
		expStatusCode               int
		expBodyData                 []controller.Card
	}
	tcs := map[string]args{
		"success with results": {
			givenAPIGatewayProxyRequest: events.APIGatewayProxyRequest{
				QueryStringParameters: map[string]string{
					"s":   "abrade",
					"lgs": "Flagship%20Games",
				},
			},
			mockSearchResponse: []controller.Card{
				{Name: "Abrade", Price: 1.5, Source: "Flagship Games", InStock: true},
			},
			mockSearchErr: nil,
			expStatusCode: http.StatusOK,
			expBodyData: []controller.Card{
				{Name: "Abrade", Price: 1.5, Source: "Flagship Games", InStock: true},
			},
		},
		"success, no results": {
			givenAPIGatewayProxyRequest: events.APIGatewayProxyRequest{
				QueryStringParameters: map[string]string{
					"s":   "shdjdhjksadjkahdjash",
					"lgs": "Flagship%20Games",
				},
			},
			mockSearchResponse: nil,
			mockSearchErr:      nil,
			expStatusCode:      http.StatusOK,
			expBodyData:        nil, // key: "data": null
		},
	}
	for s, tc := range tcs {
		t.Run(s, func(t *testing.T) {
			// Setup Mock
			originalSearchFunc := searchFunc
			defer func() { searchFunc = originalSearchFunc }()
			searchFunc = func(input controller.SearchInput) ([]controller.Card, error) {
				return tc.mockSearchResponse, tc.mockSearchErr
			}

			err := os.Setenv("ENV", config.EnvStaging)
			require.NoError(t, err)

			result, err := Search(context.Background(), tc.givenAPIGatewayProxyRequest)
			require.NoError(t, err)
			require.Equal(t, tc.expStatusCode, result.StatusCode)

			// Verify Body
			var webRes WebResponse
			err = json.Unmarshal([]byte(result.Body), &webRes)
			require.NoError(t, err)
			require.Equal(t, tc.expBodyData, webRes.Data)
		})
	}
}

func Test_Search_CORS(t *testing.T) {
	// Setup Mock
	originalSearchFunc := searchFunc
	defer func() { searchFunc = originalSearchFunc }()
	searchFunc = func(input controller.SearchInput) ([]controller.Card, error) {
		return []controller.Card{}, nil
	}

	err := os.Setenv("ENV", config.EnvStaging)
	require.NoError(t, err)

	t.Run("allowed origin returns CORS headers", func(t *testing.T) {
		req := events.APIGatewayProxyRequest{
			QueryStringParameters: map[string]string{"s": "test"},
			// Lambda proxy normalises headers to lowercase.
			Headers: map[string]string{"origin": "http://localhost:5173"},
		}
		result, err := Search(context.Background(), req)
		require.NoError(t, err)
		require.Equal(t, "http://localhost:5173", result.Headers["Access-Control-Allow-Origin"])
		require.Equal(t, "Origin", result.Headers["Vary"])
	})

	t.Run("OPTIONS request returns 204 and CORS headers", func(t *testing.T) {
		// No query params needed: OPTIONS is short-circuited before query parsing.
		req := events.APIGatewayProxyRequest{
			HTTPMethod: "OPTIONS",
			// Lambda proxy normalises headers to lowercase.
			Headers: map[string]string{"origin": "http://localhost:5173"},
		}
		result, err := Search(context.Background(), req)
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, result.StatusCode)
		require.Equal(t, "http://localhost:5173", result.Headers["Access-Control-Allow-Origin"])
		require.Equal(t, "Origin", result.Headers["Vary"])
	})

	t.Run("disallowed origin does not return CORS headers", func(t *testing.T) {
		req := events.APIGatewayProxyRequest{
			QueryStringParameters: map[string]string{"s": "test"},
			Headers:               map[string]string{"Origin": "http://malicious.com"},
		}
		result, err := Search(context.Background(), req)
		require.NoError(t, err)
		_, ok := result.Headers["Access-Control-Allow-Origin"]
		require.False(t, ok)
	})
}

func Test_Search_Err(t *testing.T) {
	type args struct {
		givenAPIGatewayProxyRequest events.APIGatewayProxyRequest
		mockSearchResponse          []controller.Card
		mockSearchErr               error
		expStatusCode               int
		expBody                     string
	}
	tcs := map[string]args{
		"empty search string": {
			givenAPIGatewayProxyRequest: events.APIGatewayProxyRequest{
				QueryStringParameters: map[string]string{"s": ""},
			},
			expStatusCode: http.StatusBadRequest,
			expBody:       "", // lambdaApiResponse returns body as is for error but empty webRes
		},
		"less than 3 characters search string": {
			givenAPIGatewayProxyRequest: events.APIGatewayProxyRequest{
				QueryStringParameters: map[string]string{"s": "ab"},
			},
			expStatusCode: http.StatusBadRequest,
			expBody:       "",
		},
		"controller error": {
			givenAPIGatewayProxyRequest: events.APIGatewayProxyRequest{
				QueryStringParameters: map[string]string{"s": "valid"},
			},
			mockSearchErr: errors.New("controller error"),
			expStatusCode: http.StatusInternalServerError,
			expBody:       "err searching for cards",
		},
	}
	for s, tc := range tcs {
		t.Run(s, func(t *testing.T) {
			// Setup Mock
			originalSearchFunc := searchFunc
			defer func() { searchFunc = originalSearchFunc }()
			searchFunc = func(input controller.SearchInput) ([]controller.Card, error) {
				return tc.mockSearchResponse, tc.mockSearchErr
			}

			err := os.Setenv("ENV", config.EnvStaging)
			require.NoError(t, err)
			result, err := Search(context.Background(), tc.givenAPIGatewayProxyRequest)
			require.NoError(t, err)
			require.Equal(t, tc.expStatusCode, result.StatusCode)

			if tc.expBody != "" {
				require.Equal(t, tc.expBody, result.Body)
			}
		})
	}
}
