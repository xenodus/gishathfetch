package gateway

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode"
)

const defaultResponseBodyPreviewLimit = 200

// ReadResponseBody reads an HTTP response body, transparently decompressing gzip
// when needed. Go's net/http does not auto-decompress when the request sets
// Accept-Encoding manually (as browser-like headers do), so callers that set
// Accept-Encoding must use this helper before parsing JSON or HTML.
func ReadResponseBody(resp *http.Response) ([]byte, error) {
	if resp == nil || resp.Body == nil {
		return nil, fmt.Errorf("gateway: nil response body")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.Uncompressed {
		return body, nil
	}
	return decompressGzipIfNeeded(body, resp.Header.Get("Content-Encoding"))
}

func decompressGzipIfNeeded(body []byte, contentEncoding string) ([]byte, error) {
	if len(body) == 0 {
		return body, nil
	}
	if !isGzipBody(body, contentEncoding) {
		return body, nil
	}

	zr, err := gzip.NewReader(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	return io.ReadAll(zr)
}

func isGzipBody(body []byte, contentEncoding string) bool {
	if strings.Contains(strings.ToLower(contentEncoding), "gzip") {
		return true
	}
	return len(body) >= 2 && body[0] == 0x1f && body[1] == 0x8b
}

// ResponseBodyPreview returns a single-line, truncated preview of body for error messages.
func ResponseBodyPreview(body []byte, maxLen int) string {
	if len(body) == 0 {
		return "(empty)"
	}
	if maxLen <= 0 {
		maxLen = defaultResponseBodyPreviewLimit
	}

	var b strings.Builder
	for _, r := range string(body) {
		switch {
		case r == '\n' || r == '\r' || r == '\t':
			b.WriteByte(' ')
		case unicode.IsPrint(r):
			b.WriteRune(r)
		default:
			b.WriteByte('?')
		}
		if b.Len() >= maxLen {
			break
		}
	}

	preview := strings.TrimSpace(b.String())
	if len(body) > len(preview) {
		return preview + "..."
	}
	return preview
}

// FormatHTTPRequestContext builds request metadata for transport-layer error messages.
func FormatHTTPRequestContext(req *http.Request, extra ...string) string {
	parts := make([]string, 0, 4+len(extra))
	if req != nil {
		parts = append(parts, "method="+req.Method)
		if req.URL != nil {
			parts = append(parts, "url="+req.URL.String())
		}
		if req.ContentLength >= 0 {
			parts = append(parts, fmt.Sprintf("request_body_len=%d", req.ContentLength))
		}
		if deadline, ok := req.Context().Deadline(); ok {
			parts = append(parts, "context_deadline="+deadline.UTC().Format(time.RFC3339))
		}
		if ctxErr := req.Context().Err(); ctxErr != nil {
			parts = append(parts, "context_err="+ctxErr.Error())
		}
	}
	parts = append(parts, extra...)
	return strings.Join(parts, " ")
}

// WrapHTTPRequestError annotates an HTTP client/transport error with request context.
func WrapHTTPRequestError(err error, req *http.Request, extra ...string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("http request failed (%s): %w", FormatHTTPRequestContext(req, extra...), err)
}

// WrapResponseBodyReadError annotates a response body read error with HTTP context.
func WrapResponseBodyReadError(err error, resp *http.Response) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("read response body failed (%s): %w", formatHTTPResponseContext(resp, nil), err)
}

func formatHTTPResponseContext(resp *http.Response, body []byte) string {
	parts := make([]string, 0, 4)
	if resp != nil {
		parts = append(parts, "status="+resp.Status)
		if contentType := resp.Header.Get("Content-Type"); contentType != "" {
			parts = append(parts, "content-type="+contentType)
		}
	}
	parts = append(parts, fmt.Sprintf("body_len=%d", len(body)))
	parts = append(parts, fmt.Sprintf("body_preview=%q", ResponseBodyPreview(body, defaultResponseBodyPreviewLimit)))
	return strings.Join(parts, " ")
}

// FormatUnexpectedHTTPStatus builds an error message for a non-success HTTP response.
func FormatUnexpectedHTTPStatus(storeName string, resp *http.Response, body []byte) string {
	status := ""
	if resp != nil {
		status = resp.Status
	}
	return fmt.Sprintf("unexpected status for %s: %s (%s)", storeName, status, formatHTTPResponseContext(resp, body))
}

// WrapJSONDecodeError annotates a JSON decode error with HTTP response context.
func WrapJSONDecodeError(err error, resp *http.Response, body []byte) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("json decode failed (%s): %w", formatHTTPResponseContext(resp, body), err)
}
