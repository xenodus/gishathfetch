package gatewaytest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	"mtg-price-checker-sg/gateway"

	"github.com/stretchr/testify/require"
)

// JSONProbe fetches an API endpoint and validates the response body shape.
type JSONProbe struct {
	Method  string
	URL     string
	Body    []byte
	Headers map[string]string
	// Validate inspects a successful response body.
	Validate func(body []byte) error
}

// RequireJSONStructure verifies the API endpoint is reachable and the response
// still unmarshals into the expected shape.
func RequireJSONStructure(t *testing.T, ctx context.Context, probe JSONProbe) {
	t.Helper()
	require.NotNil(t, probe.Validate)
	method := probe.Method
	if method == "" {
		method = http.MethodGet
	}

	var bodyReader io.Reader
	if len(probe.Body) > 0 {
		bodyReader = bytes.NewReader(probe.Body)
	}

	req, err := http.NewRequestWithContext(ctx, method, probe.URL, bodyReader)
	require.NoError(t, err)
	for key, value := range probe.Headers {
		req.Header.Set(key, value)
	}
	if err := gateway.PrepareOutboundRequest(ctx, req, gateway.OutboundRequestOptions{}); err != nil {
		require.NoError(t, err)
	}

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.True(t, resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices,
		"expected success status from %s, got %s", probe.URL, resp.Status)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, probe.Validate(body), "response from %s has unexpected structure", probe.URL)
}

// RequireJSONObjectKeys ensures the top-level JSON value is an object with keys.
func RequireJSONObjectKeys(t *testing.T, body []byte, keys ...string) {
	t.Helper()
	var payload map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(body, &payload))
	for _, key := range keys {
		_, ok := payload[key]
		require.True(t, ok, "expected JSON object key %q", key)
	}
}

// RequireJSONArray ensures the top-level JSON value is an array (empty is fine).
func RequireJSONArray(t *testing.T, body []byte) {
	t.Helper()
	var payload []json.RawMessage
	require.NoError(t, json.Unmarshal(body, &payload))
}

// BuildURL is a small helper for tests that assemble probe URLs.
func BuildURL(scheme, host, path string, query url.Values) string {
	return (&url.URL{
		Scheme:   scheme,
		Host:     host,
		Path:     path,
		RawQuery: query.Encode(),
	}).String()
}

// ValidateErrorf wraps validation failures with context.
func ValidateErrorf(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}
