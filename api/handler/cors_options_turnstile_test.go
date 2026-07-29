package handler

import (
	"context"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/require"
)

func TestSession_OptionsIncludesTurnstileCORSHeader(t *testing.T) {
	res, err := Session(context.Background(), events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodOptions,
		Headers:    map[string]string{"origin": "https://gishathfetch.com"},
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, res.StatusCode)
	require.Equal(t, "https://gishathfetch.com", res.Headers["Access-Control-Allow-Origin"])
	require.Equal(t, "true", res.Headers["Access-Control-Allow-Credentials"])
	require.Contains(t, res.Headers["Access-Control-Allow-Headers"], "CF-Turnstile-Response")
	require.Contains(t, res.Headers["Access-Control-Allow-Methods"], "GET")
}
