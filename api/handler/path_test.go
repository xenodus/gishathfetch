package handler

import (
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/require"
)

func Test_normalizeAPIPath(t *testing.T) {
	t.Parallel()

	cases := []struct {
		path string
		want string
	}{
		{"/search", "search"},
		{"/session", "session"},
		{"/api/search", "search"},
		{"/api/session", "session"},
		{"/prod/search", "search"},
		{"/default/search", "search"},
		{"/default/session", "session"},
		{"/telegram/search", "telegram/search"},
		{"/default/telegram/search", "telegram/search"},
		{"/telegram/webhook", "telegram/webhook"},
		{"/default/telegram/webhook", "telegram/webhook"},
		{"/prod/api/search", "search"},
		{"/prod/api", ""},
		{"/", ""},
	}

	for _, tc := range cases {
		req := events.APIGatewayProxyRequest{Path: tc.path}
		require.Equal(t, tc.want, normalizeAPIPath(req), "path %q", tc.path)
	}
}
