package google

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_LGSNoResultMeasurement(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/collect", r.URL.Path)
		require.Equal(t, measurementID, r.URL.Query().Get("measurement_id"))
		_, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	measurementAPIBaseUrl = server.URL + "/collect"
	t.Setenv(apiSecretKey, "test-secret")

	err := LGSNoResultMeasurement("test_lgs", "test_keyword")
	require.NoError(t, err)
}
