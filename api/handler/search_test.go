package handler

import (
	"context"
	"net/http"
	"os"
	"testing"

	"mtg-price-checker-sg/pkg/config"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/require"
)

func Test_Search_Success(t *testing.T) {
	type args struct {
		givenAPIGatewayProxyRequest events.APIGatewayProxyRequest
		expResult                   events.APIGatewayProxyResponse
	}
	tcs := map[string]args{
		"success": {
			givenAPIGatewayProxyRequest: events.APIGatewayProxyRequest{
				QueryStringParameters: map[string]string{
					"s":   "abrade",
					"lgs": "Flagship%20Games",
				},
			},
			expResult: events.APIGatewayProxyResponse{
				StatusCode: http.StatusOK,
			},
		},
		"success, no results": {
			givenAPIGatewayProxyRequest: events.APIGatewayProxyRequest{
				QueryStringParameters: map[string]string{
					"s":   "shdjdhjksadjkahdjash",
					"lgs": "Flagship%20Games",
				},
			},
			expResult: events.APIGatewayProxyResponse{
				StatusCode: http.StatusOK,
			},
		},
	}
	for s, tc := range tcs {
		t.Run(s, func(t *testing.T) {
			err := os.Setenv("ENV", config.EnvStaging)
			require.NoError(t, err)
			result, err := Search(context.Background(), tc.givenAPIGatewayProxyRequest)
			require.NoError(t, err)
			require.Equal(t, tc.expResult.StatusCode, result.StatusCode)
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
				Body:       "",
			},
		},
		"less than 3 characters search string": {
			givenAPIGatewayProxyRequest: events.APIGatewayProxyRequest{
				QueryStringParameters: map[string]string{"s": "ab"},
			},
			expResult: events.APIGatewayProxyResponse{
				StatusCode: http.StatusBadRequest,
				Body:       "",
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
