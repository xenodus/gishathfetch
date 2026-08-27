package handler

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/require"
)

func TestParseAPIRequest_HTTPAPIV2Webhook(t *testing.T) {
	t.Parallel()

	event, err := json.Marshal(events.APIGatewayV2HTTPRequest{
		Version: "2.0",
		RouteKey: "POST /telegram/webhook",
		RawPath: "/telegram/webhook",
		Headers: map[string]string{
			"x-telegram-bot-api-secret-token": "secret",
		},
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
				Method: http.MethodPost,
				Path:   "/telegram/webhook",
			},
		},
		Body: `{"message":{"chat":{"id":1},"text":"/help"}}`,
	})
	require.NoError(t, err)

	req, err := parseAPIRequest(event)
	require.NoError(t, err)
	require.Equal(t, http.MethodPost, req.HTTPMethod)
	require.Equal(t, "/telegram/webhook", req.Path)
	require.Equal(t, "telegram/webhook", normalizeAPIPath(req))
}

func TestParseAPIRequest_RESTProxySearch(t *testing.T) {
	t.Parallel()

	event, err := json.Marshal(events.APIGatewayProxyRequest{
		HTTPMethod: http.MethodGet,
		Path:       "/default/search",
		QueryStringParameters: map[string]string{
			"s": "Opt",
		},
	})
	require.NoError(t, err)

	req, err := parseAPIRequest(event)
	require.NoError(t, err)
	require.Equal(t, http.MethodGet, req.HTTPMethod)
	require.Equal(t, "search", normalizeAPIPath(req))
}
