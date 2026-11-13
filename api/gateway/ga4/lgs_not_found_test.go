package ga4

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"mtg-price-checker-sg/controller"

	"github.com/stretchr/testify/require"
)

func TestStoresWithNoResults(t *testing.T) {
	t.Parallel()

	require.Nil(t, StoresWithNoResults(nil, nil, nil))

	got := StoresWithNoResults(
		[]string{"Alpha", "Beta"},
		[]controller.StoreStat{
			{Store: "Alpha", ItemCount: 0},
			{Store: "Beta", ItemCount: 2},
		},
		nil,
	)
	require.Equal(t, []string{"Alpha"}, got)

	got = StoresWithNoResults(
		[]string{"Alpha", "Beta"},
		[]controller.StoreStat{
			{Store: "Alpha", ItemCount: 0},
			{Store: "Beta", ItemCount: 0},
		},
		[]controller.StoreError{{Store: "Beta", Error: "timeout"}},
	)
	require.Equal(t, []string{"Alpha"}, got)
}

func TestMeasurementSender_SendLgsCardNotFoundEvent(t *testing.T) {
	var gotPayload map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	require.NoError(t, sender.sendMeasurementEvent(
		t.Context(),
		"client-abc",
		LgsCardNotFoundEventName,
		map[string]any{
			"search_term": "Lightning Bolt",
			lgsParam:      "Hideout",
		},
	))

	events, ok := gotPayload["events"].([]any)
	require.True(t, ok)
	require.Len(t, events, 1)

	event, ok := events[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, LgsCardNotFoundEventName, event["name"])

	params, ok := event["params"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "Lightning Bolt", params["search_term"])
	require.Equal(t, "Hideout", params[lgsParam])
}
