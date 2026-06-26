package gateway

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
)

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
