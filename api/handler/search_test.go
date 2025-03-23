package handler

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/require"
	"mtg-price-checker-sg/pkg/config"
)

func Test_Search_Success(t *testing.T) {
	type args struct {
		givenAPIGatewayProxyRequest events.APIGatewayProxyRequest
	}
	tcs := map[string]args{
		"success": {
			givenAPIGatewayProxyRequest: events.APIGatewayProxyRequest{
				QueryStringParameters: map[string]string{
					"s":   "abrade",
					"lgs": "Flagship%20Games",
				},
			},
		},
	}
	for s, tc := range tcs {
		t.Run(s, func(t *testing.T) {
			result, err := Search(context.Background(), tc.givenAPIGatewayProxyRequest)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, result.StatusCode)
		})
	}
}

func Test_Search_Err(t *testing.T) {
	type args struct {
		givenAPIGatewayProxyRequest events.APIGatewayProxyRequest
		expResult                   events.APIGatewayProxyResponse
	}
	tcs := map[string]args{
		"empty search string": {
			givenAPIGatewayProxyRequest: events.APIGatewayProxyRequest{
				QueryStringParameters: map[string]string{"s": ""},
			},
			expResult: events.APIGatewayProxyResponse{
				StatusCode: http.StatusBadRequest,
				Body:       "{\n    \"data\": null\n}",
			},
		},
		"less than 3 characters search string": {
			givenAPIGatewayProxyRequest: events.APIGatewayProxyRequest{
				QueryStringParameters: map[string]string{"s": "ab"},
			},
			expResult: events.APIGatewayProxyResponse{
				StatusCode: http.StatusBadRequest,
				Body:       "{\n    \"data\": null\n}",
			},
		},
	}
	for s, tc := range tcs {
		t.Run(s, func(t *testing.T) {
			err := os.Setenv("ENV", config.EnvStaging)
			require.NoError(t, err)
			result, err := Search(context.Background(), tc.givenAPIGatewayProxyRequest)
			require.NoError(t, err)
			require.Equal(t, tc.expResult, result)
		})
	}
}
