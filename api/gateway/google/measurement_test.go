package google

import (
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

func init() {
	_ = godotenv.Load("../../.env")
}

func Test_LGSNoResultMeasurement(t *testing.T) {
	measurementAPIBaseUrl = "https://www.google-analytics.com/debug/mp/collect"
	err := LGSNoResultMeasurement("test_lgs", "test_keyword")
	require.NoError(t, err)
}
