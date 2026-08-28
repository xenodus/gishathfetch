package ga4

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"mtg-price-checker-sg/pkg/config"

	"github.com/stretchr/testify/require"
)

func TestMeasurementSender_SendSearchEvent(t *testing.T) {
	var gotMethod string
	var gotQuery map[string]string
	var gotPayload map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotQuery = map[string]string{
			"measurement_id": r.URL.Query().Get("measurement_id"),
			"api_secret":     r.URL.Query().Get("api_secret"),
		}

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(body, &gotPayload))
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	sender := &MeasurementSender{
		collectURL:    server.URL,
		measurementID: "G-TEST123",
		apiSecret:     "secret-value",
		httpClient:    server.Client(),
	}

	require.NoError(t, sender.SendSearchEvent(context.Background(), "Lightning Bolt"))

	require.Equal(t, http.MethodPost, gotMethod)
	require.Equal(t, "G-TEST123", gotQuery["measurement_id"])
	require.Equal(t, "secret-value", gotQuery["api_secret"])
	require.Equal(t, telegramClientID, gotPayload["client_id"])

	events, ok := gotPayload["events"].([]any)
	require.True(t, ok)
	require.Len(t, events, 1)

	event, ok := events[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, SearchEventName, event["name"])

	params, ok := event["params"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "Lightning Bolt", params["search_term"])
}

func TestMeasurementSender_SendSearchEvent_RejectsEmptyTerm(t *testing.T) {
	sender := &MeasurementSender{
		collectURL:    "http://example.com/mp/collect",
		measurementID: "G-TEST123",
		apiSecret:     "secret-value",
		httpClient:    http.DefaultClient,
	}

	err := sender.SendSearchEvent(context.Background(), "   ")
	require.Error(t, err)
}

func TestNewMeasurementSender_RequiresAPISecret(t *testing.T) {
	t.Cleanup(func() {
		_ = os.Unsetenv(config.GA4MeasurementAPISecretEnv)
		_ = os.Unsetenv(config.GA4MeasurementIDEnv)
	})

	require.NoError(t, os.Setenv(config.GA4MeasurementIDEnv, "G-CUSTOM"))
	_, err := NewMeasurementSender()
	require.Error(t, err)
}

func TestNewMeasurementSender_DefaultsMeasurementID(t *testing.T) {
	t.Cleanup(func() {
		_ = os.Unsetenv(config.GA4MeasurementAPISecretEnv)
		_ = os.Unsetenv(config.GA4MeasurementIDEnv)
	})

	require.NoError(t, os.Setenv(config.GA4MeasurementAPISecretEnv, "secret-value"))
	sender, err := NewMeasurementSender()
	require.NoError(t, err)
	require.Equal(t, DefaultMeasurementID, sender.measurementID)
}

func TestTrySendSearchEvent_SkipsWhenNotConfigured(t *testing.T) {
	t.Cleanup(func() {
		_ = os.Unsetenv(config.GA4MeasurementAPISecretEnv)
	})

	TrySendSearchEvent(context.Background(), "Opt")
}
