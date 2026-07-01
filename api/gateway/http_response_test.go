package gateway

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestReadResponseBody_plainJSON(t *testing.T) {
	body := []byte(`{"ok":true}`)
	resp := &http.Response{
		Body:          io.NopCloser(bytes.NewReader(body)),
		Uncompressed:  true,
		Header:        make(http.Header),
	}
	got, err := ReadResponseBody(resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != string(body) {
		t.Fatalf("got %q, want %q", got, body)
	}
}

func TestReadResponseBody_gzipJSON(t *testing.T) {
	plain := []byte(`{"resources":{"results":{"products":[]}}}`)
	var compressed bytes.Buffer
	zw := gzip.NewWriter(&compressed)
	if _, err := zw.Write(plain); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}

	resp := &http.Response{
		Body: io.NopCloser(bytes.NewReader(compressed.Bytes())),
		Header: http.Header{
			"Content-Encoding": []string{"gzip"},
		},
	}
	got, err := ReadResponseBody(resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != string(plain) {
		t.Fatalf("got %q, want %q", got, plain)
	}
}

func TestReadResponseBody_gzipMagicWithoutHeader(t *testing.T) {
	plain := []byte(`{"ok":true}`)
	var compressed bytes.Buffer
	zw := gzip.NewWriter(&compressed)
	if _, err := zw.Write(plain); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}

	resp := &http.Response{
		Body:   io.NopCloser(bytes.NewReader(compressed.Bytes())),
		Header: make(http.Header),
	}
	got, err := ReadResponseBody(resp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != string(plain) {
		t.Fatalf("got %q, want %q", got, plain)
	}
}

func TestResponseBodyPreview(t *testing.T) {
	if got := ResponseBodyPreview(nil, 0); got != "(empty)" {
		t.Fatalf("empty body = %q", got)
	}

	html := []byte("<!DOCTYPE html>\n<html><title>Blocked</title></html>")
	got := ResponseBodyPreview(html, 40)
	if !strings.Contains(got, "<!DOCTYPE html>") {
		t.Fatalf("preview = %q", got)
	}
	if !strings.HasSuffix(got, "...") {
		t.Fatalf("expected truncated preview, got %q", got)
	}
}

func TestWrapJSONDecodeError(t *testing.T) {
	resp := &http.Response{
		Status:     "403 Forbidden",
		StatusCode: http.StatusForbidden,
		Header: http.Header{
			"Content-Type": []string{"text/html"},
		},
	}
	body := []byte("<html>blocked</html>")
	err := WrapJSONDecodeError(json.Unmarshal(body, &map[string]any{}), resp, body)
	if err == nil {
		t.Fatal("expected error")
	}
	msg := err.Error()
	for _, want := range []string{
		"json decode failed",
		"status=403 Forbidden",
		"content-type=text/html",
		"body_len=20",
		"body_preview=",
		"invalid character",
	} {
		if !strings.Contains(msg, want) {
			t.Fatalf("error %q missing %q", msg, want)
		}
	}
}

func TestFormatUnexpectedHTTPStatus(t *testing.T) {
	resp := &http.Response{
		Status:     "503 Service Unavailable",
		StatusCode: http.StatusServiceUnavailable,
	}
	msg := FormatUnexpectedHTTPStatus("Cards & Collections", resp, []byte("down"))
	if !strings.Contains(msg, "unexpected status for Cards & Collections: 503 Service Unavailable") {
		t.Fatalf("unexpected message: %s", msg)
	}
	if !strings.Contains(msg, "body_preview=\"down\"") {
		t.Fatalf("missing body preview: %s", msg)
	}
}

func TestWrapJSONDecodeError_nil(t *testing.T) {
	if err := WrapJSONDecodeError(nil, nil, nil); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestWrapJSONDecodeError_preservesCause(t *testing.T) {
	cause := errors.New("decode failed")
	err := WrapJSONDecodeError(cause, nil, []byte("{bad"))
	if !errors.Is(err, cause) {
		t.Fatalf("expected wrapped cause, got %v", err)
	}
}
