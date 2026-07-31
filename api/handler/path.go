package handler

import (
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

// normalizeAPIPath returns the route name (e.g. search, session) after optional stage
// and legacy /api prefixes.
func normalizeAPIPath(request events.APIGatewayProxyRequest) string {
	path := strings.TrimSpace(request.Path)
	if path == "" {
		path = strings.TrimSpace(request.RequestContext.Path)
	}
	if path == "" {
		path = "/"
	}

	for _, prefix := range []string{"/prod", "/staging", "/dev", "/default"} {
		if strings.HasPrefix(path, prefix+"/") || path == prefix {
			path = strings.TrimPrefix(path, prefix)
			if path == "" {
				path = "/"
			}
		}
	}

	path = strings.TrimPrefix(path, "/api")
	path = strings.TrimPrefix(path, "/")
	return path
}
